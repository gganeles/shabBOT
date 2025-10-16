use diesel::prelude::*;
use crate::utils::*;
use super::quickshab::format_quickshab;

pub fn bring_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str, attendee_id: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        return "Command syntax: !bring [category]".to_string();
    }

    use crate::db::schema::quickshab::dsl::{chat_id as qs_chat_id, quickshab};
    use crate::db::schema::quickshab_assignments::dsl::{
        category as qa_category, chat_id as qa_chat_id, name as qa_name, quickshab_assignments,
    };
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

    let category = args[1].to_lowercase();

    if let Err(_) = diesel::insert_into(quickshab_assignments)
        .values((
            qa_chat_id.eq(chat_id),
            qa_category.eq(&category),
            qa_name.eq(attendee_id),
        ))
        .execute(db)
    {
        return "Error adding assignment".to_string();
    }

    format_quickshab(db, chat_id)
}

pub fn assign_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 3 {
        return "Command syntax: !assign [name] [category]".to_string();
    }

    use crate::db::schema::quickshab::dsl::{chat_id as qs_chat_id, quickshab};
    use crate::db::schema::quickshab_assignments::dsl::{
        category as qa_category, chat_id as qa_chat_id, name as qa_name, quickshab_assignments,
    };
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

    let name = &args[1];
    let category = args[2].to_lowercase();

    if let Err(_) = diesel::insert_into(quickshab_assignments)
        .values((
            qa_chat_id.eq(chat_id),
            qa_category.eq(&category),
            qa_name.eq(name),
        ))
        .execute(db)
    {
        return "Error adding assignment".to_string();
    }

    format_quickshab(db, chat_id)
}

pub fn unbring_cmd(db: &mut SqliteConnection, chat_id: &str, attendee_id: &str) -> String {
    use crate::db::schema::quickshab::dsl::{chat_id as qs_chat_id, quickshab};
    use crate::db::schema::quickshab_assignments::dsl::{
        chat_id as qa_chat_id, id as qa_id, name as qa_name, quickshab_assignments,
    };
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

    let assignment_id: Option<i32> = quickshab_assignments
        .filter(qa_chat_id.eq(chat_id).and(qa_name.eq(attendee_id)))
        .select(qa_id)
        .first(db)
        .ok();

    match assignment_id {
        Some(id) => {
            let result = diesel::delete(quickshab_assignments.filter(qa_id.eq(id))).execute(db);
            match result {
                Ok(_) => format_quickshab(db, chat_id),
                Err(_) => "Error removing assignment".to_string(),
            }
        }
        None => "You don't have any assignments".to_string(),
    }
}

pub fn unassign_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        return "Command syntax: !unassign [name]".to_string();
    }

    let name = args[1..].join(" ");

    use crate::db::schema::quickshab::dsl::{chat_id as qs_chat_id, quickshab};
    use crate::db::schema::quickshab_assignments::dsl::{
        chat_id as qa_chat_id, id as qa_id, name as qa_name, quickshab_assignments,
    };
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

    let assignment_id: Option<i32> = quickshab_assignments
        .filter(qa_chat_id.eq(chat_id).and(qa_name.eq(&name)))
        .select(qa_id)
        .first(db)
        .ok();

    match assignment_id {
        Some(id) => {
            let result = diesel::delete(quickshab_assignments.filter(qa_id.eq(id))).execute(db);
            match result {
                Ok(_) => format_quickshab(db, chat_id),
                Err(_) => "Error removing assignment".to_string(),
            }
        }
        None => format!("{} doesn't have any assignments", name),
    }
}
