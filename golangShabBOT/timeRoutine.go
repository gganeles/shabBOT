package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

func timeRoutine(client *whatsmeow.Client, db *sql.DB, ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check for reminders that need to be sent
			checkAndSendReminders(client, db, ctx)
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
	rows, err := tx.Query(`
		SELECT id, chat_id, message, type 
		FROM reminders 
		WHERE time > 0 AND time <= ? 
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

		// Delete the reminder after successfully sending
		// Retry up to 3 times if database is locked
		deleted := false
		for attempt := 0; attempt < 3; attempt++ {
			_, err = db.Exec(`DELETE FROM reminders WHERE id = ?`, reminder.ID)
			if err == nil {
				deleted = true
				break
			}
			if attempt < 2 {
				time.Sleep(100 * time.Millisecond)
			}
		}

		if !deleted {
			fmt.Printf("Error deleting reminder %s after 3 attempts: %v\n", reminder.ID, err)
		} else {
			sentCount++
			fmt.Printf("Sent %s reminder to %s: %s\n", reminder.ReminderType, reminder.ChatID, reminder.Message)
		}
	}

	if sentCount > 0 {
		fmt.Printf("Sent %d reminder(s) at %s\n", sentCount, time.Now().Format("2006-01-02 15:04:05"))
	}
}
