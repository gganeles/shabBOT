package commands

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types" // For JID type
)

// CommandRouter routes incoming messages to appropriate command handlers
type CommandRouter struct {
	DB *sql.DB
}

// NewCommandRouter creates a new command router with database connection
func NewCommandRouter(db *sql.DB) *CommandRouter {
	return &CommandRouter{DB: db}
}

// Route processes a message and returns the appropriate response
// prompt: The full message text (including !)
// chatID: The WhatsApp chat/group ID
// attendeeID: The sender's WhatsApp ID/phone number
func (r *CommandRouter) Route(prompt string, evtInfo types.MessageInfo) string {
	// Check if message starts with !
	if !strings.HasPrefix(prompt, "!") {
		return ""
	}

	// Remove ! and get command
	prompt = strings.TrimPrefix(prompt, "!")
	args := parseArgs(prompt)

	if len(args) == 0 {
		return ""
	}

	cmd := strings.ToLower(args[0])

	chatID := evtInfo.Chat.String()
	attendeeName := evtInfo.PushName

	// Get location for commands that need it
	location, _ := GetChatLocation(r.DB, chatID)
	if location == "" {
		location = "Haifa"
	}

	timezone_name := GetChatTimezone(r.DB, chatID)
	if timezone_name == "" {
		timezone_name = "Asia/Jerusalem"
	}

	timezone, err := time.LoadLocation(timezone_name)
	if err != nil {
		return fmt.Sprintf("Error loading timezone: %v", err)
	}

	// Route to appropriate command handler
	switch cmd {
	// Simple static commands
	case "help":
		return HelpCmd(prompt)
	case "docs":
		return DocsCmd()
	case "count":
		return CountCmd()
	// Location commands
	case "shablocation", "shabloc", "location", "setloc", "chatlocation":
		return ShabLocationCmd(r.DB, prompt, chatID)
	case "shabtimes", "shabbattimes":
		return ShabTimesCmd(r.DB, prompt, chatID)
	case "chag", "chagtimes", "holiday", "holidays", "nextholiday":
		return ChagTimesCmd(r.DB, prompt, chatID)
	// QuickShab commands
	case "quickshab", "new":
		return QuickShabCmd(r.DB, prompt, chatID)

	case "update", "up", "num", "ppl":
		return UpdateNumberCmd(r.DB, prompt, chatID)

	case "show":
		return ShowCmd(r.DB, chatID)

	case "bring", "br", "bringing":
		return BringCmd(r.DB, prompt, chatID, attendeeName)

	case "assign":
		return AssignCmd(r.DB, prompt, chatID)

	case "unbring", "unbr":
		return UnbringCmd(r.DB, chatID, attendeeName)

	case "unassign":
		return UnassignCmd(r.DB, prompt, chatID)

	// Shopping list commands
	case "shop", "shp", "sh":
		return ShopCmd(r.DB, prompt, chatID)

	case "unshop", "unshp", "unsh", "ush":
		return UnshopCmd(r.DB, prompt, chatID)

	case "shoplist", "shplist", "shoppinglist", "shlst", "shlist", "shls":
		return ShopListCmd(r.DB, chatID)

	// Reminder commands
	case "remind", "r":
		return RemindCmd(r.DB, prompt, chatID, timezone)

	case "reminders", "rems", "todo", "rls":
		return RemindersCmd(r.DB, chatID, timezone)

	case "snooze":
		return SnoozeCmd(r.DB, prompt, chatID, timezone)

	case "done":
		return DoneCmd(r.DB, prompt, chatID)

	// Scheduled message commands
	case "send":
		return SendCmd(r.DB, prompt, chatID, timezone)

	case "unsend":
		return UnsendCmd(r.DB, chatID)

	case "scheduled":
		return ScheduledCmd(r.DB, chatID)

	// Other commands
	case "needs":
		return NeedsCmd(prompt)

	case "fast", "fasting":
		return FastCmd(r.DB, chatID)

	case "start":
		return StartCmd(prompt)

	case "end":
		return EndCmd()

	case "x", "exer", "exercise":
		return ExerciseCMD(r.DB, prompt, chatID, attendeeName)
	default:
		return "" // Unknown command, no response
	}
}

// ParseMultipleCommands handles messages with multiple commands separated by newlines
func (r *CommandRouter) ParseMultipleCommands(message string, evtInfo types.MessageInfo) []string {
	var responses []string

	// Check for multiple commands on same line (e.g., "!help !docs")
	// Split on any whitespace immediately before a '!' that is followed by a word character.
	// We want to keep the '!' at the start of the subsequent part, so split on the whitespace only.
	// Example matches: "\n!", " \t!", " \r\n!" where the char after ! is a word char.
	var parts []string
	// Use regex to find the split positions. We will split on the whitespace sequence that
	// directly precedes a '!' which itself is followed by a word character.
	re := regexp.MustCompile(`\s+!`)
	parts = re.Split(message, -1)
	for i, part := range parts {
		if i > 0 {
			part = "!" + part
		}

		response := r.Route(part, evtInfo)
		if response != "" {
			responses = append(responses, response)
		}
	}

	return responses
}
