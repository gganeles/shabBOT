use chrono_tz::Tz;
use diesel::prelude::*;
use diesel::sqlite::SqliteConnection;
use regex::Regex;
use std::cmp::max;
use std::collections::HashMap;
use uuid::Uuid;

use crate::commands::*;
use crate::db::*;
use crate::deno_client::{DenoRunner, ParsedDate};
use crate::utils::*;

// Global lazy static for DenoRunner
use std::sync::Mutex;
use std::sync::OnceLock;

static DENO_RUNNER: OnceLock<Mutex<DenoRunner>> = OnceLock::new();

/// Get or initialize the global Deno runner
fn get_deno_runner() -> &'static Mutex<DenoRunner> {
    DENO_RUNNER.get_or_init(|| match DenoRunner::new() {
        Ok(runner) => Mutex::new(runner),
        Err(e) => {
            eprintln!("Failed to start Deno runner: {}", e);
            panic!("Cannot proceed without Deno runner");
        }
    })
}

/// Parse natural language date/time using chrono-node via Deno
pub fn parse_natural_time(
    expression: &str,
    reference_date: Option<&str>,
) -> Result<Vec<ParsedDate>, String> {
    let runner = get_deno_runner();
    let mut runner = runner.lock().unwrap();

    runner
        .parse_date(expression, reference_date)
        .map_err(|e| format!("Date parsing error: {}", e))
}

/// A command router that processes WhatsApp messages and routes them to appropriate handlers.
pub struct CommandRouter {
    pub db: SqliteConnection,
}

impl CommandRouter {
    pub fn new(db_path: &str) -> Result<Self, Box<dyn std::error::Error>> {
        let mut db = SqliteConnection::establish(db_path)?;
        init_db(&mut db)?;
        Ok(Self { db })
    }

    /// Route processes a message and returns the appropriate response.
    pub fn route(&mut self, prompt: &str, chat_id: &str, attendee_id: &str) -> String {
        if !prompt.starts_with('!') {
            return String::new();
        }

        let prompt_trimmed = prompt.trim_start_matches('!').trim();
        let args = parse_args(prompt_trimmed);
        if args.is_empty() {
            return String::new();
        }

        let cmd = args[0].to_lowercase();

        // Get location and timezone
        let _location =
            get_chat_location(&mut self.db, chat_id).unwrap_or_else(|_| "Haifa".to_string());
        let timezone_name = get_chat_timezone(&mut self.db, chat_id);

        // Parse timezone
        let _tz: Tz = match timezone_name.parse() {
            Ok(tz) => tz,
            Err(e) => return format!("Error loading timezone: {}", e),
        };

        match cmd.as_str() {
            "help" => help::help_cmd(prompt_trimmed),
            "docs" => help::docs_cmd(),
            "count" => misc::count_cmd(),

            "shablocation" | "shabloc" | "location" | "setloc" | "chatlocation" => {
                location::shablocation_cmd(&mut self.db, prompt_trimmed, chat_id)
            }
            "shabtimes" | "shabbattimes" => {
                location::shabtimes_cmd(&mut self.db, prompt_trimmed, chat_id)
            }
            "chag" | "chagtimes" | "holiday" | "holidays" | "nextholiday" => {
                location::chagtimes_cmd(&mut self.db, prompt_trimmed, chat_id)
            }

            "quickshab" => quickshab::quickshab_cmd(&mut self.db, prompt_trimmed, chat_id),
            "update" | "up" | "num" | "ppl" => {
                quickshab::update_number_cmd(&mut self.db, prompt_trimmed, chat_id)
            }
            "show" => quickshab::show_cmd(&mut self.db, chat_id),

            "bring" | "br" | "bringing" => {
                assignments::bring_cmd(&mut self.db, prompt_trimmed, chat_id, attendee_id)
            }
            "assign" => assignments::assign_cmd(&mut self.db, prompt_trimmed, chat_id),
            "unbring" | "unbr" => assignments::unbring_cmd(&mut self.db, chat_id, attendee_id),
            "unassign" => assignments::unassign_cmd(&mut self.db, prompt_trimmed, chat_id),

            "shop" | "shp" | "sh" => shopping::shop_cmd(&mut self.db, prompt_trimmed, chat_id),
            "unshop" | "unshp" | "unsh" => {
                shopping::unshop_cmd(&mut self.db, prompt_trimmed, chat_id)
            }
            "shoplist" | "shplist" | "shoppinglist" | "shlst" | "shlist" | "shls" => {
                shopping::shoplist_cmd(&mut self.db, chat_id)
            }

            "remind" | "r" => {
                reminders::remind_cmd(&mut self.db, prompt_trimmed, chat_id, &timezone_name)
            }
            "reminders" | "rems" | "todo" | "rls" => {
                reminders::reminders_cmd(&mut self.db, chat_id, &timezone_name)
            }
            "snooze" => {
                reminders::snooze_cmd(&mut self.db, prompt_trimmed, chat_id, &timezone_name)
            }
            "done" => reminders::done_cmd(&mut self.db, prompt_trimmed, chat_id),

            "send" => scheduled::send_cmd(&mut self.db, prompt_trimmed, chat_id, &timezone_name),
            "unsend" => scheduled::unsend_cmd(&mut self.db, chat_id),
            "scheduled" => scheduled::scheduled_cmd(&mut self.db, chat_id),

            "needs" => misc::needs_cmd(prompt_trimmed),
            "fast" | "fasting" => misc::fast_cmd(&mut self.db, chat_id),
            "start" => misc::start_cmd(prompt_trimmed),
            "end" => misc::end_cmd(),

            _ => String::new(),
        }
    }

