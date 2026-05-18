package main

import (
	"context"
	"database/sql"
	"fmt"
	"shabBOT/commands"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// TimeProvider is an interface for getting the current time (allows mocking in tests)
type TimeProvider interface {
	Now() time.Time
}

// RealTimeProvider implements TimeProvider using the actual time.Now()
type RealTimeProvider struct{}

func (RealTimeProvider) Now() time.Time {
	return time.Now()
}

func timeRoutine(client *whatsmeow.Client, db *sql.DB, ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	time_provider := RealTimeProvider{}
	fmt.Println("Time routine started")
	for range ticker.C {
		// Check for reminders that need to be sent
		checkAndSendReminders(client, db, ctx)
		// Clean up old sent reminders (older than 24 hours)
		cleanupOldReminders(db)
		// Example scheduled message at 18:19:00
		eighteenNineteen(client, ctx)

		checkAndSendExerciseResultsWithTime(client, db, ctx, time_provider)

		weeklyCleaning(client, ctx, db)

		checkAndSendTodos(client, ctx, db, time_provider)
	}
}

func checkAndSendTodos(client *whatsmeow.Client, ctx context.Context, db *sql.DB, timeProvider TimeProvider) {
	now := timeProvider.Now()

	jid_str := "9204232364276@lid"

	rows, err := db.Query(`SELECT timezone FROM chats WHERE chat_id = ?`, jid_str)
	if err != nil {
		fmt.Printf("Error querying timezone for chat %s: %v\n", jid_str, err)
		return
	}
	defer rows.Close()

	var timezone string
	if rows.Next() {
		if err := rows.Scan(&timezone); err != nil {
			fmt.Printf("Error scanning timezone: %v\n", err)
			return
		}
	} else {
		fmt.Printf("No timezone found for chat %s\n", jid_str)
		return
	}
	rows.Close()

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		fmt.Printf("Error loading location for timezone %s: %v\n", timezone, err)
		return
	}

	now = now.In(loc)

	if !((now.Hour() == 9 ||
		now.Hour() == 12 ||
		now.Hour() == 15 ||
		now.Hour() == 18 ||
		now.Hour() == 21 ||
		now.Hour() == 0) &&
		now.Minute() == 0 &&
		now.Second() < 5) {
		return
	}

	message := commands.RemindersCmd(db, jid_str, loc)
	// Send the message to the group chat
	jid, err := types.ParseJID(jid_str)
	if err != nil {
		fmt.Printf("Error parsing chat ID: %v\n", err)
		return
	}

	_, err = client.SendMessage(ctx, jid, &waE2E.Message{
		Conversation: &message,
	})

	if err != nil {
		fmt.Printf("Error sending to-do list: %v\n", err)
	} else {
		fmt.Printf("Sent to-do list at %s\n", now.Format("2006-01-02 15:04:05"))
	}
}

func weeklyCleaning(client *whatsmeow.Client, ctx context.Context, db *sql.DB) {
	haifa, _ := time.LoadLocation("Asia/Jerusalem")
	now := time.Now().In(haifa)

	hour := now.Hour()
	minute := now.Minute()
	seconds := now.Second()

	// Group chat ID
	jid, err := types.ParseJID("120363037678094725@g.us")
	if err != nil {
		fmt.Printf("Error parsing chat ID for weekly cleaning: %v\n", err)
		return
	}

	// Saturday night at 20:00 - Send progress report
	if now.Weekday() == time.Saturday && hour == 20 && minute == 0 && seconds < 5 {
		messageText := commands.FormatCleaningString(db, true)
		_, err = client.SendMessage(ctx, jid, &waE2E.Message{
			Conversation: &messageText,
		})
		if err != nil {
			fmt.Printf("Error sending weekly cleaning progress: %v\n", err)
		} else {
			fmt.Printf("Sent weekly cleaning progress at %s\n", now.Format("2006-01-02 15:04:05"))
		}
	}

	// Sunday at 9:00 - Assign new chores and send announcement
	if now.Weekday() == time.Sunday && hour == 9 && minute == 0 && seconds < 5 {
		// Assign new chores
		commands.AssignChores(db)

		// Send new assignments
		messageText := commands.FormatCleaningString(db, false)
		_, err = client.SendMessage(ctx, jid, &waE2E.Message{
			Conversation: &messageText,
		})
		if err != nil {
			fmt.Printf("Error sending weekly cleaning assignments: %v\n", err)
		} else {
			fmt.Printf("Sent weekly cleaning assignments at %s\n", now.Format("2006-01-02 15:04:05"))
		}
	}
}

