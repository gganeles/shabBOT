package commands

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// ShabLocationCmd sets or gets the chat's Shabbat location
func ShabLocationCmd(db *sql.DB, prompt, chatID string) string {
	// log.Printf("[DEBUG] ShabLocationCmd called with prompt: '%s', chatID: '%s'", prompt, chatID)
	args := parseArgs(prompt)
	// log.Printf("[DEBUG] Parsed args: %v", args)

	// If no location provided, show current location
	if len(args) < 2 {
		// log.Printf("[DEBUG] No location provided, getting current location")
		location, err := GetChatLocation(db, chatID)
		if err != nil {
			log.Printf("[ERROR] GetChatLocation failed: %v", err)
			return "Error retrieving location"
		}
		// log.Printf("[DEBUG] Current location retrieved: %s", location)
		return fmt.Sprintf("Current location: %s", location)
	}

	// Extract location from prompt (everything after the command)
	location := strings.Join(args[1:], " ")
	location = strings.TrimSpace(location)
	// log.Printf("[DEBUG] Setting location to: '%s'", location)

	// Validate location exists (simplified - in production, check against city database)
	if location == "" {
		// log.Printf("[DEBUG] Location is empty after trimming")
		return "Please provide a valid location"
	}

	// Set the location
	// log.Printf("[DEBUG] Calling SetChatLocation with chatID: %s, location: %s", chatID, location)
	name, err := SetChatLocation(db, chatID, location)
	if err != nil {
		// Return the error message to the user
		return err.Error()
	}

	// log.Printf("[DEBUG] Location set successfully to: %s", name)
	return fmt.Sprintf("Chat location was set to %s", name)
}
