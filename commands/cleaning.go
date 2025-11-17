package commands

import (
	"database/sql"
	"fmt"
	"strings"
)

// HandleCleaningCommand handles marking chores as done
func HandleCleaningCommand(db *sql.DB, chatID, sender, args string) string {
	// Check if this is the right chat
	if chatID != "120363154021870149@g.us" {
		return "This command is only available in the household group chat."
	}

	// Get sender's name from database
	var senderName string
	err := db.QueryRow(`SELECT name FROM weekly_cleaning WHERE number = ?`, sender).Scan(&senderName)
	if err != nil {
		return "You are not registered in the cleaning system."
	}

	// Parse command
	args = strings.TrimSpace(strings.ToLower(args))

	if args == "done" || args == "complete" || args == "finished" {
		// Mark their chore as done
		result, err := db.Exec(`UPDATE weekly_cleaning SET done = 1 WHERE number = ?`, sender)
		if err != nil {
			return "Error updating your chore status."
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return "No chore found for you this week."
		}

		// Get their chore name
		var chore string
		db.QueryRow(`SELECT chore FROM weekly_cleaning WHERE number = ?`, sender).Scan(&chore)

		return fmt.Sprintf("✅ Marked your chore (%s) as done!", chore)
	} else if args == "undone" || args == "incomplete" {
		// Mark their chore as not done
		result, err := db.Exec(`UPDATE weekly_cleaning SET done = 0 WHERE number = ?`, sender)
		if err != nil {
			return "Error updating your chore status."
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return "No chore found for you this week."
		}

		return "❌ Marked your chore as not done."
	} else if args == "status" || args == "" {
		// Show cleaning status
		rows, err := db.Query(`SELECT name, chore, done FROM weekly_cleaning ORDER BY name`)
		if err != nil {
			return "Error retrieving cleaning status."
		}
		defer rows.Close()

		var builder strings.Builder
		builder.WriteString("*Current Cleaning Status:*\n")

		for rows.Next() {
			var name, chore string
			var done bool
			if err := rows.Scan(&name, &chore, &done); err != nil {
				continue
			}

			builder.WriteString(fmt.Sprintf("  %s: %s ", name, chore))
			if done {
				builder.WriteString("✅")
			} else {
				builder.WriteString("❌")
			}
			builder.WriteString("\n")
		}

		return builder.String()
	}

	return "Usage: !clean [done/undone/status]"
}
