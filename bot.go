package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"shabBOT/commands"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var db *sql.DB

func main() {
	dbLog := waLog.Stdout("Database", "INFO", true)
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite3", "file:userstore.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", "INFO", true)
	WAClient := whatsmeow.NewClient(deviceStore, clientLog)

	// Initialize or connect to the SQLite database
	db, err = sql.Open("sqlite3", "./shabbot.db?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Configure connection pool to prevent database locking
	db.SetMaxOpenConns(1) // SQLite works best with a single connection
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	// Initialize database schema
	if err := commands.InitDB(db); err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}

	router := commands.NewCommandRouter(db)

	shabClient := &ShabClient{
		WAClient:       WAClient,
		context:        ctx,
		DB:             db,
		EventHandlerID: 0,
		router:         router,
	}

	shabClient.register()

	client := shabClient.WAClient

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}

		// Start a goroutine to print QR events so we still see them if pairing by phone is not used

		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}

		// If PAIR_PHONE env var is set, try code-based pairing
		phone := os.Getenv("PAIR_PHONE")
		if phone != "" {
			// Wait briefly to ensure the websocket and server state are ready.
			// The docs recommend waiting for the first QR event; sleeping a second is usually enough.
			time.Sleep(2 * time.Second)
			// Use a generic browser-like client display name. PairClientType can be left to PairClientOtherWebClient constant via 0..n mapping; here we pass PairClientOtherWebClient by value from whatsmeow
			code, err := client.PairPhone(context.Background(), "12038025238", true, whatsmeow.PairClientOtherWebClient, "Chrome (Linux)")
			if err != nil {
				fmt.Println("PairPhone error:", err)
			} else {
				fmt.Println("Phone pairing code:", code)
				if err := os.WriteFile("pairing_code.txt", []byte(code+"\n"), 0644); err != nil {
					fmt.Println("Failed to save pairing code:", err)
				} else {
					fmt.Println("Saved pairing code to pairing_code.txt")
				}
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	// Start the time routine to check for reminders
	go timeRoutine(client, db, ctx)

	// Listen to Ctrl + C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}
