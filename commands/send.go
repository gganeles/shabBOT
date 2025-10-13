package commands

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SendCmd schedules a message to be sent at a specific time
func SendCmd(db *sql.DB, prompt, chatID string, timezone *time.Location) string {
	args := parseArgs(prompt)

	if len(args) < 2 {
		return "Command usage: !send [message] [time]"
	}

	messageText := strings.Join(args[1:], " ")

	// Parse time from message
	targetTime, cleanedMessage, err := parseReminderTime(messageText, timezone)
	if err != nil {
		return "Could not parse time. Try: !send [message] at [time]"
	}

	// Ensure chat exists
	GetOrCreateChat(db, chatID)

	// Create scheduled message
	id := generateUUID()
	_, err = db.Exec(`
		INSERT INTO reminders (id, chat_id, message, time, type, snoozable) 
		VALUES (?, ?, ?, ?, 'send', 0)
	`, id, chatID, cleanedMessage, targetTime.Unix())

	if err != nil {
		return "Error scheduling message"
	}

	return fmt.Sprintf(`Message "%s" scheduled for %s`,
		cleanedMessage,
		targetTime.Format("Mon @ 3:04 PM"))
}

// UnsendCmd cancels the most recent scheduled message
func UnsendCmd(db *sql.DB, chatID string) string {
	// Get the most recent scheduled message
	var id, message string
	err := db.QueryRow(`
		SELECT id, message FROM reminders 
		WHERE chat_id = ? AND type = 'send'
		ORDER BY time DESC 
		LIMIT 1
	`, chatID).Scan(&id, &message)

	if err != nil {
		return "No scheduled messages found"
	}

	// Delete it
	_, err = db.Exec(`DELETE FROM reminders WHERE id = ?`, id)

	if err != nil {
		return "Error canceling scheduled message"
	}

	return fmt.Sprintf(`Canceled scheduled message: "%s"`, message)
}

// ScheduledCmd lists all scheduled messages
func ScheduledCmd(db *sql.DB, chatID string) string {
	rows, err := db.Query(`
		SELECT message, time FROM reminders 
		WHERE chat_id = ? AND type = 'send'
		ORDER BY time
	`, chatID)

	if err != nil {
		return "Error retrieving scheduled messages"
	}
	defer rows.Close()

	var messages []string

	for rows.Next() {
		var message string
		var sendTime int64
		rows.Scan(&message, &sendTime)

		timeStr := time.Unix(sendTime, 0).Format("Mon @ 3:04 PM")
		messages = append(messages, fmt.Sprintf("  • %s - %s", message, timeStr))
	}

	if len(messages) == 0 {
		return "No scheduled messages."
	}

	return "Scheduled Messages:\n" + strings.Join(messages, "\n")
}
