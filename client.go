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

		//fmt.Println("Received message from:", v.Info.PushName)
		for _, response := range shabcli.router.ParseMultipleCommands(text, v.Info) {
			_, err := shabcli.WAClient.SendMessage(shabcli.context, v.Info.Chat, &waE2E.Message{
				Conversation: &response,
			})
			if err != nil {
				fmt.Println("Error sending message:", err)
			}
		}

	}

}
