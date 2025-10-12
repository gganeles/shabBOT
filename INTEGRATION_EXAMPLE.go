package gabba

// This is an example of how to integrate the commands package into your WhatsApp bot
// Place this in your main bot.go or create a separate message handler file

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"

	// Import your commands package
	"yourmodule/golangShabBOT/commands"
)

// Global database connection
var db *sql.DB
var router *commands.CommandRouter

// InitializeBot sets up the database and command router
func InitializeBot() error {
	var err error

	// Open database connection
	db, err = sql.Open("sqlite3", "file:shabbot.db?_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Initialize database schema
	if err := commands.InitDB(db); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create command router
	router = commands.NewCommandRouter(db)

	fmt.Println("Bot initialized successfully")
	return nil
}

// Enhanced event handler that processes commands
func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		handleMessage(v)
	}
}

// handleMessage processes incoming WhatsApp messages
func handleMessage(evt *events.Message) {
	// Skip if no message content
	if evt.Message == nil {
		return
	}

	// Get message text
	messageText := ""
	if evt.Message.Conversation != nil {
		messageText = *evt.Message.Conversation
	} else if evt.Message.ExtendedTextMessage != nil && evt.Message.ExtendedTextMessage.Text != nil {
		messageText = *evt.Message.ExtendedTextMessage.Text
	}

	// Skip if not a command
	if !strings.HasPrefix(messageText, "!") {
		return
	}

	// Get chat and sender info
	chatID := evt.Info.Chat.String()
	senderID := evt.Info.Sender.String()

	fmt.Printf("Received command from %s in %s: %s\n", senderID, chatID, messageText)

	// Process the command(s)
	responses := router.ParseMultipleCommands(messageText, chatID, senderID)

	// Send responses back
	for _, response := range responses {
		if response != "" {
			sendResponse(evt, response)
		}
	}
}

// sendResponse sends a text response back to the chat
func sendResponse(originalMsg *events.Message, responseText string) {
	// This is a simplified version - adjust based on your whatsmeow client setup
	// You'll need to pass your client instance here

	// Example (you'll need to adapt this to your actual client):
	// client.SendMessage(originalMsg.Info.Chat, &waProto.Message{
	//     Conversation: proto.String(responseText),
	// })

	fmt.Printf("Would send response: %s\n", responseText)
}

// Example of how to modify your main() function
func exampleMain() {
	// Initialize database and command router
	if err := InitializeBot(); err != nil {
		panic(err)
	}
	defer db.Close()

	// ... rest of your WhatsApp client setup ...
	// When adding event handler, use the enhanced eventHandler function:
	// client.AddEventHandler(eventHandler)
}

// Example of sending a response using the actual whatsmeow client
func sendResponseWithClient(client *whatsmeow.Client, chatID, messageText string) error {
	// Parse chat JID
	// chatJID, err := types.ParseJID(chatID)
	// if err != nil {
	//     return err
	// }

	// Send the message
	// _, err = client.SendMessage(context.Background(), chatJID, &waProto.Message{
	//     Conversation: proto.String(messageText),
	// })

	return nil
}

// Example of handling multiple commands in one message
func exampleMultipleCommands() {
	// Initialize
	InitializeBot()
	defer db.Close()

	// Example message with multiple commands
	message := `!help
!quickshab 10
!bring main`

	chatID := "1234567890@s.whatsapp.net"
	senderID := "9876543210@s.whatsapp.net"

	// Process all commands
	responses := router.ParseMultipleCommands(message, chatID, senderID)

	// Each response corresponds to each command
	for i, response := range responses {
		fmt.Printf("Response %d:\n%s\n\n", i+1, response)
	}
}

// Example of using individual command functions directly
func exampleDirectCommandCall() {
	// Initialize
	InitializeBot()
	defer db.Close()

	chatID := "1234567890@s.whatsapp.net"
	attendeeID := "9876543210@s.whatsapp.net"

	// Create a quickshab
	response1 := commands.QuickShabCmd(db, "!quickshab 15", chatID)
	fmt.Println(response1)

	// User signs up to bring main dish
	response2 := commands.BringCmd(db, "!bring main", chatID, attendeeID)
	fmt.Println(response2)

	// Show current status
	response3 := commands.ShowCmd(db, chatID)
	fmt.Println(response3)

	// Add to shopping list
	response4 := commands.ShopCmd(db, "!shop 3 wine bottles", chatID)
	fmt.Println(response4)

	// Show shopping list
	response5 := commands.ShopListCmd(db, chatID)
	fmt.Println(response5)
}

// Example of handling reminders in a background goroutine
func startReminderTicker(client *whatsmeow.Client) {
	// This would run in a goroutine to check and send reminders
	// ticker := time.NewTicker(1 * time.Minute)
	// defer ticker.Stop()

	// for range ticker.C {
	//     checkAndSendReminders(client)
	// }
}

// checkAndSendReminders queries the database for due reminders and sends them
func checkAndSendReminders(client *whatsmeow.Client) {
	// Query for reminders where time <= now and snoozable = 0
	rows, err := db.Query(`
		SELECT id, chat_id, message, type
		FROM reminders
		WHERE time > 0 AND time <= ? AND snoozable = 0
		ORDER BY time
	`, /* current unix timestamp */)

	if err != nil {
		fmt.Println("Error querying reminders:", err)
		return
	}
	defer rows.Close()

	// For each due reminder
	for rows.Next() {
		var id, chatID, message, reminderType string
		rows.Scan(&id, &chatID, &message, &reminderType)

		// Format reminder message
		var messageToSend string
		if reminderType == "remind" {
			messageToSend = fmt.Sprintf(
				"you wanted me to remind you:\n \"%s\".\n\n"+
				"To snooze this reminder, please reply with:\n"+
				"\"!snooze   new time\" within 24 hours.",
				message,
			)

			// Mark as snoozable
			db.Exec(`UPDATE reminders SET snoozable = 1 WHERE id = ?`, id)
		} else {
			// Type is "send" - just send the message and delete
			messageToSend = message
			db.Exec(`DELETE FROM reminders WHERE id = ?`, id)
		}

		// Send the message
		// sendResponseWithClient(client, chatID, messageToSend)
		fmt.Printf("Would send reminder to %s: %s\n", chatID, messageToSend)
	}
}