    /// Parse multiple commands separated by newlines or ' !'
    pub fn parse_multiple_commands(
        &mut self,
        message: &str,
        chat_id: &str,
        attendee_id: &str,
    ) -> Vec<String> {
        let mut responses = Vec::new();
        for line in message.lines() {
            let line = line.trim();
            if line.is_empty() {
                continue;
            }

            let parts: Vec<&str> = line.split(" !").collect();
            for (i, part) in parts.iter().enumerate() {
                let p = if i > 0 {
                    format!("!{}", part)
                } else {
                    part.to_string()
                };
                let resp = self.route(&p, chat_id, attendee_id);
                if !resp.is_empty() {
                    responses.push(resp);
                }
            }
        }
        responses
    }
}

// ============================================================================
// SIMPLE COMMANDS
// ============================================================================

fn help_cmd(_prompt: &str) -> String {
    "Allow me to introduce myself!\n\n\
     i am the SHABbot!!\n\
     I can help you with you shabbat meals, as well as other things\n\n\
     The way it works:\n\n\
      - Use !quickshab to keep track of what everyone's bringing\n\
      - Use !shabtimes to find out when shabbat starts\n\
      - Use !remind to set reminders for yourself\n\n\
     For more information, type !docs to see the full documentation"
        .to_string()
}

fn docs_cmd() -> String {
    "github.com/gganeles/shabBOT".to_string()
}

fn count_cmd() -> String {
    let link = "https://www.chabad.org/holidays/sefirah/omer-count_cdo/jewish/Count-the-Omer.htm";
    format!(
        "you forgot that your suppose to count the omer!? ha! no, not tonight\n{}",
        link
    )
}

fn needs_cmd(_prompt: &str) -> String {
    "The needs command requires the events system which is being deprecated.\n\
     Please use !quickshab instead to track meal items."
        .to_string()
}

fn fast_cmd(db: &mut SqliteConnection, chat_id: &str) -> String {
    let location = get_chat_location(db, chat_id).unwrap_or_else(|_| "Haifa".to_string());

    format!(
        "Fast day checking for {}\n\n\
         To implement this properly, you need to:\n\
         1. Integrate with a Hebrew calendar library (e.g., hebcal.com API)\n\
         2. Get current Hebrew date\n\
         3. Check if current/upcoming date is a fast day\n\
         4. Calculate fast start/end times based on location\n\n\
         Fast days include:\n\
         - Fast of Gedaliah (3 Tishrei)\n\
         - Yom Kippur (10 Tishrei)\n\
         - Fast of Tevet (10 Tevet)\n\
         - Fast of Esther (13 Adar)\n\
         - Fast of Tammuz (17 Tammuz)\n\
         - Tisha B'Av (9 Av)",
        location
    )
}

