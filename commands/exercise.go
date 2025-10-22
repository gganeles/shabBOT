package commands

import (
	"database/sql"
	"regexp"
	"strconv"
	"strings"
)

func ExerciseCMD(db *sql.DB, prompt string, chatID string, attendeeName string) string {
	r := regexp.MustCompile(`\b\d+\b`)
	numberStr := r.FindString(prompt)

	if len(numberStr) > 0 {
		number, _ := strconv.Atoi(numberStr)

		GetOrCreateChat(db, chatID)

		_, err := db.Exec(`
				INSERT INTO exercise_counter (chat_id, user, value) VALUES (?, ?, ?)
				ON CONFLICT(chat_id) DO UPDATE SET value = value + ?
			`, chatID, attendeeName, number, number)

		if err != nil {
			return "Error ticking exercise counter: " + err.Error()
		}

		return numberStr + " minutes have been added to your time.\n\nCurrent Standings:\n" + GetExerciseResults(db, chatID)
	} else {
		return "Couldn't parse number of minutes"
	}
}

func GetExerciseResults(db *sql.DB, chatID string) string {
	rows, err := db.Query(`
		SELECT user, value FROM exercise_counter 
		WHERE chat_id = ?
	`, chatID)
	if err != nil {
		return "Error retrieving assignments"
	}
	defer rows.Close()

	// Build assignments map
	results := make(map[string]int)
	for rows.Next() {
		var user string
		var val int
		rows.Scan(&user, &val)
		results[user] = val
	}

	lines := make([]string, 0, 20)
	for name, val := range results {
		lines = append(lines, name+": "+strconv.Itoa(val))
	}

	return strings.Join(lines, "\n")
}
