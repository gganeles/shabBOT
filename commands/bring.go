package commands

import (
	"database/sql"
	"strings"
)

// BringCmd allows users to sign up to bring items
func BringCmd(db *sql.DB, prompt, chatID, attendeeID string) string {
	args := parseArgs(prompt)
	
	if len(args) < 2 {
		return "Command syntax: !bring [category]"
	}

	// Check if quickshab exists
	var exists int
	err := db.QueryRow(`SELECT COUNT(*) FROM quickshab WHERE chat_id = ?`, chatID).Scan(&exists)
	if err != nil || exists == 0 {
		return "You need to create a quickshab first by typing !quickshab [number of people]"
	}

	category := strings.ToLower(args[1])
	
	// Add assignment
	_, err = db.Exec(`
		INSERT INTO quickshab_assignments (chat_id, category, name) 
		VALUES (?, ?, ?)
	`, chatID, category, attendeeID)
	
	if err != nil {
		return "Error adding assignment"
	}

	return formatQuickShab(db, chatID)
}

// AssignCmd assigns items to specific people
func AssignCmd(db *sql.DB, prompt, chatID string) string {
	args := parseArgs(prompt)
	
	if len(args) < 3 {
		return "Command syntax: !assign [name] [category]"
	}

	// Check if quickshab exists
	var exists int
	err := db.QueryRow(`SELECT COUNT(*) FROM quickshab WHERE chat_id = ?`, chatID).Scan(&exists)
	if err != nil || exists == 0 {
		return "You need to create a quickshab first by typing !quickshab [number of people]"
	}

	name := args[1]
	category := strings.ToLower(args[2])
	
	// Add assignment
	_, err = db.Exec(`
		INSERT INTO quickshab_assignments (chat_id, category, name) 
		VALUES (?, ?, ?)
	`, chatID, category, name)
	
	if err != nil {
		return "Error adding assignment"
	}

	return formatQuickShab(db, chatID)
}

// UnbringCmd removes user's assignment from quickshab
func UnbringCmd(db *sql.DB, chatID, attendeeID string) string {
	// Check if quickshab exists
	var exists int
	err := db.QueryRow(`SELECT COUNT(*) FROM quickshab WHERE chat_id = ?`, chatID).Scan(&exists)
	if err != nil || exists == 0 {
		return "You need to create a quickshab first by typing !quickshab [number of people]"
	}

	// Remove the user's assignment
	result, err := db.Exec(`
		DELETE FROM quickshab_assignments 
		WHERE chat_id = ? AND name = ?
		AND id = (
			SELECT id FROM quickshab_assignments 
			WHERE chat_id = ? AND name = ? 
			LIMIT 1
		)
	`, chatID, attendeeID, chatID, attendeeID)
	
	if err != nil {
		return "Error removing assignment"
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return "You don't have any assignments"
	}

	return formatQuickShab(db, chatID)
}

// UnassignCmd removes a specific person's assignment
func UnassignCmd(db *sql.DB, prompt, chatID string) string {
	args := parseArgs(prompt)
	
	if len(args) < 2 {
		return "Command syntax: !unassign [name]"
	}

	name := strings.Join(args[1:], " ")
	
	// Check if quickshab exists
	var exists int
	err := db.QueryRow(`SELECT COUNT(*) FROM quickshab WHERE chat_id = ?`, chatID).Scan(&exists)
	if err != nil || exists == 0 {
		return "You need to create a quickshab first by typing !quickshab [number of people]"
	}

	// Remove the assignment
	result, err := db.Exec(`
		DELETE FROM quickshab_assignments 
		WHERE chat_id = ? AND name = ?
		AND id = (
			SELECT id FROM quickshab_assignments 
			WHERE chat_id = ? AND name = ? 
			LIMIT 1
		)
	`, chatID, name, chatID, name)
	
	if err != nil {
		return "Error removing assignment"
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return name + " doesn't have any assignments"
	}

	return formatQuickShab(db, chatID)
}
