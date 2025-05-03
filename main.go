package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/w33ladalah/split-billing-whatsapp/internal/bot"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading environment variables:", err)
	}

	// Create a new bot instance
	whatsappBot, err := bot.NewBot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing bot: %v\n", err)
		os.Exit(1)
	}

	// Connect to WhatsApp
	if err := whatsappBot.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to WhatsApp: %v\n", err)
		os.Exit(1)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Received shutdown signal, disconnecting...")
		whatsappBot.Disconnect()
		cancel()
	}()

	fmt.Println("Split-billing WhatsApp bot started. Press Ctrl+C to exit.")
	<-ctx.Done()
}
