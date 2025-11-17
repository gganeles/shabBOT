package commands

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// ChatData represents the data structure for a chat
type ChatData struct {
	ChatID       string
	Location     string
	QuickShab    *QuickShab
	ShoppingList []ShoppingItem
}

// QuickShab represents meal planning data
type QuickShab struct {
	Number     int
	Categories map[string][]string // category -> list of names
}

// ShoppingItem represents an item in the shopping list
type ShoppingItem struct {
	Item     string
	Quantity int
}

// Reminder represents a reminder (timed or untimed)
type Reminder struct {
	ID        string
	Message   string
	Time      int64 // Unix timestamp, 0 for untimed
	ChatID    string
	Type      string // "remind" or "send"
	Snoozable bool
}

type Cleaner struct {
	Name   string
	Number string
	Done   string
	Chore  string
}

// InitDB initializes the database schema
func InitDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS chats (
		chat_id TEXT PRIMARY KEY,
		location TEXT DEFAULT 'Haifa',
		timezone TEXT DEFAULT 'Asia/Jerusalem'
	);

	CREATE TABLE IF NOT EXISTS quickshab (
		chat_id TEXT PRIMARY KEY,
		number INTEGER DEFAULT 0,
		FOREIGN KEY (chat_id) REFERENCES chats(chat_id)
	);

	CREATE TABLE IF NOT EXISTS quickshab_assignments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id TEXT NOT NULL,
		category TEXT NOT NULL,
		name TEXT NOT NULL,
		FOREIGN KEY (chat_id) REFERENCES chats(chat_id)
	);

	CREATE TABLE IF NOT EXISTS shopping_list (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id TEXT NOT NULL,
		item TEXT NOT NULL,
		quantity INTEGER DEFAULT 0,
		FOREIGN KEY (chat_id) REFERENCES chats(chat_id)
	);

	CREATE TABLE IF NOT EXISTS reminders (
		id TEXT PRIMARY KEY,
		chat_id TEXT NOT NULL,
		message TEXT NOT NULL,
		time INTEGER DEFAULT 0,
		type TEXT NOT NULL,
		snoozable INTEGER DEFAULT 0,
		sent_time INTEGER DEFAULT 0,
		FOREIGN KEY (chat_id) REFERENCES chats(chat_id)
	);

	CREATE TABLE IF NOT EXISTS exercise_counter (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id TEXT NOT NULL,
		user TEXT NOT NULL,
		value INTEGER DEFAULT 0,
		FOREIGN KEY (chat_id) REFERENCES chats(chat_id),
		UNIQUE(chat_id, user)
	);

	CREATE TABLE IF NOT EXISTS weekly_cleaning (
		number VARCHAR PRIMARY KEY,
		name VARCHAR NOT NULL,
		chore VARCHAR DEFAULT '',
		done BOOLEAN DEFAULT FALSE
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	// Migration: Remove duplicate cleaners from weekly_cleaning table
	// This handles existing databases that have duplicates from before the PRIMARY KEY was added
	// Keep only the first occurrence of each number (phone number)
	_, err = db.Exec(`
		DELETE FROM weekly_cleaning 
		WHERE rowid NOT IN (
			SELECT MIN(rowid) 
			FROM weekly_cleaning 
			GROUP BY number
		)
	`)
	if err != nil {
		log.Printf("Warning: Failed to remove duplicate cleaners: %v", err)
		// Don't return error, just log it - this is a non-critical migration
	}

	cleaners := map[string]string{
		"972542254475": "Lidor",
		"972524673512": "Jeremy",
		"972587920084": "Liron",
		"972587120601": "Gabe",
		"972586350530": "Nico",
		"972586251000": "Luke",
	}

	for dude := range cleaners {
		db.Exec(`
			INSERT OR IGNORE INTO weekly_cleaning (number, name, chore, done)
		 	VALUES (?,?,?,?)
			`,
			dude, cleaners[dude], "", false)
	}

	// Migration: Add sent_time column if it doesn't exist
	// This handles existing databases that don't have the column
	_, err = db.Exec(`ALTER TABLE reminders ADD COLUMN sent_time INTEGER DEFAULT 0`)
	// Ignore error if column already exists
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		// Only return error if it's not a "duplicate column" error
		return nil // SQLite returns "duplicate column name" for existing columns
	}

	// Migration: Add timezone column to chats if it doesn't exist
	_, err = db.Exec(`ALTER TABLE chats ADD COLUMN timezone TEXT DEFAULT 'Asia/Jerusalem'`)
	// Ignore error if column already exists
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return nil
	}

	// ensure unique index exists for existing DBs (adds UNIQUE for older DBs)
	_, err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_exercise_chat_user ON exercise_counter(chat_id, user)`)
	if err != nil {
		return err
	}

	return nil
}

// GetOrCreateChat ensures a chat exists in the database
func GetOrCreateChat(db *sql.DB, chatID string) error {
	_, err := db.Exec(`INSERT OR IGNORE INTO chats (chat_id) VALUES (?)`, chatID)
	return err
}

// GetChatLocation retrieves the location for a chat
func GetChatLocation(db *sql.DB, chatID string) (string, error) {
	var location string
	err := db.QueryRow(`SELECT location FROM chats WHERE chat_id = ?`, chatID).Scan(&location)
	if err == sql.ErrNoRows {
		return "Haifa", nil
	}
	return location, err
}

// GetChatTimezone retrieves the timezone for a chat
func GetChatTimezone(db *sql.DB, chatID string) string {
	var timezone string
	err := db.QueryRow(`SELECT timezone FROM chats WHERE chat_id = ?`, chatID).Scan(&timezone)
	if err != nil {
		return "Asia/Jerusalem" // Default fallback
	}
	return timezone
}

// SetChatLocation sets the location for a chat and updates timezone based on geo data
func SetChatLocation(db *sql.DB, chatID, location string) (string, error) {
	GetOrCreateChat(db, chatID)

	// Try to find location in geoNamesList
	geoLocation := findGeoLocation(location)

	// log.Printf("[DEBUG] findGeoLocation returned: %v", geoLocation)

	// If location not found in geo database, return error
	if geoLocation == nil {
		log.Printf("[ERROR] Location '%s' not found in geo database", location)
		return "", fmt.Errorf("location '%s' not found. Please use a valid city name", location)
	}

	// Use timezone from geo data, or default if not available
	timezone := "Asia/Jerusalem"
	if geoLocation.Timezone != "" {
		timezone = geoLocation.Timezone
	}

	// log.Printf("[DEBUG] Using geo location: %s, timezone: %s", geoLocation.Name, timezone)

	// log.Printf("[DEBUG] Executing UPDATE query with location: %s, timezone: %s, chatID: %s", geoLocation.Name, timezone, chatID)
	_, err := db.Exec(`UPDATE chats SET location = ?, timezone = ? WHERE chat_id = ?`, geoLocation.Name, timezone, chatID)
	if err != nil {
		log.Printf("[ERROR] UPDATE query failed: %v", err)
		return "", err
	}
	return geoLocation.Name, nil
}

// Utility function to capitalize first letter
func CapitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
