use crate::db::*;
use crate::utils::*;
use chrono::{DateTime, Utc};
use diesel::prelude::*;
use uuid::Uuid;

/// Format relative time like "today", "tomorrow", "in 3 days", "next week", etc.
fn format_relative_time<Tz1, Tz2>(target_time: &DateTime<Tz1>, now: &DateTime<Tz2>) -> String
where
    Tz1: chrono::TimeZone,
    Tz2: chrono::TimeZone,
{
    // Calculate difference
    let target_naive = target_time.naive_utc();
    let now_naive = now.naive_utc();
    let diff = target_naive.signed_duration_since(now_naive);

    // For past times
    if diff.num_seconds() < 0 {
        return String::new();
    }

    // Same day - compare using naive dates
    if target_naive.date() == now_naive.date() {
        return "today".to_string();
    }

    // Tomorrow
    let tomorrow_naive = now_naive.date() + chrono::Duration::days(1);
    if target_naive.date() == tomorrow_naive {
        return "tomorrow".to_string();
    }

    // Within the next week (2-7 days)
    let days = diff.num_days();
    if days >= 2 && days <= 7 {
        return format!("in {} days", days);
    }

    // Next week (8-14 days)
    if days >= 8 && days <= 14 {
        return "next week".to_string();
    }

    // Weeks (15-60 days)
    if days >= 15 && days <= 60 {
        let weeks = (days + 3) / 7; // Round to nearest week
        if weeks == 2 {
            return "in 2 weeks".to_string();
        }
        return format!("in {} weeks", weeks);
    }

    // Months
    if days >= 61 && days <= 365 {
        let months = (days + 15) / 30; // Rough month approximation
        if months == 1 {
            return "next month".to_string();
        }
        return format!("in {} months", months);
    }

    // Years
    if days > 365 {
        let years = (days + 180) / 365;
        if years == 1 {
            return "next year".to_string();
        }
        return format!("in {} years", years);
    }

    String::new()
}

pub fn remind_cmd(
    db: &mut SqliteConnection,
    prompt: &str,
    chat_id: &str,
    timezone: &str,
) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        return "command usage: !remind <text> [datetime]".to_string();
    }

    let full_text = args[1..].join(" ");
    let _ = get_or_create_chat(db, chat_id);

    use crate::db::schema::reminders::dsl::{
        chat_id as r_chat_id, id as r_id, message as r_message, reminders,
        snoozable as r_snoozable, time as r_time, type_,
    };

    let id = Uuid::new_v4().to_string();

    // Try to extract datetime from the text with timezone awareness
    let (message, reminder_time, reminder_type) =
        match extract_datetime_from_text(&full_text, timezone) {
            Ok((remaining_text, datetime)) => {
                // Found a datetime - create timed reminder (already cleaned by extract_datetime_from_text)
                let timestamp = datetime.timestamp() as i32;
                (remaining_text, timestamp, "timed")
            }
            Err(_) => {
                // No datetime found - create timeless reminder, clean the text
                let cleaned = clean_reminder_text(&full_text);
                (cleaned, 0, "timeless")
            }
        };

    if message.is_empty() || message == "nothing" {
        return "Error: reminder text cannot be empty".to_string();
    }

    if let Err(e) = diesel::insert_into(reminders)
        .values((
            r_id.eq(&id),
            r_chat_id.eq(chat_id),
            r_message.eq(&message),
            r_time.eq(&reminder_time),
            type_.eq(reminder_type),
            r_snoozable.eq(&0i32),
        ))
        .execute(db)
    {
        return format!("Error creating reminder: {}", e);
    }

    if reminder_type == "timed" {
        // Parse the timezone to show confirmation in user's local time
        use chrono_tz::Tz;
        let tz: Tz = timezone.parse().unwrap_or(chrono_tz::Asia::Jerusalem);

        let dt = DateTime::<Utc>::from_timestamp(reminder_time as i64, 0)
            .unwrap_or_else(|| DateTime::<Utc>::from_timestamp(0, 0).unwrap())
            .with_timezone(&tz);
        let time_str = dt.format("%a %b %d @ %I:%M %p").to_string();
        format!("Ok, \"{}\" scheduled for {}", message, time_str)
    } else {
        format!("Ok, \"{}\" added to your reminders", message)
    }
}

pub fn reminders_cmd(db: &mut SqliteConnection, chat_id: &str, timezone: &str) -> String {
    use crate::db::schema::reminders::dsl::{
        chat_id as r_chat_id, message as r_message, reminders, time as r_time, type_,
    };

    // Parse the timezone
    use chrono_tz::Tz;
    let tz: Tz = timezone.parse().unwrap_or(chrono_tz::Asia::Jerusalem);

    // Get current time
    let now_utc = chrono::Utc::now();
    let now_in_tz = now_utc.with_timezone(&tz);
    let now_timestamp = now_utc.timestamp();

    // Fetch timed reminders ordered by time (earliest first)
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
                .unwrap_or_else(|| chrono::DateTime::from_timestamp(0, 0).unwrap())
                .with_timezone(&tz); // Convert to user's timezone

            let time_str = dt.format("%a @ %I:%M %p").to_string();
            let relative_str = format_relative_time(&dt, &now_in_tz);

            // Check if past
            if (*time as i64) < now_timestamp {
                format!("  • {} - {} - past", message, time_str)
            } else if !relative_str.is_empty() {
                format!("  • {} - {} ({})", message, time_str, relative_str)
            } else {
                format!("  • {} - {}", message, time_str)
            }
        })
        .collect();

    // Fetch untimed reminders ordered by ROWID (insertion order, oldest first)
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

