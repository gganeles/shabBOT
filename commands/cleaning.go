package commands

import (
	"database/sql"
	"fmt"
	"strings"
)

var WeeklyOffset int = 0

var Chores []string = []string{"Kitchen", "Living Room and Trash", "Floors"}

func FormatCleaningString(db *sql.DB, showProgress bool) string {
	// Only show cleaners with assigned chores (active this week)
	rows, err := db.Query(`SELECT name, chore, done FROM weekly_cleaning WHERE chore != '' ORDER BY name`)
	if err != nil {
		fmt.Printf("Error querying weekly_cleaning: %v\n", err)
		return "Error retrieving cleaning data"
	}
	defer rows.Close()

	var builder strings.Builder
	builder.Grow(200)

	if showProgress {
		builder.WriteString("*Chores Done So Far:*\n")
	} else {
		builder.WriteString("*New Chores This Week:*\n")
	}

	for rows.Next() {
		var name, chore string
		var done bool
		if err := rows.Scan(&name, &chore, &done); err != nil {
			continue
		}

		builder.WriteString(fmt.Sprintf("  %s: %s", name, chore))

		if showProgress {
			if done {
				builder.WriteString(" ✅")
			} else {
				builder.WriteString(" ❌")
			}
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

func AssignChores(db *sql.DB) {
	// First, clear all current chore assignments
	_, err := db.Exec(`UPDATE weekly_cleaning SET chore = '', done = 0`)
	if err != nil {
		fmt.Printf("Error clearing chore assignments: %v\n", err)
		return
	}

	// Determine if upstirs or downstairs week

	upstairsWeek := WeeklyOffset%2 == 0

	// Get cleaners
	rows, err := db.Query(`SELECT name FROM weekly_cleaning WHERE up = ? ORDER BY name`, upstairsWeek)
	if err != nil {
		fmt.Printf("Error querying cleaners: %v\n", err)
		return
	}
	defer rows.Close()

	var allCleaners []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		allCleaners = append(allCleaners, name)
	}
	rows.Close()

	// Rotate the weekly offset to determine which cleaners are active
	fmt.Printf("Assigning chores for week with offset %d (upstairsWeek=%v), and the length of the cleaner list is %d\n", WeeklyOffset, upstairsWeek, len(allCleaners))
	WeeklyOffset = (WeeklyOffset + 1) % 6

	// Select cleaners for this week (rotate through the list)
	numChores := len(Chores)
	if numChores > len(allCleaners) {
		numChores = len(allCleaners)
	}

	// Assign chores to the selected cleaners
	for i := 0; i < numChores; i++ {
		cleanerIndex := (WeeklyOffset + i) % len(allCleaners)
		cleaner := allCleaners[cleanerIndex]
		chore := Chores[i]

		_, err := db.Exec(`UPDATE weekly_cleaning SET chore = ?, done = 0 WHERE name = ?`, chore, cleaner)
		if err != nil {
			fmt.Printf("Error assigning chore to %s: %v\n", cleaner, err)
		}
	}
}

// HandleCleaningCommand handles marking chores as done
func HandleCleaningCommand(db *sql.DB, chatID, sender, args string) string {
	// if the sender is me, allow reassignment

	if sender == "972587120601" && args == "redo" {
		// reassign all chores
		AssignChores(db)
		return FormatCleaningString(db, false)
	}

	// Check if this is the right chat
	if chatID != "120363037678094725@g.us" {
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
		// Show cleaning status - only show active cleaners with assigned chores
		rows, err := db.Query(`SELECT name, chore, done FROM weekly_cleaning WHERE chore != '' ORDER BY name`)
		if err != nil {
			return "Error retrieving cleaning status."
		}
		defer rows.Close()

		var builder strings.Builder
		builder.WriteString("*Current Cleaning Status:*\n")

		hasAny := false
		for rows.Next() {
			hasAny = true
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

		if !hasAny {
			return "No chores assigned yet. Chores will be assigned on Sunday at 9:00 AM."
		}

		return builder.String()
	}

	return "Usage: !clean [done/undone/status]"
}
