# Reminder Routine Implementation

## Overview
This document describes the background reminder checking and sending system that was added to the Rust WhatsApp bot.

## Implementation Summary

### Files Created
- **`reminder_routine.rs`**: New module containing the reminder checking logic

### Files Modified
- **`main.rs`**: Updated to spawn the reminder routine as a background task

## How It Works

### Background Thread
The reminder routine runs in a separate async task that:
1. Checks for due reminders every 5 seconds
2. Queries the database for reminders where:
   - `time > 0` (not timeless)
   - `time <= now` (due or past due)
   - `sent_time = 0` (not yet sent)
3. Sends each due reminder via WhatsApp
4. Updates the database appropriately

### Message Sending
When a reminder is due:
- **For "remind" type**: Message is prefixed with "⏰ Reminder: " and marked as sent (but kept in database for snoozing)
- **For "send" type**: Message is sent as-is and deleted from database immediately after sending

### Database Updates
After successfully sending:
- **"remind" type reminders**: 
  - `sent_time` set to current timestamp
  - `snoozable` set to 1 (allows snoozing)
- **"send" type reminders**: 
  - Deleted from database

### Error Handling
- Invalid chat IDs are logged and the reminder is deleted
- Failed message sends are logged but don't stop other reminders
- Database errors are logged and the routine continues

## Matching Go Implementation

The Rust implementation matches the Go version (`timeRoutine.go`) in:
1. 5-second check interval
2. Database query logic (same WHERE conditions)
3. Message formatting (emoji for reminders, plain for sends)
4. Reminder deletion/update logic
5. Error handling approach

## Key Features

### Concurrency Safe
- Uses separate database connections per check
- No long-running transactions that could block the main bot

### Non-blocking
- Runs in its own tokio task
- Doesn't interfere with message handling
- Errors in reminder sending don't crash the bot

### Logging
- Info logs for reminders found and sent
- Error logs for failures with context

## Usage

The reminder routine starts automatically when the bot starts:
```rust
// In main.rs, after building the bot
let client_for_reminders = bot.client().clone();

tokio::spawn(async move {
    reminder_routine::start_reminder_routine(client_for_reminders, "shabbot.db".to_string())
        .await;
});
```

No manual intervention needed - it runs continuously in the background.

## Future Enhancements

Optional additions that could be made:
1. Cleanup routine for old sent reminders (>24 hours) - function already exists: `cleanup_old_reminders()`
2. Retry logic for failed sends
3. Configurable check interval
4. Metrics/monitoring for reminder sending success rate

## Testing

To test:
1. Add a reminder via the bot commands (e.g., `!remind me to test in 1 minute`)
2. Wait for the due time
3. Check logs for "Found X due reminder(s)" and "Successfully sent reminder"
4. Verify the reminder message was received in WhatsApp

## Dependencies

No new dependencies were added. Uses existing crates:
- `diesel` for database queries
- `chrono` for timestamps
- `tokio` for async runtime
- `log` for logging
- `whatsapp_rust` for sending messages
