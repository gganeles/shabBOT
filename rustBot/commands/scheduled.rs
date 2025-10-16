use crate::db::*;
use crate::utils::*;
use diesel::prelude::*;
use uuid::Uuid;

pub fn send_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str, timezone: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        return "command usage: !send <message> <datetime>".to_string();
    }

    let full_text = args[1..].join(" ");
    let _ = get_or_create_chat(db, chat_id);

    // Extract datetime from the text with timezone awareness
    let (message, datetime) = match extract_datetime_from_text(&full_text, timezone) {
        Ok((remaining_text, dt)) => {
            if remaining_text.is_empty() {
                return "Error: message text cannot be empty".to_string();
            }
            (remaining_text, dt)
        }
        Err(e) => {
            return format!(
                "Error parsing datetime: {}. Usage: !send <message> <datetime>",
                e
            );
        }
    };

    use crate::db::schema::reminders::dsl::{
        chat_id as r_chat_id, id as r_id, message as r_message, reminders,
        snoozable as r_snoozable, time as r_time, type_,
    };

    let id = Uuid::new_v4().to_string();
    let timestamp = datetime.timestamp() as i32;

    if let Err(e) = diesel::insert_into(reminders)
        .values((
            r_id.eq(&id),
            r_chat_id.eq(chat_id),
            r_message.eq(&message),
            r_time.eq(&timestamp),
            type_.eq("send"),
            r_snoozable.eq(&0i32),
        ))
        .execute(db)
    {
        return format!("Error scheduling message: {}", e);
    }

    let time_str = datetime.format("%a %b %d @ %I:%M %p").to_string();
    format!("Message \"{}\" scheduled for {}", message, time_str)
}

pub fn unsend_cmd(db: &mut SqliteConnection, chat_id: &str) -> String {
    use crate::db::schema::reminders::dsl::{
        chat_id as r_chat_id, id as r_id, message as r_message, reminders, time as r_time, type_,
    };

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

pub fn scheduled_cmd(db: &mut SqliteConnection, chat_id: &str) -> String {
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
