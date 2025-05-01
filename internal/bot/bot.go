package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/w33ladalah/split-billing-whatsapp/internal/handlers"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "github.com/mattn/go-sqlite3"
)

type Bot struct {
	client     *whatsmeow.Client
	eventsChan chan interface{}
	handler    *handlers.MessageHandler
}

func NewBot() (*Bot, error) {
	// Create WhatsApp store
	dbPath := "./whatsapp-data.db"
	container, err := sqlstore.New("sqlite3", "file:"+dbPath+"?_foreign_keys=on", waLog.Stdout("Database", "DEBUG", true))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get device store
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	// Create WhatsApp client
	client := whatsmeow.NewClient(deviceStore, waLog.Stdout("Client", "INFO", true))

	// Create message handler
	handler := handlers.NewMessageHandler()

	bot := &Bot{
		client:     client,
		eventsChan: make(chan interface{}, 100),
		handler:    handler,
	}

	// Register event handler
	client.AddEventHandler(bot.eventHandler)

	return bot, nil
}

func (b *Bot) Connect() error {
	if b.client.Store.ID == nil {
		// No device ID, need to pair
		qrChan, _ := b.client.GetQRChannel(context.Background())
		err := b.client.Connect()
		if err != nil {
			return err
		}

		for evt := range qrChan {
			if evt.Event == "code" {
				// Print the QR code to the terminal
				fmt.Println("QR code:", evt.Code)
				fmt.Println("Scan this QR code with your WhatsApp app to log in")
			} else {
				fmt.Println("QR event:", evt.Event)
			}
		}
	} else {
		// Already have device ID, just connect
		err := b.client.Connect()
		if err != nil {
			return err
		}
	}

	// Wait for connection to establish
	for !b.client.IsConnected() {
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("Connected to WhatsApp")
	return nil
}

func (b *Bot) Disconnect() {
	b.client.Disconnect()
}

func (b *Bot) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		// Handle incoming messages
		go b.handler.HandleMessage(b.client, v)
	}
}
