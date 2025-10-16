// Integration test for reminder commands with datetime parsing
// Run with: cargo test --test test_reminder_commands

use diesel::prelude::*;
use diesel::sqlite::SqliteConnection;
use rustWABot::commands::reminders::*;
use rustWABot::commands::scheduled::*;

fn setup_test_db() -> SqliteConnection {
    let mut conn =
        SqliteConnection::establish(":memory:").expect("Failed to create in-memory database");

    // Create tables
    diesel::sql_query(
        "CREATE TABLE IF NOT EXISTS chats (
            id TEXT PRIMARY KEY,
            timezone TEXT
        )",
    )
    .execute(&mut conn)
    .expect("Failed to create chats table");

    diesel::sql_query(
        "CREATE TABLE IF NOT EXISTS reminders (
            id TEXT PRIMARY KEY,
            chat_id TEXT NOT NULL,
            message TEXT NOT NULL,
            time INTEGER NOT NULL,
            type TEXT NOT NULL,
            snoozable INTEGER NOT NULL DEFAULT 0
        )",
    )
    .execute(&mut conn)
    .expect("Failed to create reminders table");

    conn
}

#[test]
fn test_remind_with_datetime() {
    let mut db = setup_test_db();
    let response = remind_cmd(
        &mut db,
        "!remind buy milk tomorrow at 3pm",
        "test_chat",
        "UTC",
    );

    println!("Response: {}", response);
    assert!(response.contains("buy milk"));
    assert!(response.contains("scheduled for") || response.contains("added to your reminders"));
}

#[test]
fn test_remind_without_datetime() {
    let mut db = setup_test_db();
    let response = remind_cmd(&mut db, "!remind call mom", "test_chat", "UTC");

    println!("Response: {}", response);
    assert!(response.contains("call mom"));
    assert!(response.contains("added to your reminders"));
}

#[test]
fn test_send_with_datetime() {
    let mut db = setup_test_db();
    let response = send_cmd(
        &mut db,
        "!send Happy Birthday tomorrow at 10am",
        "test_chat",
        "UTC",
    );

    println!("Response: {}", response);
    assert!(response.contains("Happy Birthday"));
    assert!(response.contains("scheduled for"));
}

#[test]
fn test_send_without_datetime() {
    let mut db = setup_test_db();
    let response = send_cmd(&mut db, "!send Just a message", "test_chat", "UTC");

    println!("Response: {}", response);
    assert!(response.contains("Error parsing datetime"));
}

#[test]
fn test_reminders_list() {
    let mut db = setup_test_db();

    // Add a reminder
    remind_cmd(
        &mut db,
        "!remind test reminder tomorrow at 3pm",
        "test_chat",
        "UTC",
    );

    // List reminders
    let response = reminders_cmd(&mut db, "test_chat", "UTC");

    println!("Response: {}", response);
    assert!(response.contains("test reminder") || response.contains("Timed Reminders"));
}

#[test]
fn test_done_cmd() {
    let mut db = setup_test_db();

    // Add a reminder
    remind_cmd(&mut db, "!remind buy groceries", "test_chat", "UTC");

    // Mark as done
    let response = done_cmd(&mut db, "!done groceries", "test_chat");

    println!("Response: {}", response);
    assert!(response.contains("done") || response.contains("marked"));
}
