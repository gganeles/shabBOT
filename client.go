package main

import (
	"context"
	"database/sql"
	"fmt"
	"shabBOT/commands"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
)

type ShabClient struct {
	WAClient       *whatsmeow.Client
	context        context.Context
	DB             *sql.DB
	EventHandlerID uint32
	router         *commands.CommandRouter
}

func (shabcli *ShabClient) register() {
	shabcli.EventHandlerID = shabcli.WAClient.AddEventHandler(shabcli.respond)
}

func (shabcli *ShabClient) respond(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:

		text := v.Message.GetConversation()
		if text == "" {
			return
		}

		fmt.Printf("Received message from %s: \n\"%s\"\n", v.Info.Sender.User, text)
		for _, response := range shabcli.router.ParseMultipleCommands(text, v.Info) {
			fmt.Println("sending response:", response)
			_, err := shabcli.WAClient.SendMessage(shabcli.context, v.Info.Chat, &waE2E.Message{
				Conversation: &response,
			})
			if err != nil {
				fmt.Println("Error sending message:", err)
			}
		}

	}

}
