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
func RemindCmd(db *sql.DB, prompt, chatID string, timezone *time.Location) string {
	args := parseArgs(prompt)

	if len(args) < 2 {
		return "command usage: !remind   thing to remind   when"
	}

	// Remove the command word
	reminderCMDText := strings.Join(args[1:], " ")

	// Ensure chat exists
	GetOrCreateChat(db, chatID)

	var response []string

	for reminderText := range strings.SplitSeq(reminderCMDText, ",") {
		// Parse time
		targetTime, message, err := parseReminderTime(reminderText, timezone)
		if err != nil {
			// If parsing fails, create as untimed
			message = cleanReminderText(reminderText)
			id := generateUUID()

			_, err := db.Exec(`
				INSERT INTO reminders (id, chat_id, message, time, type, snoozable) 
				VALUES (?, ?, ?, 0, 'timeless', 0)
			`, id, chatID, message)

			if err != nil {
				response = append(response, fmt.Sprintf("Error creating reminder: %v", err))
			} else {
				response = append(response, fmt.Sprintf(`Ok, "%s" will be added to your reminders`, message))
			}
			continue
		}

		// Create timed reminder
		id := generateUUID()
		_, err = db.Exec(`
			INSERT INTO reminders (id, chat_id, message, time, type, snoozable) 
			VALUES (?, ?, ?, ?, 'remind', 0)
		`, id, chatID, message, targetTime.Unix())

		if err != nil {
			response = append(response, fmt.Sprintf("Error creating reminder: %v", err))
		} else {
			response = append(response, fmt.Sprintf(`Ok, I'll remind you "%s" on %s`,
				message,
				targetTime.Format("Mon @ 3:04 PM")))
		}
	}
	return strings.Join(response, "\n")
}

// RemindersCmd lists all active reminders
func RemindersCmd(db *sql.DB, chatID string, timezone *time.Location) string {
	// Fetch timed reminders ordered by time (earliest first)
	timedRows, err := db.Query(`
		SELECT id, message, time, type FROM reminders 
		WHERE chat_id = ? AND type != 'timeless' AND time > 0
		ORDER BY time ASC
	`, chatID)

	if err != nil {
		return fmt.Sprintf("Error retrieving reminders: %v", err)
	}
	defer timedRows.Close()

	var timedReminders []string
	now := time.Now().Unix()
	nowTime := time.Now().In(timezone)

	for timedRows.Next() {
		var id, message, reminderType string
		var reminderTime int64
		timedRows.Scan(&id, &message, &reminderTime, &reminderType)

		reminderTimeObj := time.Unix(reminderTime, 0).In(timezone)
		timeStr := reminderTimeObj.Format("Mon @ 3:04 PM")
		relativeStr := formatRelativeTime(reminderTimeObj, nowTime)

		if reminderTime < now {
			timedReminders = append(timedReminders, fmt.Sprintf("  • %s - %s - past", message, timeStr))
		} else {
			if relativeStr != "" {
				timedReminders = append(timedReminders, fmt.Sprintf("  • %s - %s (%s)", message, timeStr, relativeStr))
			} else {
				timedReminders = append(timedReminders, fmt.Sprintf("  • %s - %s", message, timeStr))
			}
		}
	}

	// Fetch untimed reminders ordered by ROWID (insertion order, oldest first)
	untimedRows, err := db.Query(`
		SELECT id, message FROM reminders 
		WHERE chat_id = ? AND (type = 'timeless' OR time = 0)
		ORDER BY ROWID ASC
	`, chatID)

	if err != nil {
		return fmt.Sprintf("Error retrieving untimed reminders: %v", err)
	}
	defer untimedRows.Close()

	var untimedReminders []string

	for untimedRows.Next() {
		var id, message string
		untimedRows.Scan(&id, &message)
		untimedReminders = append(untimedReminders, fmt.Sprintf("  • %s", message))
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
		result += "Todo:\n" + strings.Join(untimedReminders, "\n")
	}

	return result
}

