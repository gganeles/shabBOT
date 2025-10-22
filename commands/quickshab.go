package commands

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var typeList = []string{
	"main",
	"side",
	"plastics",
	"drinks",
	"wine",
	"challah",
	"dips",
	"dessert",
}

// needsNumber calculates how many of each category is needed based on number of people
func needsNumber(n int) map[string]int {
	return map[string]int{
		"main": int(math.Ceil(float64(n+1) / 4)),
		"side": int(math.Ceil(float64(n) / 6)),
		"plastics": func() int {
			if n > 5 {
				return int(math.Ceil(float64(n) / 20))
			}
			return 0
		}(),
		"drinks": func() int {
			if n > 5 {
				return int(math.Ceil(float64(n) / 12))
			}
			return 0
		}(),
		"wine":    int(math.Ceil(float64(n) / 8)),
		"challah": int(math.Ceil(float64(n) / 10)),
		"dips":    int(math.Ceil(float64(n+1) / 12)),
		"dessert": int(math.Ceil(float64(n-5) / 8)),
	}
}

// QuickShabCmd initializes a quickshab meal planning tracker
func QuickShabCmd(db *sql.DB, prompt, chatID string) string {
	args := parseArgs(prompt)

	if len(args) < 2 {
		return "Command syntax: !new [number of people]"
	}

	numPeople, err := strconv.Atoi(args[1])
	if err != nil || numPeople <= 0 {
		return "Command syntax: !new [number of people]"
	}

	// Ensure chat exists
	GetOrCreateChat(db, chatID)

	// Clear existing quickshab data
	db.Exec(`DELETE FROM quickshab_assignments WHERE chat_id = ?`, chatID)

	// Create or update quickshab
	_, err = db.Exec(`
		INSERT INTO quickshab (chat_id, number) VALUES (?, ?)
		ON CONFLICT(chat_id) DO UPDATE SET number = ?
	`, chatID, numPeople, numPeople)

	if err != nil {
		return "Error creating quickshab"
	}

	return formatQuickShab(db, chatID)
}

// UpdateNumberCmd updates the number of people for quickshab
func UpdateNumberCmd(db *sql.DB, prompt, chatID string) string {
	args := parseArgs(prompt)

	if len(args) < 2 {
		return "Command syntax: !update [number of people]"
	}

	numPeople, err := strconv.Atoi(args[1])
	if err != nil || numPeople <= 0 {
		return "Command syntax: !update [number of people]"
	}

	// Check if quickshab exists
	var exists int
	err = db.QueryRow(`SELECT COUNT(*) FROM quickshab WHERE chat_id = ?`, chatID).Scan(&exists)
	if err != nil || exists == 0 {
		return "You need to create a quickshab first by typing !quickshab [number of people]"
	}

	// Update number
	_, err = db.Exec(`UPDATE quickshab SET number = ? WHERE chat_id = ?`, numPeople, chatID)
	if err != nil {
		return "Error updating quickshab"
	}

	return formatQuickShab(db, chatID)
}

// formatQuickShab formats the quickshab display
func formatQuickShab(db *sql.DB, chatID string) string {
	// Get number of people
	var numPeople int
	err := db.QueryRow(`SELECT number FROM quickshab WHERE chat_id = ?`, chatID).Scan(&numPeople)
	if err != nil {
		return "No quickshab found"
	}

	needs := needsNumber(numPeople)

	// Get assignments
	rows, err := db.Query(`
		SELECT category, name FROM quickshab_assignments 
		WHERE chat_id = ? ORDER BY category, id
	`, chatID)
	if err != nil {
		return "Error retrieving assignments"
	}
	defer rows.Close()

	// Build assignments map
	assignments := make(map[string][]string)
	for rows.Next() {
		var category, name string
		rows.Scan(&category, &name)
		assignments[category] = append(assignments[category], name)
	}

	// Format output
	var result []string

	// Process standard categories
	for _, category := range typeList {
		needed := needs[category]
		assigned := assignments[category]
		upperLim := max(needed, len(assigned))

		for i := 0; i < upperLim; i++ {
			name := ""
			if i < len(assigned) {
				name = assigned[i]
			}
			result = append(result, fmt.Sprintf("%s: %s", CapitalizeFirst(category), name))
		}
	}

	// Process custom categories (not in typeList)
	for category, names := range assignments {
		if !contains(typeList, category) {
			for _, name := range names {
				result = append(result, fmt.Sprintf("%s: %s", CapitalizeFirst(category), name))
			}
		}
	}

	return strings.Join(result, "\n") + "\n\nExample Usage:\n   • !bring main\n   • !assign gabe main\n\n" +
		"You can also use !unbring and !unassign"
}

// ShowCmd displays current quickshab assignments
func ShowCmd(db *sql.DB, chatID string) string {
	// Check if quickshab exists
	var exists int
	err := db.QueryRow(`SELECT COUNT(*) FROM quickshab WHERE chat_id = ?`, chatID).Scan(&exists)
	if err != nil || exists == 0 {
		return "No quickshab found. Create one with !quickshab [number]"
	}

	return formatQuickShab(db, chatID)
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