pub fn snooze_cmd(
    db: &mut SqliteConnection,
    prompt: &str,
    chat_id: &str,
    timezone: &str,
) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        return "Command usage: !snooze [duration] [optional reminder query]".to_string();
    }

    let full_text = args[1..].join(" ");

    // Parse snooze duration from arguments
    let (new_time, search_query) = match extract_datetime_from_text(&full_text, timezone) {
        Ok((remaining, dt)) => (dt, remaining),
        Err(_) => {
            return "Could not parse snooze duration. Try something like '10 minutes' or '2 hours'"
                .to_string();
        }
    };

    use crate::db::schema::reminders::dsl::{
        chat_id as r_chat_id, id as r_id, message as r_message, reminders,
        sent_time as r_sent_time, snoozable as r_snoozable, time as r_time, type_,
    };

    // Query all reminders for this chat (type = 'remind')
    let all_reminders: Vec<(String, String, i32)> = reminders
        .filter(r_chat_id.eq(chat_id).and(type_.eq("remind")))
        .select((r_id, r_message, r_time))
        .order(r_time.asc())
        .load::<(String, String, i32)>(db)
        .unwrap_or_default();

    if all_reminders.is_empty() {
        return "No reminders found to snooze.".to_string();
    }

    // Build search objects
    let search_objects: Vec<SearchObject> = all_reminders
        .iter()
        .enumerate()
        .map(|(index, (id, message, _))| SearchObject {
            id: id.clone(),
            message: message.clone(),
            index,
        })
        .collect();

    // Find matching reminders using whatKeys
    let matching_ids = what_keys(&search_query, &search_objects);

    let reminder_ids: Vec<String> = if matching_ids.is_empty() {
        // If search query was provided but no matches
        if !search_query.trim().is_empty() {
            return "No matching reminders found to snooze.".to_string();
        }
        // Default to earliest reminder (first in list)
        vec![search_objects[0].id.clone()]
    } else {
        matching_ids
    };

    // Update all matching reminders
    let mut snoozed_messages = Vec::new();
    let mut success_count = 0;

    for reminder_id in &reminder_ids {
        // Find the message for this ID
        let message = all_reminders
            .iter()
            .find(|(id, _, _)| id == reminder_id)
            .map(|(_, msg, _)| msg.clone())
            .unwrap_or_default();

        // Update reminder with new time and reset sent_time and snoozable flag
        let result = diesel::update(reminders.filter(r_id.eq(reminder_id)))
            .set((
                r_time.eq(new_time.timestamp() as i32),
                r_sent_time.eq(0),
                r_snoozable.eq(0),
            ))
            .execute(db);

        if result.is_ok() {
            success_count += 1;
            snoozed_messages.push(message);
        }
    }

    if success_count == 0 {
        return "Error snoozing reminders".to_string();
    }

    // Parse timezone for display
    use chrono_tz::Tz;
    let tz: Tz = timezone.parse().unwrap_or(chrono_tz::Asia::Jerusalem);
    let display_time = new_time.with_timezone(&tz);
    let time_str = display_time.format("%a @ %I:%M %p").to_string();

    // Format response based on number of reminders snoozed
    if success_count == 1 {
        format!(
            "Reminder \"{}\" snoozed until {}",
            snoozed_messages[0], time_str
        )
    } else {
        format!(
            "{} reminders snoozed until {}:\n  • {}",
            success_count,
            time_str,
            snoozed_messages.join("\n  • ")
        )
    }
}

pub fn done_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
    let args = parse_args(prompt);

    use crate::db::schema::reminders::dsl::{
        chat_id as r_chat_id, id as r_id, message as r_message, reminders, time as r_time, type_,
    };

    // Query all reminders for this chat (both timed and untimed)
    let all_reminders: Vec<(String, String, i32, String)> = reminders
        .filter(r_chat_id.eq(chat_id))
        .select((r_id, r_message, r_time, type_))
        .order((r_time.desc(), r_id.desc()))
        .load::<(String, String, i32, String)>(db)
        .unwrap_or_default();

    if all_reminders.is_empty() {
        return "No reminders found".to_string();
    }

    // Build search objects
    let search_objects: Vec<SearchObject> = all_reminders
        .iter()
        .enumerate()
        .map(|(index, (id, message, _, _))| SearchObject {
            id: id.clone(),
            message: message.clone(),
            index,
        })
        .collect();

    // If user provided search terms, use them
    let reminder_ids: Vec<String> = if args.len() > 1 {
        // User provided search query
        let search_query = args[1..].join(" ");
        let matching_ids = what_keys(&search_query, &search_objects);

        if matching_ids.is_empty() {
            return "Could not find matching reminder(s). Try: !done or !done [reminder text]"
                .to_string();
        }
        matching_ids
    } else {
        // Default to most recent reminder (first in list due to DESC order)
        vec![search_objects[0].id.clone()]
    };

    // Delete all matching reminders
    let mut deleted_messages = Vec::new();
    let mut success_count = 0;

    for reminder_id in &reminder_ids {
        // Find the message for this ID
        let message = all_reminders
            .iter()
            .find(|(id, _, _, _)| id == reminder_id)
            .map(|(_, msg, _, _)| msg.clone())
            .unwrap_or_default();

        // Delete the reminder
        let result = diesel::delete(reminders.filter(r_id.eq(reminder_id))).execute(db);

        if result.is_ok() {
            success_count += 1;
            deleted_messages.push(message);
        }
    }

    if success_count == 0 {
        return "Error removing reminders".to_string();
    }

    // Format response based on number of reminders deleted
    if success_count == 1 {
        format!(
            "Reminder \"{}\" marked as done and removed",
            deleted_messages[0]
        )
    } else {
        format!(
            "{} reminders marked as done and removed:\n  • {}",
            success_count,
            deleted_messages.join("\n  • ")
        )
    }
}