func checkAndSendExerciseResults(client *whatsmeow.Client, db *sql.DB, ctx context.Context) {
	checkAndSendExerciseResultsWithTime(client, db, ctx, RealTimeProvider{})
}

func checkAndSendExerciseResultsWithTime(client *whatsmeow.Client, db *sql.DB, ctx context.Context, timeProvider TimeProvider) {

	rows, err := db.Query(`
		SELECT DISTINCT
			e.chat_id, c.timezone
		FROM
			exercise_counter AS e
		JOIN
			chats as c ON e.chat_id = c.chat_id
		`)

	if err != nil {
		fmt.Printf("Error querying exercise counters: %v\n", err)
		return
	}

	// Collect all chat data into a slice first, then close rows before processing
	type chatData struct {
		chatID   string
		timezone string
	}
	var chats []chatData

	for rows.Next() {
		var cd chatData
		rows.Scan(&cd.chatID, &cd.timezone)
		chats = append(chats, cd)
	}
	rows.Close() // Close rows before doing any updates

	for _, cd := range chats {
		timezone, err := time.LoadLocation(cd.timezone)
		if err != nil {
			fmt.Printf("Error loading timezone for chat %s: %v\n", cd.chatID, err)
			continue
		}

		now := timeProvider.Now().In(timezone)

		if time.Sunday != now.Weekday() {
			continue
		}

		hour := now.Hour()
		minute := now.Minute()
		seconds := now.Second()
		jid, err := types.ParseJID(cd.chatID)

		if err != nil {
			fmt.Printf("Error parsing chat ID %s: %v\n", cd.chatID, err)
			continue
		}

		if hour == 0 && minute == 0 && seconds < 5 {
			// Send exercise results message
			messageText := "🏋️‍♂️ Exercise Results:\n\n" + commands.GetExerciseResults(db, cd.chatID)
			_, err = client.SendMessage(ctx, jid, &waE2E.Message{
				Conversation: &messageText,
			})

			if err != nil {
				fmt.Printf("Error sending exercise results to %s: %v\n", cd.chatID, err)
			}

			_, err = db.Exec(`UPDATE exercise_counter SET value = 0 WHERE chat_id = ?`, cd.chatID)

			if err != nil {
				fmt.Printf("Error resetting exercise counter for %s: %v\n", cd.chatID, err)
			}
		}
	}
}