fn start_cmd(prompt: &str) -> String {
    format!("Start: {}", prompt)
}

fn end_cmd() -> String {
    "Goodbye".to_string()
}

// ============================================================================
// LOCATION COMMANDS
// ============================================================================

fn shablocation_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        let location = get_chat_location(db, chat_id).unwrap_or_else(|_| "Haifa".to_string());
        return format!("Current location: {}", location);
    }

    let location = args[1..].join(" ");
    let location = location.trim();

    if location.is_empty() {
        return "Please provide a valid location".to_string();
    }

    let geo = find_geo_location(location);
    if let Some(geo_data) = geo {
        if let Err(e) = set_chat_location(db, chat_id, &geo_data.name, &geo_data.timezone) {
            return format!("Error setting location: {}", e);
        }
        format!("Chat location was set to {}", geo_data.name)
    } else {
        "Location not found. Using default location.".to_string()
    }
}

fn shabtimes_cmd(db: &mut SqliteConnection, _prompt: &str, chat_id: &str) -> String {
    let location = get_chat_location(db, chat_id).unwrap_or_else(|_| "Haifa".to_string());
    format!(
        "Shabbat times for {} - Implementation requires zmanim library integration",
        location
    )
}

fn chagtimes_cmd(db: &mut SqliteConnection, _prompt: &str, chat_id: &str) -> String {
    let location = get_chat_location(db, chat_id).unwrap_or_else(|_| "Haifa".to_string());
    format!(
        "Holiday times for {} - Implementation requires Hebrew calendar API",
        location
    )
}

// ============================================================================
// QUICKSHAB COMMANDS
// ============================================================================

fn needs_number(n: i32) -> HashMap<String, i32> {
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

fn format_quickshab(db: &mut SqliteConnection, chat_id: &str) -> String {
    use crate::db::schema::quickshab::dsl::{chat_id as qs_chat_id, number, quickshab};
    use crate::db::schema::quickshab_assignments::dsl::{
        category as qa_category, chat_id as qa_chat_id, name as qa_name, quickshab_assignments,
    };

    // Get number of people
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

    // Get assignments
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

    // Process standard categories
    for category in &type_list {
        let needed = *needs.get(*category).unwrap_or(&0);
        let assigned = assignments.get(*category).map(|v| v.len()).unwrap_or(0);
        let upper_lim = max(needed, assigned as i32);

        for i in 0..upper_lim {
            let name = assignments
                .get(*category)
                .and_then(|v| v.get(i as usize))
                .map(|s| s.as_str())
                .unwrap_or("");
            result.push(format!("{}: {}", capitalize_first(category), name));
        }
    }

    // Process custom categories
    for (category, names) in &assignments {
        if !type_list.contains(&category.as_str()) {
            for name in names {
                result.push(format!("{}: {}", capitalize_first(category), name));
            }
        }
    }

    result.join("\n")
}

fn quickshab_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
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

    // Delete existing assignments
    let _ = delete(quickshab_assignments.filter(qa_chat_id.eq(chat_id))).execute(db);

    // Insert or update quickshab
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

fn update_number_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
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

    // Check if quickshab exists
    let exists: i64 = quickshab
        .filter(qs_chat_id.eq(chat_id))
        .select(count_star())
        .first(db)
        .unwrap_or(0);

    if exists == 0 {
        return "You need to create a quickshab first by typing !quickshab [number of people]"
            .to_string();
    }

    // Update the number
    if let Err(_) = diesel::update(quickshab.filter(qs_chat_id.eq(chat_id)))
        .set(number.eq(num_people))
        .execute(db)
    {
        return "Error updating quickshab".to_string();
    }

    format_quickshab(db, chat_id)
}

