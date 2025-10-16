use crate::db::Reminder;
use crate::db::schema::reminders as reminders_table;
use chrono::Utc;
use diesel::prelude::*;
use diesel::sqlite::SqliteConnection;
use log::{error, info};
use std::sync::Arc;
use std::time::Duration;
use tokio::time::sleep;
use waproto::whatsapp as wa;
use whatsapp_rust::client::Client;

/// Start the reminder checking routine
/// Checks for due reminders every 5 seconds
pub async fn start_reminder_routine(client: Arc<Client>, db_path: String) {
    info!("Starting reminder routine...");

    loop {
        // Sleep for 5 seconds
        sleep(Duration::from_secs(5)).await;

        // Check and send reminders
        if let Err(e) = check_and_send_reminders(&client, &db_path).await {
            error!("Error in reminder routine: {}", e);
        }
    }
}

/// Check for due reminders and send them
async fn check_and_send_reminders(
    client: &Arc<Client>,
    db_path: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    // Connect to database
    let mut conn = SqliteConnection::establish(db_path)?;

    let now = Utc::now().timestamp() as i32;

    // Query for reminders whose time has passed and haven't been sent yet
    use reminders_table::dsl::*;
    let due_reminders: Vec<Reminder> = reminders
        .filter(time.gt(0))
        .filter(time.le(now))
        .filter(sent_time.eq(0))
        .order(time.asc())
        .load::<Reminder>(&mut conn)?;

    if due_reminders.is_empty() {
        return Ok(());
    }

    info!("Found {} due reminder(s)", due_reminders.len());

    // Send each reminder
    for reminder in due_reminders {
        if let Err(e) = send_reminder(client, &mut conn, reminder).await {
            error!("Failed to send reminder: {}", e);
        }
    }

    Ok(())
}

/// Send a single reminder
async fn send_reminder(
    client: &Arc<Client>,
    conn: &mut SqliteConnection,
    reminder: Reminder,
) -> Result<(), Box<dyn std::error::Error>> {
    // Parse the chat ID (JID)
    let chat_jid = match reminder.chat_id.parse() {
        Ok(jid) => jid,
        Err(e) => {
            error!("Invalid chat ID {}: {}", reminder.chat_id, e);
            // Delete invalid reminder
            use reminders_table::dsl::*;
            diesel::delete(reminders.filter(id.eq(&reminder.id))).execute(conn)?;
            return Err(format!("Invalid chat ID: {}", e).into());
        }
    };

    // Format message based on type
    let message_text = if reminder.type_ == "send" {
        reminder.message.clone()
    } else {
        format!("⏰ Reminder: {}", reminder.message)
    };

    // Create the message
    let message = wa::Message {
        conversation: Some(message_text.clone()),
        ..Default::default()
    };

    // Send the message
    info!("Sending reminder to {}: {}", reminder.chat_id, message_text);

    match client.send_message(chat_jid, message).await {
        Ok(_) => {
            info!("Successfully sent reminder {}", reminder.id);

            // Update database based on reminder type
            use reminders_table::dsl::*;

            if reminder.type_ == "send" {
                // Delete send-type reminders after sending
                diesel::delete(reminders.filter(id.eq(&reminder.id))).execute(conn)?;
            } else {
                // Mark remind-type reminders as sent
                let current_time = Utc::now().timestamp() as i32;
                diesel::update(reminders.filter(id.eq(&reminder.id)))
                    .set((
                        sent_time.eq(current_time),
                        snoozable.eq(1), // Make it snoozable
                    ))
                    .execute(conn)?;
            }

            Ok(())
        }
        Err(e) => {
            error!("Failed to send message to {}: {}", reminder.chat_id, e);
            Err(e.into())
        }
    }
}

/// Clean up old sent reminders (older than 24 hours)
pub async fn cleanup_old_reminders(db_path: &str) -> Result<(), Box<dyn std::error::Error>> {
    let mut conn = SqliteConnection::establish(db_path)?;

    let cutoff_time = (Utc::now().timestamp() - 86400) as i32; // 24 hours ago

    use reminders_table::dsl::*;
    let deleted_count = diesel::delete(
        reminders
            .filter(sent_time.gt(0))
            .filter(sent_time.lt(cutoff_time)),
    )
    .execute(&mut conn)?;

    if deleted_count > 0 {
        info!("Cleaned up {} old reminders", deleted_count);
    }

    Ok(())
}
