use diesel::prelude::*;
use diesel::sqlite::SqliteConnection;

use crate::db::schema::*;

#[derive(Queryable, Insertable, AsChangeset, Debug)]
#[diesel(table_name = chats)]
pub struct Chat {
    pub chat_id: String,
    pub location: String,
    pub timezone: String,
}

#[derive(Queryable, Insertable, AsChangeset, Debug)]
#[diesel(table_name = quickshab)]
pub struct QuickShab {
    pub chat_id: String,
    pub number: i32,
}

#[derive(Queryable, Insertable, Debug)]
#[diesel(table_name = quickshab_assignments)]
pub struct QuickShabAssignment {
    pub id: i32,
    pub chat_id: String,
    pub category: String,
    pub name: String,
}

#[derive(Insertable, Debug)]
#[diesel(table_name = quickshab_assignments)]
pub struct NewQuickShabAssignment {
    pub chat_id: String,
    pub category: String,
    pub name: String,
}

#[derive(Queryable, Insertable, Debug)]
#[diesel(table_name = shopping_list)]
pub struct ShoppingItem {
    pub id: i32,
    pub chat_id: String,
    pub item: String,
    pub quantity: i32,
}

#[derive(Insertable, Debug)]
#[diesel(table_name = shopping_list)]
pub struct NewShoppingItem {
    pub chat_id: String,
    pub item: String,
    pub quantity: i32,
}

#[derive(Queryable, Insertable, Debug)]
#[diesel(table_name = reminders)]
pub struct Reminder {
    pub id: String,
    pub chat_id: String,
    pub message: String,
    pub time: i32,
    pub type_: String,
    pub snoozable: i32,
    pub sent_time: i32,
}

/// Initialize the database schema
pub fn init_db(conn: &mut SqliteConnection) -> Result<(), Box<dyn std::error::Error>> {
    diesel::sql_query(
        "CREATE TABLE IF NOT EXISTS chats (
            chat_id TEXT PRIMARY KEY,
            location TEXT DEFAULT 'Haifa',
            timezone TEXT DEFAULT 'Asia/Jerusalem'
        )",
    )
    .execute(conn)?;

    diesel::sql_query(
        "CREATE TABLE IF NOT EXISTS quickshab (
            chat_id TEXT PRIMARY KEY,
            number INTEGER DEFAULT 0,
            FOREIGN KEY (chat_id) REFERENCES chats(chat_id)
        )",
    )
    .execute(conn)?;

    diesel::sql_query(
        "CREATE TABLE IF NOT EXISTS quickshab_assignments (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            chat_id TEXT NOT NULL,
            category TEXT NOT NULL,
            name TEXT NOT NULL,
            FOREIGN KEY (chat_id) REFERENCES chats(chat_id)
        )",
    )
    .execute(conn)?;

    diesel::sql_query(
        "CREATE TABLE IF NOT EXISTS shopping_list (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            chat_id TEXT NOT NULL,
            item TEXT NOT NULL,
            quantity INTEGER DEFAULT 0,
            FOREIGN KEY (chat_id) REFERENCES chats(chat_id)
        )",
    )
    .execute(conn)?;

    diesel::sql_query(
        "CREATE TABLE IF NOT EXISTS reminders (
            id TEXT PRIMARY KEY,
            chat_id TEXT NOT NULL,
            message TEXT NOT NULL,
            time INTEGER DEFAULT 0,
            type TEXT NOT NULL,
            snoozable INTEGER DEFAULT 0,
            sent_time INTEGER DEFAULT 0,
            FOREIGN KEY (chat_id) REFERENCES chats(chat_id)
        )",
    )
    .execute(conn)?;

    Ok(())
}

/// Ensure a chat exists in the database
pub fn get_or_create_chat(
    conn: &mut SqliteConnection,
    chat_id: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    use crate::db::schema::chats::dsl;

    let existing = dsl::chats.find(chat_id).first::<Chat>(conn).optional()?;

    if existing.is_none() {
        let new_chat = Chat {
            chat_id: chat_id.to_string(),
            location: "Haifa".to_string(),
            timezone: "Asia/Jerusalem".to_string(),
        };
        diesel::insert_into(dsl::chats)
            .values(&new_chat)
            .execute(conn)?;
    }

    Ok(())
}

/// Get the location for a chat
pub fn get_chat_location(
    conn: &mut SqliteConnection,
    chat_id: &str,
) -> Result<String, Box<dyn std::error::Error>> {
    use crate::db::schema::chats::dsl;

    let chat = dsl::chats.find(chat_id).first::<Chat>(conn).optional()?;

    Ok(chat
        .map(|c| c.location)
        .unwrap_or_else(|| "Haifa".to_string()))
}

/// Get the timezone for a chat
pub fn get_chat_timezone(conn: &mut SqliteConnection, chat_id: &str) -> String {
    use crate::db::schema::chats::dsl;

    dsl::chats
        .find(chat_id)
        .first::<Chat>(conn)
        .map(|c| c.timezone)
        .unwrap_or_else(|_| "Asia/Jerusalem".to_string())
}

/// Set the location for a chat and update timezone based on geo data
pub fn set_chat_location(
    conn: &mut SqliteConnection,
    chat_id: &str,
    location: &str,
    timezone: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    use crate::db::schema::chats::dsl;

    get_or_create_chat(conn, chat_id)?;

    diesel::update(dsl::chats.find(chat_id))
        .set((dsl::location.eq(location), dsl::timezone.eq(timezone)))
        .execute(conn)?;

    Ok(())
}

/// Capitalize the first letter of a string
pub fn capitalize_first(s: &str) -> String {
    let mut chars = s.chars();
    match chars.next() {
        None => String::new(),
        Some(first) => first.to_uppercase().collect::<String>() + chars.as_str(),
    }
}