fn show_cmd(db: &mut SqliteConnection, chat_id: &str) -> String {
    use crate::db::schema::quickshab::dsl::{chat_id as qs_chat_id, quickshab};
    use diesel::dsl::count_star;

    // Check if quickshab exists
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

// ============================================================================
// BRING/ASSIGN COMMANDS
// ============================================================================

fn bring_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str, attendee_id: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        return "Command syntax: !bring [category]".to_string();
    }

    use crate::db::schema::quickshab::dsl::{chat_id as qs_chat_id, quickshab};
    use crate::db::schema::quickshab_assignments::dsl::{
        category as qa_category, chat_id as qa_chat_id, name as qa_name, quickshab_assignments,
    };
    use diesel::dsl::count_star;

    // Check if quickshab exists
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

    // Insert assignment
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

fn assign_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 3 {
        return "Command syntax: !assign [name] [category]".to_string();
    }

    use crate::db::schema::quickshab::dsl::{chat_id as qs_chat_id, quickshab};
    use crate::db::schema::quickshab_assignments::dsl::{
        category as qa_category, chat_id as qa_chat_id, name as qa_name, quickshab_assignments,
    };
    use diesel::dsl::count_star;

    // Check if quickshab exists
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

    // Insert assignment
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

fn unbring_cmd(db: &mut SqliteConnection, chat_id: &str, attendee_id: &str) -> String {
    use crate::db::schema::quickshab::dsl::{chat_id as qs_chat_id, quickshab};
    use crate::db::schema::quickshab_assignments::dsl::{
        chat_id as qa_chat_id, id as qa_id, name as qa_name, quickshab_assignments,
    };
    use diesel::dsl::count_star;

    // Check if quickshab exists
    let exists: i64 = quickshab
        .filter(qs_chat_id.eq(chat_id))
        .select(count_star())
        .first(db)
        .unwrap_or(0);

    if exists == 0 {
        return "You need to create a quickshab first by typing !quickshab [number of people]"
            .to_string();
    }

    // Get the first assignment for this attendee
    let assignment_id: Option<i32> = quickshab_assignments
        .filter(qa_chat_id.eq(chat_id).and(qa_name.eq(attendee_id)))
        .select(qa_id)
        .first(db)
        .ok();

    match assignment_id {
        Some(id) => {
            // Delete that specific assignment
            let result = diesel::delete(quickshab_assignments.filter(qa_id.eq(id))).execute(db);
            match result {
                Ok(_) => format_quickshab(db, chat_id),
                Err(_) => "Error removing assignment".to_string(),
            }
        }
        None => "You don't have any assignments".to_string(),
    }
}

fn unassign_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
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

    // Check if quickshab exists
    let exists: i64 = quickshab
        .filter(qs_chat_id.eq(chat_id))
        .select(count_star())
        .first(db)
        .unwrap_or(0);

    if exists == 0 {
        return "You need to create a quickshab first by typing !quickshab [number of people]"
            .to_string();
    }

    // Get the first assignment for this name
    let assignment_id: Option<i32> = quickshab_assignments
        .filter(qa_chat_id.eq(chat_id).and(qa_name.eq(&name)))
        .select(qa_id)
        .first(db)
        .ok();

    match assignment_id {
        Some(id) => {
            // Delete that specific assignment
            let result = diesel::delete(quickshab_assignments.filter(qa_id.eq(id))).execute(db);
            match result {
                Ok(_) => format_quickshab(db, chat_id),
                Err(_) => "Error removing assignment".to_string(),
            }
        }
        None => format!("{} doesn't have any assignments", name),
    }
}

// ============================================================================
// SHOPPING LIST COMMANDS
// ============================================================================

