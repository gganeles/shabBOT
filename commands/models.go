package commands

import (
	"database/sql"
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

// InitDB initializes the database schema
func InitDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS chats (
		chat_id TEXT PRIMARY KEY,
		location TEXT DEFAULT 'Haifa'
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
	);`

	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	// Migration: Add sent_time column if it doesn't exist
	// This handles existing databases that don't have the column
	_, err = db.Exec(`
		ALTER TABLE reminders ADD COLUMN sent_time INTEGER DEFAULT 0
	`)
	// Ignore error if column already exists
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		// Only return error if it's not a "duplicate column" error
		return nil // SQLite returns "duplicate column name" for existing columns
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

// SetChatLocation sets the location for a chat
func SetChatLocation(db *sql.DB, chatID, location string) error {
	GetOrCreateChat(db, chatID)
	_, err := db.Exec(`UPDATE chats SET location = ? WHERE chat_id = ?`, location, chatID)
	return err
}

// Utility function to capitalize first letter
func CapitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
