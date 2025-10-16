use diesel::prelude::*;
use std::collections::HashMap;
use crate::db::*;
use crate::utils::*;

pub fn needs_number(n: i32) -> HashMap<String, i32> {
    let mut needs = HashMap::new();
    needs.insert("main".to_string(), ((n + 1) as f32 / 4.0).ceil() as i32);
    needs.insert("side".to_string(), (n as f32 / 6.0).ceil() as i32);
    needs.insert(
        "plastics".to_string(),
        if n > 5 {
            (n as f32 / 20.0).ceil() as i32
        } else {
            0
        },
    );
    needs.insert(
        "drinks".to_string(),
        if n > 5 {
            (n as f32 / 12.0).ceil() as i32
        } else {
            0
        },
    );
    needs.insert("wine".to_string(), (n as f32 / 8.0).ceil() as i32);
    needs.insert("challah".to_string(), (n as f32 / 10.0).ceil() as i32);
    needs.insert("dips".to_string(), ((n + 1) as f32 / 12.0).ceil() as i32);
    needs.insert("dessert".to_string(), ((n - 5) as f32 / 8.0).ceil() as i32);
    needs
}

pub fn format_quickshab(db: &mut SqliteConnection, chat_id: &str) -> String {
    use crate::db::schema::quickshab::dsl::{chat_id as qs_chat_id, number, quickshab};
    use crate::db::schema::quickshab_assignments::dsl::{
        category as qa_category, chat_id as qa_chat_id, name as qa_name, quickshab_assignments,
    };

    let num_people: i32 = match quickshab
        .filter(qs_chat_id.eq(chat_id))
        .select(number)
        .first::<i32>(db)
    {
        Ok(n) => n,
        Err(_) => return "No quickshab found".to_string(),
    };

    let needs = needs_number(num_people);

    let type_list = vec![
        "main", "side", "plastics", "drinks", "wine", "challah", "dips", "dessert",
    ];

    let assignment_rows: Vec<(String, String)> = match quickshab_assignments
        .filter(qa_chat_id.eq(chat_id))
        .select((qa_category, qa_name))
        .order((qa_category, crate::db::schema::quickshab_assignments::id))
        .load::<(String, String)>(db)
    {
        Ok(rows) => rows,
        Err(_) => return "Error retrieving assignments".to_string(),
    };

    let mut assignments: HashMap<String, Vec<String>> = HashMap::new();
    for (category, name) in assignment_rows {
        assignments
            .entry(category)
            .or_insert_with(Vec::new)
            .push(name);
    }

    let mut result = Vec::new();

    for category in &type_list {
        let needed = *needs.get(*category).unwrap_or(&0);
        let assigned = assignments.get(*category).map(|v| v.len()).unwrap_or(0);
        let upper_lim = std::cmp::max(needed, assigned as i32);

        for i in 0..upper_lim {
            let name = assignments
                .get(*category)
                .and_then(|v| v.get(i as usize))
                .map(|s| s.as_str())
                .unwrap_or("");
            result.push(format!("{}: {}", capitalize_first(category), name));
        }
    }

    for (category, names) in &assignments {
        if !type_list.contains(&category.as_str()) {
            for name in names {
                result.push(format!("{}: {}", capitalize_first(category), name));
            }
        }
    }

    result.join("\n")
}

pub fn quickshab_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        return "Command syntax: !quickshab [number of people]".to_string();
    }

    let num_people = match args[1].parse::<i32>() {
        Ok(n) if n > 0 => n,
        _ => return "Command syntax: !quickshab [number of people]".to_string(),
    };

    use crate::db::schema::quickshab::dsl::{chat_id as qs_chat_id, number, quickshab};
    use crate::db::schema::quickshab_assignments::dsl::{chat_id as qa_chat_id, quickshab_assignments};
    use diesel::{delete, insert_into};

    let _ = get_or_create_chat(db, chat_id);

    let _ = delete(quickshab_assignments.filter(qa_chat_id.eq(chat_id))).execute(db);

    if let Err(_) = insert_into(quickshab)
        .values((qs_chat_id.eq(chat_id), number.eq(num_people)))
        .on_conflict(qs_chat_id)
        .do_update()
        .set(number.eq(num_people))
        .execute(db)
    {
        return "Error creating quickshab".to_string();
    }

    format!(
        "{}\n\nYou can sign yourself up by writing !bring [category]",
        format_quickshab(db, chat_id)
    )
}

pub fn update_number_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        return "Command syntax: !update [number of people]".to_string();
    }

    let num_people = match args[1].parse::<i32>() {
        Ok(n) if n > 0 => n,
        _ => return "Command syntax: !update [number of people]".to_string(),
    };

    use crate::db::schema::quickshab::dsl::{chat_id as qs_chat_id, number, quickshab};
    use diesel::dsl::count_star;

    let exists: i64 = quickshab
        .filter(qs_chat_id.eq(chat_id))
        .select(count_star())
        .first(db)
        .unwrap_or(0);

    if exists == 0 {
        return "You need to create a quickshab first by typing !quickshab [number of people]"
            .to_string();
    }

    if let Err(_) = diesel::update(quickshab.filter(qs_chat_id.eq(chat_id)))
        .set(number.eq(num_people))
        .execute(db)
    {
        return "Error updating quickshab".to_string();
    }

    format_quickshab(db, chat_id)
}

pub fn show_cmd(db: &mut SqliteConnection, chat_id: &str) -> String {
    use crate::db::schema::quickshab::dsl::{chat_id as qs_chat_id, quickshab};
    use diesel::dsl::count_star;

    let exists: i64 = quickshab
        .filter(qs_chat_id.eq(chat_id))
        .select(count_star())
        .first(db)
        .unwrap_or(0);

    if exists == 0 {
        return "No quickshab found. Create one with !quickshab [number]".to_string();
    }

    format!(
        "{}\n\nExample Usage:\n   • !bring main\n   • !assign gabe main\n\n\
         You can also use !unbring and !unassign",
        format_quickshab(db, chat_id)
    )
}