// SnoozeCmd snoozes reminders by keyword
func SnoozeCmd(db *sql.DB, prompt, chatID string, timezone *time.Location) string {
	args := parseArgs(prompt)

	if len(args) < 2 {
		return "Command usage: !snooze [duration] [optional reminder query]"
	}

	// Parse snooze duration from the arguments
	newTime, searchQuery, err := parseReminderTime(strings.Join(args[1:], " "), timezone)

	fmt.Printf("Search Query: |%s|\n", searchQuery)

	if err != nil {
		return "Could not parse snooze duration. Try something like '10 minutes' or '2 hours'"
	}

	// Query all reminders for this chat (not just snoozable ones)
	rows, err := db.Query(`
		SELECT id, message, time FROM reminders 
		WHERE chat_id = ? AND type = 'remind'
		ORDER BY time ASC
	`, chatID)

	if err != nil {
		return fmt.Sprintf("Error querying reminders: %v", err)
	}
	defer rows.Close()

	var reminderIDs []string

	// Build searchObjects for all reminders so we can pick appropriate default
	searchObjects := []searchObject{}
	idTimes := map[string]int64{}

	for rows.Next() {
		var id, message string
		var reminderTime int64
		if err := rows.Scan(&id, &message, &reminderTime); err != nil {
			continue
		}

		searchObjects = append(searchObjects, searchObject{
			id:      id,
			message: message,
			index:   len(searchObjects),
		})
		idTimes[id] = reminderTime
	}

	if len(searchObjects) == 0 {
		return "No reminders found to snooze."
	}

	// If no search query provided, pick the most-recent past reminder (largest time < now).
	// If none are past, pick the earliest upcoming reminder.
	if strings.TrimSpace(searchQuery) == "" {
		nowUnix := time.Now().Unix()
		var chosenID string
		var maxPastTime int64 = -1 << 62
		var minFutureTime int64 = 1<<62 - 1
		var chosenFutureID string

		for id, t := range idTimes {
			if t < nowUnix {
				if t > maxPastTime {
					maxPastTime = t
					chosenID = id
				}
			} else {
				if t < minFutureTime {
					minFutureTime = t
					chosenFutureID = id
				}
			}
		}

		if chosenID != "" {
			reminderIDs = []string{chosenID}
		} else if chosenFutureID != "" {
			reminderIDs = []string{chosenFutureID}
		} else {
			return "No reminders found to snooze."
		}
	} else {
		// Find matching reminders using whatKeys
		matchingIDs := whatKeys(searchQuery, searchObjects)

		if len(matchingIDs) == 0 {
			// If search query was provided but no matches, return nothing
			if strings.TrimSpace(searchQuery) != "" {
				return "No matching reminders found to snooze."
			}
			// Default to earliest reminder (first in list)
			reminderIDs = []string{searchObjects[0].id}
		} else {
			reminderIDs = matchingIDs
		}
	}
	// Update all matching reminders
	var snoozedMessages []string
	var successCount int

	for _, reminderID := range reminderIDs {
		// Find the message for this ID
		var message string
		for _, obj := range searchObjects {
			if obj.id == reminderID {
				message = obj.message
				break
			}
		}

		// Update reminder with new time and reset sent_time and snoozable flag
		_, err = db.Exec(`
			UPDATE reminders 
			SET time = ?, sent_time = 0, snoozable = 0 
			WHERE id = ?
		`, newTime.Unix(), reminderID)

		if err != nil {
			continue
		}

		successCount++
		snoozedMessages = append(snoozedMessages, message)
	}

	if successCount == 0 {
		return "Error snoozing reminders"
	}

	// Format response based on number of reminders snoozed
	if successCount == 1 {
		return fmt.Sprintf(`Reminder "%s" snoozed until %s`, snoozedMessages[0], newTime.Format("Mon @ 3:04 PM"))
	}

	return fmt.Sprintf(`%d reminders snoozed until %s:
  • %s`, successCount, newTime.Format("Mon @ 3:04 PM"), strings.Join(snoozedMessages, "\n  • "))
}