func checkAndSendReminders(client *whatsmeow.Client, db *sql.DB, ctx context.Context) {
	now := time.Now().Unix()

	// Use a transaction for better concurrency control
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("Error starting transaction: %v\n", err)
		return
	}
	defer tx.Rollback() // Will be no-op if transaction is committed

	// Query for reminders whose time has passed (and are not timeless reminders)
	// Exclude reminders that have already been sent (sent_time > 0)
	rows, err := tx.Query(`
		SELECT id, chat_id, message, type
		FROM reminders
		WHERE time > 0 AND time <= ? AND sent_time = 0
		ORDER BY time ASC
	`, now)

	if err != nil {
		fmt.Printf("Error querying reminders: %v\n", err)
		return
	}
	defer rows.Close()

	// Collect reminders to send
	type reminderToSend struct {
		ID           string
		ChatID       string
		Message      string
		ReminderType string
	}
	var reminders []reminderToSend

	for rows.Next() {
		var r reminderToSend
		if err := rows.Scan(&r.ID, &r.ChatID, &r.Message, &r.ReminderType); err != nil {
			fmt.Printf("Error scanning reminder: %v\n", err)
			continue
		}
		reminders = append(reminders, r)
	}
	rows.Close()

	// Commit transaction to release the lock before sending messages
	if err := tx.Commit(); err != nil {
		fmt.Printf("Error committing transaction: %v\n", err)
		return
	}

	var sentCount int

	// Now send messages outside of transaction
	for _, reminder := range reminders {
		// Parse the chat ID to JID
		jid, err := types.ParseJID(reminder.ChatID)
		if err != nil {
			fmt.Printf("Error parsing chat ID %s: %v\n", reminder.ChatID, err)
			// Delete invalid reminder
			db.Exec(`DELETE FROM reminders WHERE id = ?`, reminder.ID)
			continue
		}

		// Format the message based on type
		var messageText string
		if reminder.ReminderType == "send" {
			messageText = reminder.Message
		} else {
			messageText = fmt.Sprintf("⏰ Reminder: %s", reminder.Message)
		}

		// Send the message
		_, err = client.SendMessage(ctx, jid, &waE2E.Message{
			Conversation: &messageText,
		})

		if err != nil {
			fmt.Printf("Error sending reminder to %s: %v\n", reminder.ChatID, err)
			continue
		}

		// Mark the reminder as sent and snoozable (for "remind" type only)
		// "send" type messages get deleted immediately
		// Retry up to 3 times if database is locked
		updated := false
		for attempt := 0; attempt < 3; attempt++ {
			if reminder.ReminderType == "send" {
				// Delete send-type reminders after sending
				_, err = db.Exec(`DELETE FROM reminders WHERE id = ?`, reminder.ID)
			} else {
				// Mark remind-type reminders as sent and snoozable
				_, err = db.Exec(`
					UPDATE reminders
					SET sent_time = ?, snoozable = 1
					WHERE id = ?
				`, now, reminder.ID)
			}

			if err == nil {
				updated = true
				break
			}
			if attempt < 2 {
				time.Sleep(100 * time.Millisecond)
			}
		}

		if !updated {
			fmt.Printf("Error updating reminder %s after 3 attempts: %v\n", reminder.ID, err)
		} else {
			sentCount++
			fmt.Printf("Sent %s reminder to %s: %s\n", reminder.ReminderType, reminder.ChatID, reminder.Message)
		}
	}

	if sentCount > 0 {
		fmt.Printf("Sent %d reminder(s) at %s\n", sentCount, time.Now().Format("2006-01-02 15:04:05"))
	}
}

// cleanupOldReminders deletes reminders that were sent more than 24 hours ago
func cleanupOldReminders(db *sql.DB) {
	cutoffTime := time.Now().Unix() - (24 * 60 * 60) // 24 hours ago

	result, err := db.Exec(`
		DELETE FROM reminders
		WHERE sent_time > 0 AND sent_time < ?
	`, cutoffTime)

	if err != nil {
		fmt.Printf("Error cleaning up old reminders: %v\n", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("Cleaned up %d old reminder(s) at %s\n", rowsAffected, time.Now().Format("2006-01-02 15:04:05"))
	}
}

func eighteenNineteen(client *whatsmeow.Client, ctx context.Context) bool {
	haifaTZ, _ := time.LoadLocation("Asia/Jerusalem")
	now := time.Now().In(haifaTZ)
	hour := now.Hour()
	minute := now.Minute()
	seconds := now.Second()

	jid, err := types.ParseJID("120363154021870149@g.us")
	message := "It's that time of day again if we're still doing this"

	if err != nil {
		fmt.Printf("Error parsing chat ID: %v\n", err)
		return false
	}

	if hour == 18 && minute == 19 && seconds < 5 {
		// Send message to all chats
		client.SendMessage(ctx, jid, &waE2E.Message{
			Conversation: &message,
		})
	}

	return true
}