fn shop_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        return "Usage: !shop <item> or !shop clear".to_string();
    }

    let item_text = args[1..].join(" ");

    use crate::db::schema::shopping_list::dsl::{
        chat_id as sl_chat_id, id as sl_id, item as sl_item, quantity as sl_quantity, shopping_list,
    };

    if item_text.to_lowercase() == "clear" {
        if let Err(_) = diesel::delete(shopping_list.filter(sl_chat_id.eq(chat_id))).execute(db) {
            return "Error clearing shopping list".to_string();
        }
        return "Shopping list cleared.".to_string();
    }

    let quantity_regex = Regex::new(r"\d+").unwrap();
    let quantity: i32 = quantity_regex
        .find(&item_text)
        .and_then(|m| m.as_str().parse().ok())
        .unwrap_or(0);

    let item_text = quantity_regex
        .replace_all(&item_text, "")
        .trim()
        .to_string();

    if item_text.is_empty() {
        return "Usage: !shop <item>".to_string();
    }

    let _ = get_or_create_chat(db, chat_id);

    // Check if item exists (case insensitive) - load all items and filter
    let all_items: Vec<(i32, String, i32)> = shopping_list
        .filter(sl_chat_id.eq(chat_id))
        .select((sl_id, sl_item, sl_quantity))
        .load::<(i32, String, i32)>(db)
        .unwrap_or_default();

    let existing = all_items
        .iter()
        .find(|(_, item, _)| item.to_lowercase() == item_text.to_lowercase())
        .map(|(id, _, qty)| (*id, *qty));

    if let Some((id, _)) = existing {
        if quantity > 0 {
            if let Err(_) = diesel::update(shopping_list.filter(sl_id.eq(id)))
                .set(sl_quantity.eq(quantity))
                .execute(db)
            {
                return "Error updating item".to_string();
            }
            return format!("\"{}\" quantity updated to {}.", item_text, quantity);
        }
        return format!("\"{}\" is already in the shopping list.", item_text);
    }

    // Add new item
    if let Err(_) = diesel::insert_into(shopping_list)
        .values((
            sl_chat_id.eq(chat_id),
            sl_item.eq(&item_text),
            sl_quantity.eq(quantity),
        ))
        .execute(db)
    {
        return "Error adding item to shopping list".to_string();
    }

    let quantity_str = if quantity > 0 {
        format!("{} x ", quantity)
    } else {
        String::new()
    };

    let verb = if quantity > 1 { "have" } else { "has" };

    format!(
        "{}\"{}\" {} been added to shopping list",
        quantity_str, item_text, verb
    )
}

fn shoplist_cmd(db: &mut SqliteConnection, chat_id: &str) -> String {
    use crate::db::schema::shopping_list::dsl::{
        chat_id as sl_chat_id, item as sl_item, quantity as sl_quantity, shopping_list,
    };

    // Get all items for this chat
    let items: Vec<(String, i32)> = shopping_list
        .filter(sl_chat_id.eq(chat_id))
        .select((sl_item, sl_quantity))
        .order(crate::db::schema::shopping_list::id)
        .load::<(String, i32)>(db)
        .unwrap_or_default();

    if items.is_empty() {
        return "Your shopping list is empty.".to_string();
    }

    let formatted_items: Vec<String> = items
        .iter()
        .enumerate()
        .map(|(i, (item, quantity))| {
            let quantity_str = if *quantity > 0 {
                format!("{} x ", quantity)
            } else {
                String::new()
            };
            format!("  {}.  {}{}", i + 1, quantity_str, item)
        })
        .collect();

    format!("Shopping list:\n{}", formatted_items.join("\n"))
}