// DoneCmd marks reminders as complete
func DoneCmd(db *sql.DB, prompt, chatID string) string {
	args := parseArgs(prompt)

	// Query all reminders for this chat (both timed and untimed)
	rows, err := db.Query(`
		SELECT id, message, time, type FROM reminders 
		WHERE chat_id = ? 
		ORDER BY time DESC, id DESC
	`, chatID)

	if err != nil {
		return fmt.Sprintf("Error querying reminders: %v", err)
	}
	defer rows.Close()

	searchObjects := []searchObject{}

	for rows.Next() {
		var id, message, reminderType string
		var reminderTime int64
		if err := rows.Scan(&id, &message, &reminderTime, &reminderType); err != nil {
			continue
		}

		searchObjects = append(searchObjects, searchObject{
			id:      id,
			message: message,
			index:   len(searchObjects),
		})
	}

	if len(searchObjects) == 0 {
		return "No reminders found"
	}

	// If user provided search terms, use them
	var reminderIDs []string
	if len(args) > 1 {
		// User provided search query
		searchQuery := strings.Join(args[1:], " ")
		matchingIDs := whatKeys(searchQuery, searchObjects)

		if len(matchingIDs) == 0 {
			return "Could not find matching reminder(s). Try: !done or !done [reminder text]"
		}
		reminderIDs = matchingIDs
	} else {
		// Default to most recent reminder
		reminderIDs = []string{searchObjects[0].id}
	}

	// Delete all matching reminders
	var deletedMessages []string
	var successCount int

	for _, reminderID := range reminderIDs {
		// Find the message for this ID
		var message string
		for _, obj := range searchObjects {
			if obj.id == reminderID {
				message = obj.message
				break
			}
		}

		// Delete the reminder
		_, err = db.Exec(`DELETE FROM reminders WHERE id = ?`, reminderID)

		if err != nil {
			continue
		}

		successCount++
		deletedMessages = append(deletedMessages, message)
	}

	if successCount == 0 {
		return "Error removing reminders"
	}

	// Format response based on number of reminders deleted
	if successCount == 1 {
		return fmt.Sprintf(`Reminder "%s" marked as done and removed`, deletedMessages[0])
	}

	return fmt.Sprintf(`%d reminders marked as done and removed:
  • %s`, successCount, strings.Join(deletedMessages, "\n  • "))
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

func formatRelativeTime(targetTime time.Time, now time.Time) string {
	// Calculate difference
	diff := targetTime.Sub(now)

	// For past times
	if diff < 0 {
		return ""
	}

	// Same day
	if targetTime.Year() == now.Year() && targetTime.YearDay() == now.YearDay() {
		return "today"
	}

	// Tomorrow
	tomorrow := now.AddDate(0, 0, 1)
	if targetTime.Year() == tomorrow.Year() && targetTime.YearDay() == tomorrow.YearDay() {
		return "tomorrow"
	}

	// Within the next week (2-7 days)
	days := int(diff.Hours() / 24)
	if days >= 2 && days <= 7 {
		return fmt.Sprintf("in %d days", days)
	}

	// Next week (8-14 days)
	if days >= 8 && days <= 14 {
		return "next week"
	}

	// Weeks (15-60 days)
	if days >= 15 && days <= 60 {
		weeks := (days + 3) / 7 // Round to nearest week
		if weeks == 2 {
			return "in 2 weeks"
		}
		return fmt.Sprintf("in %d weeks", weeks)
	}

	// Months
	if days >= 61 && days <= 365 {
		months := (days + 15) / 30 // Rough month approximation
		if months == 1 {
			return "next month"
		}
		return fmt.Sprintf("in %d months", months)
	}

	// Years
	if days > 365 {
		years := (days + 180) / 365
		if years == 1 {
			return "next year"
		}
		return fmt.Sprintf("in %d years", years)
	}

	return ""
}

func parseReminderTime(text string, timezone *time.Location) (time.Time, string, error) {
	// Use the parseTime utility function as first resort
	// It handles natural language dates like "tomorrow at 3pm", "in 5 minutes", etc.
	timestamp, message, err := parseTime(text, timezone)
	if err == nil {
		// Successfully parsed - now extract just the message part
		// Try to find common time prepositions to split the message
		// Clean up common prefixes
		message = cleanReminderText(message)
		return time.Unix(timestamp, 0).In(timezone), message, nil
	}

	return time.Time{}, "", fmt.Errorf("could not parse time")
}
