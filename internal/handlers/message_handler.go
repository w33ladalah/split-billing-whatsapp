package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/w33ladalah/split-billing-whatsapp/internal/models"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	waProto "go.mau.fi/whatsmeow/binary/proto"
)

type MessageHandler struct {
	ActiveBills map[string]*models.Bill
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		ActiveBills: make(map[string]*models.Bill),
	}
}

func (h *MessageHandler) HandleMessage(client *whatsmeow.Client, msg *events.Message) {
	// Skip messages sent by our bot
	if msg.Info.IsFromMe {
		return
	}

	// Get message text
	if msg.Message.GetConversation() == "" && msg.Message.GetExtendedTextMessage() == nil {
		return
	}

	var text string
	if msg.Message.GetConversation() != "" {
		text = msg.Message.GetConversation()
	} else if msg.Message.GetExtendedTextMessage() != nil {
		text = msg.Message.GetExtendedTextMessage().GetText()
	}

	// Process command
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "/") {
		h.handleCommand(client, msg, text)
	}
}

func (h *MessageHandler) handleCommand(client *whatsmeow.Client, msg *events.Message, text string) {
	parts := strings.Fields(text)
	command := strings.ToLower(parts[0])

	chatID := msg.Info.Chat
	chatIDStr := chatID.String()

	switch command {
	case "/help":
		h.sendHelp(client, chatID)
	case "/newbill":
		if len(parts) < 2 {
			h.sendMessage(client, chatID, "Please provide a bill name. Example: /newbill Dinner")
			return
		}
		billName := strings.Join(parts[1:], " ")
		h.createBill(client, chatID, billName)
	case "/add":
		if len(parts) < 3 {
			h.sendMessage(client, chatID, "Please provide an item and amount. Example: /add Pizza 25.50")
			return
		}
		itemName := strings.Join(parts[1:len(parts)-1], " ")
		amount := parts[len(parts)-1]
		h.addItem(client, chatID, itemName, amount)
	case "/join":
		h.joinBill(client, chatID, msg.Info.Sender)
	case "/calculate":
		h.calculateBill(client, chatID)
	case "/close":
		h.closeBill(client, chatID)
	default:
		h.sendMessage(client, chatID, "Unknown command. Type /help for available commands.")
	}
}

func (h *MessageHandler) sendHelp(client *whatsmeow.Client, chatID types.JID) {
	helpText := `*Split Billing Bot Help*

Available commands:
/newbill [name] - Create a new bill
/add [item] [amount] - Add an item to the current bill
/join - Join the current bill as a participant
/calculate - Calculate and show how much each person owes
/close - Close the current bill
/help - Show this help message

Example usage:
1. /newbill Dinner at Restaurant
2. Each person types /join to participate
3. /add Pizza 25.50
4. /add Drinks 15.75
5. /calculate
6. /close when done
`
	h.sendMessage(client, chatID, helpText)
}

func (h *MessageHandler) createBill(client *whatsmeow.Client, chatID types.JID, name string) {
	chatIDStr := chatID.String()

	// Check if there's already an active bill
	if _, exists := h.ActiveBills[chatIDStr]; exists {
		h.sendMessage(client, chatID, "There's already an active bill in this chat. Please close it with /close first.")
		return
	}

	// Create new bill
	bill := models.NewBill(name)
	h.ActiveBills[chatIDStr] = bill

	h.sendMessage(client, chatID, fmt.Sprintf("Created new bill: *%s*\nEveryone who wants to participate, please type /join", name))
}

func (h *MessageHandler) joinBill(client *whatsmeow.Client, chatID types.JID, senderJID *types.JID) {
	chatIDStr := chatID.String()

	// Check if there's an active bill
	bill, exists := h.ActiveBills[chatIDStr]
	if !exists {
		h.sendMessage(client, chatID, "No active bill in this chat. Create one with /newbill first.")
		return
	}

	// Get user's name or phone number
	name := senderJID.User

	// Add participant
	added := bill.AddParticipant(name)
	if added {
		h.sendMessage(client, chatID, fmt.Sprintf("*%s* joined the bill.", name))
	} else {
		h.sendMessage(client, chatID, fmt.Sprintf("*%s* is already a participant.", name))
	}
}

func (h *MessageHandler) addItem(client *whatsmeow.Client, chatID types.JID, itemName, amountStr string) {
	chatIDStr := chatID.String()

	// Check if there's an active bill
	bill, exists := h.ActiveBills[chatIDStr]
	if !exists {
		h.sendMessage(client, chatID, "No active bill in this chat. Create one with /newbill first.")
		return
	}

	// Parse amount
	amount, err := bill.AddItem(itemName, amountStr)
	if err != nil {
		h.sendMessage(client, chatID, fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	h.sendMessage(client, chatID, fmt.Sprintf("Added *%s* ($%.2f) to the bill.", itemName, amount))
}

func (h *MessageHandler) calculateBill(client *whatsmeow.Client, chatID types.JID) {
	chatIDStr := chatID.String()

	// Check if there's an active bill
	bill, exists := h.ActiveBills[chatIDStr]
	if !exists {
		h.sendMessage(client, chatID, "No active bill in this chat. Create one with /newbill first.")
		return
	}

	// Calculate bill
	if len(bill.Participants) == 0 {
		h.sendMessage(client, chatID, "No participants in the bill. Ask people to join with /join first.")
		return
	}

	if len(bill.Items) == 0 {
		h.sendMessage(client, chatID, "No items in the bill. Add items with /add first.")
		return
	}

	// Generate summary
	summary := bill.GenerateSummary()
	h.sendMessage(client, chatID, summary)
}

func (h *MessageHandler) closeBill(client *whatsmeow.Client, chatID types.JID) {
	chatIDStr := chatID.String()

	// Check if there's an active bill
	bill, exists := h.ActiveBills[chatIDStr]
	if !exists {
		h.sendMessage(client, chatID, "No active bill in this chat.")
		return
	}

	// Generate final summary
	summary := fmt.Sprintf("*Bill Closed: %s*\n\n", bill.Name)
	summary += bill.GenerateSummary()

	// Delete the bill
	delete(h.ActiveBills, chatIDStr)

	h.sendMessage(client, chatID, summary)
	h.sendMessage(client, chatID, "The bill has been closed. Start a new one with /newbill.")
}

func (h *MessageHandler) sendMessage(client *whatsmeow.Client, chatID types.JID, text string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.SendMessage(ctx, chatID, &waProto.Message{
		Conversation: &text,
	})

	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}
}
