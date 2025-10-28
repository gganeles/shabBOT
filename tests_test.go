package main

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// MockTimeProvider allows us to control the current time in tests
type MockTimeProvider struct {
	CurrentTime time.Time
}

func (m MockTimeProvider) Now() time.Time {
	return m.CurrentTime
}

// TestCheckAndSendExerciseResults tests the checkAndSendExerciseResults function
func TestCheckAndSendExerciseResults(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer db.Close()

	// Create necessary tables
	_, err = db.Exec(`
		CREATE TABLE chats (
			chat_id TEXT PRIMARY KEY,
			timezone TEXT NOT NULL DEFAULT 'Asia/Jerusalem'
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create chats table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE exercise_counter (
			chat_id TEXT NOT NULL,
			user TEXT NOT NULL,
			value INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (chat_id, user)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create exercise_counter table: %v", err)
	}

	// Test case 1: No exercise data
	t.Run("NoExerciseData", func(t *testing.T) {
		// This should not panic or error when there's no data
		ctx := context.Background()
		checkAndSendExerciseResults(nil, db, ctx)
		// If we reach here without panic, the test passes
	})

	// Test case 2: Exercise data exists but it's not Sunday at 9:00 AM
	t.Run("NotSundayMorning", func(t *testing.T) {
		// Insert test data
		chatID := "1234567890@g.us"
		timezone := "America/New_York"

		_, err := db.Exec(`INSERT INTO chats (chat_id, timezone) VALUES (?, ?)`, chatID, timezone)
		if err != nil {
			t.Fatalf("Failed to insert chat: %v", err)
		}

		_, err = db.Exec(`INSERT INTO exercise_counter (chat_id, user, value) VALUES (?, ?, ?)`,
			chatID, "TestUser", 5)
		if err != nil {
			t.Fatalf("Failed to insert exercise counter: %v", err)
		}

		ctx := context.Background()
		// Call the function - it should skip sending since it's likely not Sunday 9 AM
		checkAndSendExerciseResults(nil, db, ctx)

		// Verify the counter was NOT reset (since message wasn't sent)
		var value int
		err = db.QueryRow(`SELECT value FROM exercise_counter WHERE chat_id = ? AND user = ?`,
			chatID, "TestUser").Scan(&value)
		if err != nil {
			t.Fatalf("Failed to query exercise counter: %v", err)
		}
		if value != 5 {
			t.Errorf("Expected value to remain 5, got %d", value)
		}

		// Cleanup
		db.Exec(`DELETE FROM exercise_counter WHERE chat_id = ?`, chatID)
		db.Exec(`DELETE FROM chats WHERE chat_id = ?`, chatID)
	})

	// Test case 3: Query execution with timezone handling
	t.Run("TimezoneHandling", func(t *testing.T) {
		// Test with multiple timezones
		testCases := []struct {
			chatID   string
			timezone string
		}{
			{"chat1@g.us", "America/Los_Angeles"},
			{"chat2@g.us", "Europe/London"},
			{"chat3@g.us", "Asia/Jerusalem"},
			{"chat4@g.us", "Australia/Sydney"},
		}

		for _, tc := range testCases {
			_, err := db.Exec(`INSERT INTO chats (chat_id, timezone) VALUES (?, ?)`,
				tc.chatID, tc.timezone)
			if err != nil {
				t.Fatalf("Failed to insert chat %s: %v", tc.chatID, err)
			}

			_, err = db.Exec(`INSERT INTO exercise_counter (chat_id, user, value) VALUES (?, ?, ?)`,
				tc.chatID, "User1", 3)
			if err != nil {
				t.Fatalf("Failed to insert exercise counter for %s: %v", tc.chatID, err)
			}
		}

		ctx := context.Background()
		// This should process all chats without errors
		checkAndSendExerciseResults(nil, db, ctx)

		// Verify all data still exists (not Sunday 9 AM)
		for _, tc := range testCases {
			var count int
			err := db.QueryRow(`SELECT COUNT(*) FROM exercise_counter WHERE chat_id = ?`,
				tc.chatID).Scan(&count)
			if err != nil {
				t.Fatalf("Failed to query for chat %s: %v", tc.chatID, err)
			}
			if count != 1 {
				t.Errorf("Expected 1 exercise record for %s, got %d", tc.chatID, count)
			}
		}

		// Cleanup
		for _, tc := range testCases {
			db.Exec(`DELETE FROM exercise_counter WHERE chat_id = ?`, tc.chatID)
			db.Exec(`DELETE FROM chats WHERE chat_id = ?`, tc.chatID)
		}
	})

	// Test case 4: Invalid timezone handling
	t.Run("InvalidTimezone", func(t *testing.T) {
		chatID := "invalid_tz@g.us"

		_, err := db.Exec(`INSERT INTO chats (chat_id, timezone) VALUES (?, ?)`,
			chatID, "Invalid/Timezone")
		if err != nil {
			t.Fatalf("Failed to insert chat: %v", err)
		}

		_, err = db.Exec(`INSERT INTO exercise_counter (chat_id, user, value) VALUES (?, ?, ?)`,
			chatID, "TestUser", 7)
		if err != nil {
			t.Fatalf("Failed to insert exercise counter: %v", err)
		}

		ctx := context.Background()
		// This should handle the error gracefully and continue
		checkAndSendExerciseResults(nil, db, ctx)

		// Cleanup
		db.Exec(`DELETE FROM exercise_counter WHERE chat_id = ?`, chatID)
		db.Exec(`DELETE FROM chats WHERE chat_id = ?`, chatID)
	})

	// Test case 5: Invalid chat JID
	t.Run("InvalidJID", func(t *testing.T) {
		chatID := "not-a-valid-jid"

		_, err := db.Exec(`INSERT INTO chats (chat_id, timezone) VALUES (?, ?)`,
			chatID, "America/New_York")
		if err != nil {
			t.Fatalf("Failed to insert chat: %v", err)
		}

		_, err = db.Exec(`INSERT INTO exercise_counter (chat_id, user, value) VALUES (?, ?, ?)`,
			chatID, "TestUser", 2)
		if err != nil {
			t.Fatalf("Failed to insert exercise counter: %v", err)
		}

		ctx := context.Background()
		// This should handle the error gracefully
		checkAndSendExerciseResults(nil, db, ctx)

		// Cleanup
		db.Exec(`DELETE FROM exercise_counter WHERE chat_id = ?`, chatID)
		db.Exec(`DELETE FROM chats WHERE chat_id = ?`, chatID)
	})
}

// TestCheckAndSendExerciseResultsOnSunday tests behavior specifically on Sunday at 9 AM
func TestCheckAndSendExerciseResultsOnSunday(t *testing.T) {
	t.Run("SundayAt9AM_ResetCounters", func(t *testing.T) {
		// Create an in-memory SQLite database for testing
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create in-memory database: %v", err)
		}
		defer db.Close()

		// Create necessary tables
		_, err = db.Exec(`
			CREATE TABLE chats (
				chat_id TEXT PRIMARY KEY,
				timezone TEXT NOT NULL DEFAULT 'Asia/Jerusalem'
			)
		`)
		if err != nil {
			t.Fatalf("Failed to create chats table: %v", err)
		}

		_, err = db.Exec(`
			CREATE TABLE exercise_counter (
				chat_id TEXT NOT NULL,
				user TEXT NOT NULL,
				value INTEGER NOT NULL DEFAULT 0,
				PRIMARY KEY (chat_id, user)
			)
		`)
		if err != nil {
			t.Fatalf("Failed to create exercise_counter table: %v", err)
		}

		chatID := "sunday_test@g.us"
		timezone := "America/New_York"

		// Insert test data
		_, err = db.Exec(`INSERT INTO chats (chat_id, timezone) VALUES (?, ?)`, chatID, timezone)
		if err != nil {
			t.Fatalf("Failed to insert chat: %v", err)
		}

		_, err = db.Exec(`INSERT INTO exercise_counter (chat_id, user, value) VALUES (?, ?, ?)`,
			chatID, "Alice", 5)
		if err != nil {
			t.Fatalf("Failed to insert exercise counter: %v", err)
		}

		_, err = db.Exec(`INSERT INTO exercise_counter (chat_id, user, value) VALUES (?, ?, ?)`,
			chatID, "Bob", 3)
		if err != nil {
			t.Fatalf("Failed to insert exercise counter: %v", err)
		}

		// Verify table exists and data is there before calling function
		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM exercise_counter WHERE chat_id = ?`, chatID).Scan(&count)
		if err != nil {
			t.Fatalf("Table doesn't exist or query failed: %v", err)
		}
		if count != 2 {
			t.Fatalf("Expected 2 records, got %d", count)
		}

		// Test that UPDATE works directly
		_, err = db.Exec(`UPDATE exercise_counter SET value = 999 WHERE chat_id = ? AND user = ?`, chatID, "Alice")
		if err != nil {
			t.Fatalf("Direct UPDATE failed: %v", err)
		}
		// Verify the update worked
		var testVal int
		err = db.QueryRow(`SELECT value FROM exercise_counter WHERE chat_id = ? AND user = ?`, chatID, "Alice").Scan(&testVal)
		if err != nil || testVal != 999 {
			t.Fatalf("UPDATE test failed: err=%v, val=%d", err, testVal)
		}
		// Reset it back
		db.Exec(`UPDATE exercise_counter SET value = 5 WHERE chat_id = ? AND user = ?`, chatID, "Alice")

		// Create a mock time for Sunday at 9:00:02 AM Eastern Time
		// December 29, 2024 is a Sunday
		loc, _ := time.LoadLocation("America/New_York")
		sundayAt9AM := time.Date(2024, 12, 29, 9, 0, 2, 0, loc)
		mockTime := MockTimeProvider{CurrentTime: sundayAt9AM}

		ctx := context.Background()
		// Call the function with mocked time (client can be nil for this test)
		checkAndSendExerciseResultsWithTime(nil, db, ctx, mockTime)

		// Verify the counters were reset to 0
		var aliceValue, bobValue int
		err = db.QueryRow(`SELECT value FROM exercise_counter WHERE chat_id = ? AND user = ?`,
			chatID, "Alice").Scan(&aliceValue)
		if err != nil {
			t.Fatalf("Failed to query Alice's counter: %v", err)
		}

		err = db.QueryRow(`SELECT value FROM exercise_counter WHERE chat_id = ? AND user = ?`,
			chatID, "Bob").Scan(&bobValue)
		if err != nil {
			t.Fatalf("Failed to query Bob's counter: %v", err)
		}

		if aliceValue != 0 {
			t.Errorf("Expected Alice's counter to be reset to 0, got %d", aliceValue)
		}
		if bobValue != 0 {
			t.Errorf("Expected Bob's counter to be reset to 0, got %d", bobValue)
		}
	})

	t.Run("SundayAt9_05AM_NoReset", func(t *testing.T) {
		// Create an in-memory SQLite database for testing
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create in-memory database: %v", err)
		}
		defer db.Close()

		// Create necessary tables
		_, err = db.Exec(`
			CREATE TABLE chats (
				chat_id TEXT PRIMARY KEY,
				timezone TEXT NOT NULL DEFAULT 'Asia/Jerusalem'
			)
		`)
		if err != nil {
			t.Fatalf("Failed to create chats table: %v", err)
		}

		_, err = db.Exec(`
			CREATE TABLE exercise_counter (
				chat_id TEXT NOT NULL,
				user TEXT NOT NULL,
				value INTEGER NOT NULL DEFAULT 0,
				PRIMARY KEY (chat_id, user)
			)
		`)
		if err != nil {
			t.Fatalf("Failed to create exercise_counter table: %v", err)
		}

		chatID := "sunday_late@g.us"
		timezone := "America/New_York"

		// Insert test data
		_, err = db.Exec(`INSERT INTO chats (chat_id, timezone) VALUES (?, ?)`, chatID, timezone)
		if err != nil {
			t.Fatalf("Failed to insert chat: %v", err)
		}

		_, err = db.Exec(`INSERT INTO exercise_counter (chat_id, user, value) VALUES (?, ?, ?)`,
			chatID, "Charlie", 7)
		if err != nil {
			t.Fatalf("Failed to insert exercise counter: %v", err)
		}

		// Create a mock time for Sunday at 9:00:05 AM (outside the 5-second window)
		loc, _ := time.LoadLocation("America/New_York")
		sundayAt9_05AM := time.Date(2024, 12, 29, 9, 0, 5, 0, loc)
		mockTime := MockTimeProvider{CurrentTime: sundayAt9_05AM}

		ctx := context.Background()
		checkAndSendExerciseResultsWithTime(nil, db, ctx, mockTime)

		// Verify the counter was NOT reset (outside the time window)
		var charlieValue int
		err = db.QueryRow(`SELECT value FROM exercise_counter WHERE chat_id = ? AND user = ?`,
			chatID, "Charlie").Scan(&charlieValue)
		if err != nil {
			t.Fatalf("Failed to query Charlie's counter: %v", err)
		}

		if charlieValue != 7 {
			t.Errorf("Expected Charlie's counter to remain 7 (outside time window), got %d", charlieValue)
		}
	})

	t.Run("MondayAt9AM_NoReset", func(t *testing.T) {
		// Create an in-memory SQLite database for testing
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create in-memory database: %v", err)
		}
		defer db.Close()

		// Create necessary tables
		_, err = db.Exec(`
			CREATE TABLE chats (
				chat_id TEXT PRIMARY KEY,
				timezone TEXT NOT NULL DEFAULT 'Asia/Jerusalem'
			)
		`)
		if err != nil {
			t.Fatalf("Failed to create chats table: %v", err)
		}

		_, err = db.Exec(`
			CREATE TABLE exercise_counter (
				chat_id TEXT NOT NULL,
				user TEXT NOT NULL,
				value INTEGER NOT NULL DEFAULT 0,
				PRIMARY KEY (chat_id, user)
			)
		`)
		if err != nil {
			t.Fatalf("Failed to create exercise_counter table: %v", err)
		}

		chatID := "monday_test@g.us"
		timezone := "America/New_York"

		// Insert test data
		_, err = db.Exec(`INSERT INTO chats (chat_id, timezone) VALUES (?, ?)`, chatID, timezone)
		if err != nil {
			t.Fatalf("Failed to insert chat: %v", err)
		}

		_, err = db.Exec(`INSERT INTO exercise_counter (chat_id, user, value) VALUES (?, ?, ?)`,
			chatID, "David", 8)
		if err != nil {
			t.Fatalf("Failed to insert exercise counter: %v", err)
		}

		// Create a mock time for Monday at 9:00:02 AM
		// December 30, 2024 is a Monday
		loc, _ := time.LoadLocation("America/New_York")
		mondayAt9AM := time.Date(2024, 12, 30, 9, 0, 2, 0, loc)
		mockTime := MockTimeProvider{CurrentTime: mondayAt9AM}

		ctx := context.Background()
		checkAndSendExerciseResultsWithTime(nil, db, ctx, mockTime)

		// Verify the counter was NOT reset (wrong day)
		var davidValue int
		err = db.QueryRow(`SELECT value FROM exercise_counter WHERE chat_id = ? AND user = ?`,
			chatID, "David").Scan(&davidValue)
		if err != nil {
			t.Fatalf("Failed to query David's counter: %v", err)
		}

		if davidValue != 8 {
			t.Errorf("Expected David's counter to remain 8 (not Sunday), got %d", davidValue)
		}
	})

	t.Run("SundayAt9AM_MultipleTimezones", func(t *testing.T) {
		// Create an in-memory SQLite database for testing
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create in-memory database: %v", err)
		}
		defer db.Close()

		// Create necessary tables
		_, err = db.Exec(`
			CREATE TABLE chats (
				chat_id TEXT PRIMARY KEY,
				timezone TEXT NOT NULL DEFAULT 'Asia/Jerusalem'
			)
		`)
		if err != nil {
			t.Fatalf("Failed to create chats table: %v", err)
		}

		_, err = db.Exec(`
			CREATE TABLE exercise_counter (
				chat_id TEXT NOT NULL,
				user TEXT NOT NULL,
				value INTEGER NOT NULL DEFAULT 0,
				PRIMARY KEY (chat_id, user)
			)
		`)
		if err != nil {
			t.Fatalf("Failed to create exercise_counter table: %v", err)
		}

		// Test that different timezones are handled correctly
		testCases := []struct {
			chatID   string
			timezone string
			mockTime time.Time
			expected int // expected value after function runs (0 if reset, original if not)
		}{
			// Sunday 9 AM in New York
			{"chat_ny@g.us", "America/New_York",
				time.Date(2024, 12, 29, 9, 0, 2, 0, mustLoadLocation("America/New_York")), 0},
			// Sunday 9 AM in London
			{"chat_london@g.us", "Europe/London",
				time.Date(2024, 12, 29, 9, 0, 2, 0, mustLoadLocation("Europe/London")), 0},
			// Sunday 8 AM in Sydney (not 9 AM)
			{"chat_sydney@g.us", "Australia/Sydney",
				time.Date(2024, 12, 29, 8, 0, 2, 0, mustLoadLocation("Australia/Sydney")), 10},
		}

		for _, tc := range testCases {
			// Insert test data
			_, err := db.Exec(`INSERT OR REPLACE INTO chats (chat_id, timezone) VALUES (?, ?)`,
				tc.chatID, tc.timezone)
			if err != nil {
				t.Fatalf("Failed to insert chat %s: %v", tc.chatID, err)
			}

			_, err = db.Exec(`INSERT OR REPLACE INTO exercise_counter (chat_id, user, value) VALUES (?, ?, ?)`,
				tc.chatID, "User1", 10)
			if err != nil {
				t.Fatalf("Failed to insert exercise counter for %s: %v", tc.chatID, err)
			}

			mockTime := MockTimeProvider{CurrentTime: tc.mockTime}
			ctx := context.Background()
			checkAndSendExerciseResultsWithTime(nil, db, ctx, mockTime)

			var value int
			err = db.QueryRow(`SELECT value FROM exercise_counter WHERE chat_id = ? AND user = ?`,
				tc.chatID, "User1").Scan(&value)
			if err != nil {
				t.Fatalf("Failed to query counter for %s: %v", tc.chatID, err)
			}

			if value != tc.expected {
				t.Errorf("Chat %s: expected value %d, got %d", tc.chatID, tc.expected, value)
			}
		}
	})
}

// Helper function for loading timezone
func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(fmt.Sprintf("Failed to load location %s: %v", name, err))
	}
	return loc
}

// BenchmarkCheckAndSendExerciseResults benchmarks the function performance
func BenchmarkCheckAndSendExerciseResults(b *testing.B) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer db.Close()

	// Setup tables
	db.Exec(`CREATE TABLE chats (chat_id TEXT PRIMARY KEY, timezone TEXT NOT NULL DEFAULT 'Asia/Jerusalem')`)
	db.Exec(`CREATE TABLE exercise_counter (chat_id TEXT NOT NULL, user TEXT NOT NULL, value INTEGER NOT NULL DEFAULT 0, PRIMARY KEY (chat_id, user))`)

	// Insert test data
	for i := 0; i < 10; i++ {
		chatID := time.Now().Format("20060102150405") + "@g.us"
		db.Exec(`INSERT INTO chats (chat_id, timezone) VALUES (?, ?)`, chatID, "Asia/Jerusalem")
		db.Exec(`INSERT INTO exercise_counter (chat_id, user, value) VALUES (?, ?, ?)`, chatID, "User1", 5)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkAndSendExerciseResults(nil, db, ctx)
	}
}
