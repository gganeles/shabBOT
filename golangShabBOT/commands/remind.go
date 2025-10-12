package commands

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// generateUUID generates a simple UUID-like string
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// RemindCmd creates timed or untimed reminders
func RemindCmd(db *sql.DB, prompt, chatID, location string) string {
	args := parseArgs(prompt)

	if len(args) < 2 {
		return "command usage: !remind   thing to remind   when"
	}

	// Remove the command word
	reminderText := strings.Join(args[1:], " ")

	// Try to parse a date/time from the reminder text
	// For now, we'll create a simplified version
	// In production, you'd use a date parsing library similar to chrono-node

	// Check if the reminder contains time indicators
	hasTime := containsTimeIndicator(reminderText)

	// Ensure chat exists
	GetOrCreateChat(db, chatID)

	if !hasTime {
		// Create untimed reminder
		message := cleanReminderText(reminderText)
		id := generateUUID()

		_, err := db.Exec(`
			INSERT INTO reminders (id, chat_id, message, time, type, snoozable) 
			VALUES (?, ?, ?, 0, 'timeless', 0)
		`, id, chatID, message)

		if err != nil {
			return fmt.Sprintf("Error creating reminder: %v", err)
		}

		return fmt.Sprintf(`Ok, "%s" will be added to your reminders`, message)
	}

	// Parse time (simplified - would use proper date parsing in production)
	targetTime, message, err := parseReminderTime(reminderText, location)
	if err != nil {
		// If parsing fails, create as untimed
		message = cleanReminderText(reminderText)
		id := generateUUID()

		_, err := db.Exec(`
			INSERT INTO reminders (id, chat_id, message, time, type, snoozable) 
			VALUES (?, ?, ?, 0, 'timeless', 0)
		`, id, chatID, message)

		if err != nil {
			return fmt.Sprintf("Error creating reminder: %v", err)
		}

		return fmt.Sprintf(`Ok, "%s" will be added to your reminders`, message)
	}

	// Create timed reminder
	id := generateUUID()
	_, err = db.Exec(`
		INSERT INTO reminders (id, chat_id, message, time, type, snoozable) 
		VALUES (?, ?, ?, ?, 'remind', 0)
	`, id, chatID, message, targetTime.Unix())

	if err != nil {
		return fmt.Sprintf("Error creating reminder: %v", err)
	}

	return fmt.Sprintf(`Ok, I'll remind you "%s" on %s`,
		message,
		targetTime.Format("Mon @ 3:04 PM"))
}

// RemindersCmd lists all active reminders
func RemindersCmd(db *sql.DB, chatID string) string {
	rows, err := db.Query(`
		SELECT id, message, time, type FROM reminders 
		WHERE chat_id = ? 
		ORDER BY time, id
	`, chatID)

	if err != nil {
		return fmt.Sprintf("Error retrieving reminders: %v", err)
	}
	defer rows.Close()

	var timedReminders []string
	var untimedReminders []string

	for rows.Next() {
		var id, message, reminderType string
		var reminderTime int64
		rows.Scan(&id, &message, &reminderTime, &reminderType)

		if reminderType == "timeless" || reminderTime == 0 {
			untimedReminders = append(untimedReminders, fmt.Sprintf("  • %s", message))
		} else {
			timeStr := time.Unix(reminderTime, 0).Format("Mon @ 3:04 PM")
			timedReminders = append(timedReminders, fmt.Sprintf("  • %s - %s", message, timeStr))
		}
	}

	if len(timedReminders) == 0 && len(untimedReminders) == 0 {
		return "You have no reminders."
	}

	result := ""
	if len(timedReminders) > 0 {
		result += "Timed Reminders:\n" + strings.Join(timedReminders, "\n")
	}
	if len(untimedReminders) > 0 {
		if result != "" {
			result += "\n\n"
		}
		result += "Untimed Reminders:\n" + strings.Join(untimedReminders, "\n")
	}

	return result
}