fn unshop_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        return "Usage: !unshop <item>".to_string();
    }

    let item_to_remove = args[1..].join(" ");

    use crate::db::schema::shopping_list::dsl::{
        chat_id as sl_chat_id, id as sl_id, item as sl_item, shopping_list,
    };
    use diesel::dsl::count_star;

    // Check if list is empty
    let count: i64 = shopping_list
        .filter(sl_chat_id.eq(chat_id))
        .select(count_star())
        .first(db)
        .unwrap_or(0);

    if count == 0 {
        return "Your shopping list is empty.".to_string();
    }

    // Check if it's a number (index)
    if let Ok(index) = item_to_remove.parse::<usize>() {
        // Get the item at this index
        let items: Vec<(i32, String)> = shopping_list
            .filter(sl_chat_id.eq(chat_id))
            .select((sl_id, sl_item))
            .order(sl_id)
            .load::<(i32, String)>(db)
            .unwrap_or_default();

        if index == 0 || index > items.len() {
            return "Invalid item index.".to_string();
        }

        let (item_id, item_name) = &items[index - 1];

        if let Err(_) = diesel::delete(shopping_list.filter(sl_id.eq(item_id))).execute(db) {
            return "Error removing item".to_string();
        }
        return format!(
            "\"{}\" has been removed from your shopping list.",
            item_name
        );
    }

    // Remove by name (partial match using LIKE) - load all items and filter
    let all_items: Vec<(i32, String)> = shopping_list
        .filter(sl_chat_id.eq(chat_id))
        .select((sl_id, sl_item))
        .load::<(i32, String)>(db)
        .unwrap_or_default();

    let matching_item = all_items
        .iter()
        .find(|(_, item)| item.to_lowercase().contains(&item_to_remove.to_lowercase()))
        .map(|(id, item)| (*id, item.clone()));

    if let Some((item_id, item_name)) = matching_item {
        if let Err(_) = diesel::delete(shopping_list.filter(sl_id.eq(item_id))).execute(db) {
            return "Error removing item".to_string();
        }
        format!(
            "\"{}\" has been removed from your shopping list.",
            item_name
        )
    } else {
        format!("Item \"{}\" not found in shopping list.", item_to_remove)
    }
}

// ============================================================================
// REMINDER COMMANDS
// ============================================================================

fn remind_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str, _timezone: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        return "command usage: !remind   thing to remind   when".to_string();
    }

    let reminder_text = args[1..].join(" ");
    let _ = get_or_create_chat(db, chat_id);

    use crate::db::schema::reminders::dsl::{
        chat_id as r_chat_id, id as r_id, message as r_message, reminders,
        snoozable as r_snoozable, time as r_time, type_,
    };

    // For now, create as untimed reminder
    // TODO: Parse time from reminder_text
    let id = Uuid::new_v4().to_string();
    let message = reminder_text;

    if let Err(e) = diesel::insert_into(reminders)
        .values((
            r_id.eq(&id),
            r_chat_id.eq(chat_id),
            r_message.eq(&message),
            r_time.eq(&0),
            type_.eq("timeless"),
            r_snoozable.eq(&0i32),
        ))
        .execute(db)
    {
        return format!("Error creating reminder: {}", e);
    }

    format!("Ok, \"{}\" will be added to your reminders", message)
}

fn reminders_cmd(db: &mut SqliteConnection, chat_id: &str, _timezone: &str) -> String {
    use crate::db::schema::reminders::dsl::{
        chat_id as r_chat_id, message as r_message, reminders, time as r_time, type_,
    };

    // Fetch timed reminders
    let timed_reminders: Vec<(String, i32)> = reminders
        .filter(
            r_chat_id
                .eq(chat_id)
                .and(type_.ne("timeless"))
                .and(r_time.gt(0)),
        )
        .select((r_message, r_time))
        .order(r_time.asc())
        .load::<(String, i32)>(db)
        .unwrap_or_default();

    let timed_formatted: Vec<String> = timed_reminders
        .iter()
        .map(|(message, time)| {
            let dt = chrono::DateTime::from_timestamp(*time as i64, 0)
                .unwrap_or_else(|| chrono::DateTime::from_timestamp(0, 0).unwrap());
            let time_str = dt.format("%a @ %I:%M %p").to_string();
            format!("  • {} - {}", message, time_str)
        })
        .collect();

    // Fetch untimed reminders
    let untimed_reminders: Vec<String> = reminders
        .filter(
            r_chat_id
                .eq(chat_id)
                .and(type_.eq("timeless").or(r_time.eq(0))),
        )
        .select(r_message)
        .load::<String>(db)
        .unwrap_or_default();

    let untimed_formatted: Vec<String> = untimed_reminders
        .iter()
        .map(|message| format!("  • {}", message))
        .collect();

    if timed_formatted.is_empty() && untimed_formatted.is_empty() {
        return "You have no reminders.".to_string();
    }

    let mut result = String::new();
    if !timed_formatted.is_empty() {
        result.push_str("Timed Reminders:\n");
        result.push_str(&timed_formatted.join("\n"));
    }
    if !untimed_formatted.is_empty() {
        if !result.is_empty() {
            result.push_str("\n\n");
        }
        result.push_str("Todo:\n");
        result.push_str(&untimed_formatted.join("\n"));
    }

    result
}

