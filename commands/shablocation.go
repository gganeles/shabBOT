package commands

import (
	"database/sql"
	"fmt"
	"strings"
)

// ShabLocationCmd sets or gets the chat's Shabbat location
func ShabLocationCmd(db *sql.DB, prompt, chatID string) string {
	args := parseArgs(prompt)

	// If no location provided, show current location
	if len(args) < 2 {
		location, err := GetChatLocation(db, chatID)
		if err != nil {
			return "Error retrieving location"
		}
		return fmt.Sprintf("Current location: %s", location)
	}

	// Extract location from prompt (everything after the command)
	location := strings.Join(args[1:], " ")
	location = strings.TrimSpace(location)

	// Validate location exists (simplified - in production, check against city database)
	if location == "" {
		return "Please provide a valid location"
	}

	// Set the location
	name, err := SetChatLocation(db, chatID, location)
	if err != nil {
		return "Error setting location"
	}

	return fmt.Sprintf("Chat location was set to %s", name)
}