// SnoozeCmd snoozes reminders
func SnoozeCmd(db *sql.DB, prompt, chatID, location string) string {
	args := parseArgs(prompt)

	if len(args) < 2 {
		return "Command usage: !snooze [duration]"
	}

	// Get the most recent snoozable reminder
	var id, message string
	var reminderTime int64
	err := db.QueryRow(`
		SELECT id, message, time FROM reminders 
		WHERE chat_id = ? AND snoozable = 1 AND type = 'remind'
		ORDER BY time DESC 
		LIMIT 1
	`, chatID).Scan(&id, &message, &reminderTime)

	if err != nil {
		return "No snoozable reminders found"
	}

	// Parse snooze duration
	durationText := strings.Join(args[1:], " ")
	newTime, _, err := parseReminderTime("in "+durationText, location)

	if err != nil {
		return "Could not parse snooze duration. Try something like '10 minutes' or '2 hours'"
	}

	// Update reminder
	_, err = db.Exec(`
		UPDATE reminders 
		SET time = ?, snoozable = 0 
		WHERE id = ?
	`, newTime.Unix(), id)

	if err != nil {
		return fmt.Sprintf("Error snoozing reminder: %v", err)
	}

	return fmt.Sprintf("Alarm snoozed until %s", newTime.Format("Mon @ 3:04 PM"))
}

// DoneCmd marks reminders as complete
func DoneCmd(db *sql.DB, prompt, chatID string) string {
	// Try to find and remove the most recent reminder
	// This is simplified - the JS version has more complex logic

	var id string
	err := db.QueryRow(`
		SELECT id FROM reminders 
		WHERE chat_id = ? 
		ORDER BY time DESC, id DESC 
		LIMIT 1
	`, chatID).Scan(&id)

	if err != nil {
		return "No reminders found"
	}

	_, err = db.Exec(`DELETE FROM reminders WHERE id = ?`, id)

	if err != nil {
		return fmt.Sprintf("Error removing reminder: %v", err)
	}

	return "Reminder marked as done and removed"
}

// Helper functions

func containsTimeIndicator(text string) bool {
	timeWords := []string{"in", "at", "on", "tomorrow", "today", "tonight",
		"minute", "hour", "day", "week", "month", "am", "pm"}

	lowerText := strings.ToLower(text)
	for _, word := range timeWords {
		if strings.Contains(lowerText, word) {
			return true
		}
	}
	return false
}

func cleanReminderText(text string) string {
	// Remove common filler words
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "me to ")
	text = strings.TrimPrefix(text, "me ")
	text = strings.TrimSuffix(text, " in")
	text = strings.TrimSuffix(text, " at")
	return strings.TrimSpace(text)
}

func parseReminderTime(text, location string) (time.Time, string, error) {
	// Use the parseTime utility function as first resort
	// It handles natural language dates like "tomorrow at 3pm", "in 5 minutes", etc.

	// Try parsing the entire text first
	timestamp, err := parseTime(text, location)
	if err == nil {
		// Successfully parsed - now extract just the message part
		// Try to find common time prepositions to split the message
		message := extractMessageFromTimeText(text)
		return time.Unix(timestamp, 0), message, nil
	}

	// If direct parsing failed, try to find time expressions in the text
	// Look for patterns like "at [time]", "on [date]", "in [duration]"
	timePatterns := []string{" at ", " on ", " in ", " tomorrow ", " tonight ", " today "}

	for _, pattern := range timePatterns {
		idx := strings.Index(strings.ToLower(text), pattern)
		if idx != -1 {
			// Split at the time indicator
			messagePart := strings.TrimSpace(text[:idx])
			timePart := strings.TrimSpace(text[idx:])

			// Try to parse just the time part
			timestamp, err := parseTime(timePart, location)
			if err == nil {
				return time.Unix(timestamp, 0), messagePart, nil
			}
		}
	}

	return time.Time{}, "", fmt.Errorf("could not parse time")
}

func extractMessageFromTimeText(text string) string {
	// Remove common time-related phrases to extract the message
	message := text

	// Remove time indicators at the end
	timeIndicators := []string{
		" at ", " on ", " in ", " tomorrow", " tonight", " today",
		" am", " pm", " minutes", " hours", " days", " weeks",
	}

	lowerText := strings.ToLower(text)
	for _, indicator := range timeIndicators {
		if idx := strings.Index(lowerText, indicator); idx != -1 {
			// Keep only the part before the time indicator
			message = strings.TrimSpace(text[:idx])
			break
		}
	}

	// Clean up common prefixes
	message = strings.TrimPrefix(message, "me to ")
	message = strings.TrimPrefix(message, "me ")
	message = strings.TrimPrefix(message, "to ")

	return strings.TrimSpace(message)
}