fn snooze_cmd(
    _db: &mut SqliteConnection,
    _prompt: &str,
    _chat_id: &str,
    _timezone: &str,
) -> String {
    "Snooze functionality not yet implemented".to_string()
}

fn done_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        return "Command usage: !done [reminder query]".to_string();
    }

    let keyword = args[1..].join(" ");

    use crate::db::schema::reminders::dsl::{
        chat_id as r_chat_id, id as r_id, message as r_message, reminders,
    };

    // Find matching reminder (partial match)
    let all_reminders: Vec<(String, String)> = reminders
        .filter(r_chat_id.eq(chat_id))
        .select((r_id, r_message))
        .load::<(String, String)>(db)
        .unwrap_or_default();

    let matching = all_reminders
        .iter()
        .find(|(_, msg)| msg.to_lowercase().contains(&keyword.to_lowercase()))
        .map(|(id, _)| id.clone());

    match matching {
        Some(id) => {
            let result = diesel::delete(reminders.filter(r_id.eq(&id))).execute(db);
            match result {
                Ok(_) => "Reminder marked as done".to_string(),
                Err(e) => format!("Error: {}", e),
            }
        }
        None => format!("No reminder found matching \"{}\"", keyword),
    }
}

// ============================================================================
// SEND/SCHEDULED COMMANDS
// ============================================================================

fn send_cmd(_db: &mut SqliteConnection, _prompt: &str, _chat_id: &str, _timezone: &str) -> String {
    "Send command requires time parsing implementation".to_string()
}

fn unsend_cmd(db: &mut SqliteConnection, chat_id: &str) -> String {
    use crate::db::schema::reminders::dsl::{
        chat_id as r_chat_id, id as r_id, message as r_message, reminders, time as r_time, type_,
    };

    // Get the most recent scheduled message
    let result: Option<(String, String)> = reminders
        .filter(r_chat_id.eq(chat_id).and(type_.eq("send")))
        .select((r_id, r_message))
        .order(r_time.desc())
        .first::<(String, String)>(db)
        .ok();

    match result {
        Some((id, message)) => {
            if let Err(_) = diesel::delete(reminders.filter(r_id.eq(&id))).execute(db) {
                return "Error canceling scheduled message".to_string();
            }
            format!("Canceled scheduled message: \"{}\"", message)
        }
        None => "No scheduled messages found".to_string(),
    }
}

fn scheduled_cmd(db: &mut SqliteConnection, chat_id: &str) -> String {
    use crate::db::schema::reminders::dsl::{
        chat_id as r_chat_id, message as r_message, reminders, time as r_time, type_,
    };

    let messages: Vec<(String, i32)> = reminders
        .filter(r_chat_id.eq(chat_id).and(type_.eq("send")))
        .select((r_message, r_time))
        .order(r_time.asc())
        .load::<(String, i32)>(db)
        .unwrap_or_default();

    if messages.is_empty() {
        return "No scheduled messages.".to_string();
    }

    let formatted: Vec<String> = messages
        .iter()
        .map(|(message, time)| {
            let dt = chrono::DateTime::from_timestamp(*time as i64, 0)
                .unwrap_or_else(|| chrono::DateTime::from_timestamp(0, 0).unwrap());
            let time_str = dt.format("%a @ %I:%M %p").to_string();
            format!("  • {} - {}", message, time_str)
        })
        .collect();

    format!("Scheduled Messages:\n{}", formatted.join("\n"))
}
