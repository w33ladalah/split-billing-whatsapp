package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/w33ladalah/split-billing-whatsapp/internal/models"
	"github.com/w33ladalah/split-billing-whatsapp/internal/processor"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	waProto "go.mau.fi/whatsmeow/binary/proto"
)

type MessageHandler struct {
	// ActiveBills[chatID][billName] = *Bill
	ActiveBills map[string]map[string]*models.Bill
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		ActiveBills: make(map[string]map[string]*models.Bill),
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

	switch command {
	case "/help":
		h.sendHelp(client, chatID)
	case "/newbill":
		if len(parts) < 2 {
			h.sendMessage(client, chatID, "Please provide a bill name. Example: /newbill Sarapan or send /newbill Sarapan with a bill photo.")
			return
		}
		billName := parts[1]

		// Check for image attachment
		var imgData []byte
		if msg.Message.GetImageMessage() != nil {
			img := msg.Message.GetImageMessage()
			if img.GetURL() != "" && img.GetDirectPath() != "" {
				data, err := client.Download(img)
				if err == nil {
					imgData = data
				}
			}
		}

		if imgData != nil {
			// Process the image with OpenAI API
			proc := processor.NewImageProcessor()
			bill, err := proc.ProcessBillImage(imgData)
			if err != nil {
				h.sendMessage(client, chatID, fmt.Sprintf("Failed to process bill image: %v", err))
				return
			}
			// Save the bill and send summary
			chatIDStr := chatID.String()
			if h.ActiveBills[chatIDStr] == nil {
				h.ActiveBills[chatIDStr] = make(map[string]*models.Bill)
			}
			if _, exists := h.ActiveBills[chatIDStr][billName]; exists {
				h.sendMessage(client, chatID, fmt.Sprintf("A bill named '%s' already exists in this chat. Use a different name.", billName))
				return
			}
			bill.Name = billName
			h.ActiveBills[chatIDStr][billName] = bill
			summary := fmt.Sprintf("Created new bill from image: *%s*\n", billName)
			summary += bill.GenerateSummary()
			h.sendMessage(client, chatID, summary)
			h.sendMessage(client, chatID, fmt.Sprintf("Everyone who wants to participate, please type /join %s", billName))
			return
		}
		// No image, fallback to normal create
		h.createBill(client, chatID, billName)
	case "/add":
		if len(parts) < 4 {
			h.sendMessage(client, chatID, "Please provide bill name, item, and amount. Example: /add Sarapan Nasi_Goreng 25000")
			return
		}
		billName := parts[1]
		itemName := strings.Join(parts[2:len(parts)-1], " ")
		amount := parts[len(parts)-1]
		h.addItem(client, chatID, billName, itemName, amount)
	case "/join":
		if len(parts) < 2 {
			h.sendMessage(client, chatID, "Please provide bill name. Example: /join Sarapan")
			return
		}
		billName := parts[1]
		h.joinBill(client, chatID, billName, &msg.Info.Sender)
	case "/calculate":
		if len(parts) < 2 {
			h.sendMessage(client, chatID, "Please provide bill name. Example: /calculate Sarapan")
			return
		}
		billName := parts[1]
		h.calculateBill(client, chatID, billName)
	case "/close":
		if len(parts) < 2 {
			h.sendMessage(client, chatID, "Please provide bill name. Example: /close Sarapan")
			return
		}
		billName := parts[1]
		h.closeBill(client, chatID, billName)
	default:
		h.sendMessage(client, chatID, "Unknown command. Type /help for available commands.")
	}
}

func (h *MessageHandler) sendHelp(client *whatsmeow.Client, chatID types.JID) {
	helpText := `*Split Billing Bot Help*

_How to Use WhatsApp Split Bill Bot:_

1. Create a new bill:
   /newbill Breakfast at Padang Restaurant
2. Each participant types /join to participate
3. Add items and amounts:
   /add Fried Rice 25000
   /add Fried Chicken 15000
4. Calculate the split:
   /calculate
5. Close the bill when finished:
   /close

*Command List:*
/newbill [name] - Create a new bill
/add [item] [amount] - Add item to the bill
/join - Join the bill as a participant
/calculate - Calculate and show the split
/close - Close the bill
/help - Show usage instructions and command list

Usage example:
1. /newbill Breakfast at Padang Restaurant
2. Everyone types /join
3. /add Fried Rice 25000
4. /add Fried Chicken 15000
5. /calculate
6. /close when finished
`
	h.sendMessage(client, chatID, helpText)
}

func (h *MessageHandler) createBill(client *whatsmeow.Client, chatID types.JID, name string) {
	chatIDStr := chatID.String()

	if h.ActiveBills[chatIDStr] == nil {
		h.ActiveBills[chatIDStr] = make(map[string]*models.Bill)
	}
	if _, exists := h.ActiveBills[chatIDStr][name]; exists {
		h.sendMessage(client, chatID, fmt.Sprintf("A bill named '%s' already exists in this chat. Use a different name.", name))
		return
	}

	bill := models.NewBill(name)
	h.ActiveBills[chatIDStr][name] = bill

	h.sendMessage(client, chatID, fmt.Sprintf("Created new bill: *%s*\nEveryone who wants to participate, please type /join %s", name, name))
}

func (h *MessageHandler) joinBill(client *whatsmeow.Client, chatID types.JID, billName string, senderJID *types.JID) {
	chatIDStr := chatID.String()
	bills, exists := h.ActiveBills[chatIDStr]
	if !exists || bills[billName] == nil {
		h.sendMessage(client, chatID, fmt.Sprintf("No active bill named '%s' in this chat. Create one with /newbill %s first.", billName, billName))
		return
	}
	bill := bills[billName]
	name := senderJID.User
	added := bill.AddParticipant(name)
	if added {
		h.sendMessage(client, chatID, fmt.Sprintf("*%s* joined the bill *%s*.", name, billName))
	} else {
		h.sendMessage(client, chatID, fmt.Sprintf("*%s* is already a participant in bill *%s*.", name, billName))
	}
}

func (h *MessageHandler) addItem(client *whatsmeow.Client, chatID types.JID, billName, itemName, amountStr string) {
	chatIDStr := chatID.String()
	bills, exists := h.ActiveBills[chatIDStr]
	if !exists || bills[billName] == nil {
		h.sendMessage(client, chatID, fmt.Sprintf("No active bill named '%s' in this chat. Create one with /newbill %s first.", billName, billName))
		return
	}
	bill := bills[billName]
	amount, err := bill.AddItem(itemName, amountStr)
	if err != nil {
		h.sendMessage(client, chatID, fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	h.sendMessage(client, chatID, fmt.Sprintf("Added *%s* (%s) to bill *%s*.", itemName, formatIDRLocal(amount), billName))
}

// formatIDRLocal is a local copy of the formatIDR function for Indonesian Rupiah formatting
func formatIDRLocal(amount float64) string {
	n := int64(amount + 0.5) // round to nearest rupiah
	s := fmt.Sprintf("%d", n)
	var out []byte
	count := 0
	for i := len(s) - 1; i >= 0; i-- {
		if count > 0 && count%3 == 0 {
			out = append([]byte{"."[0]}, out...)
		}
		out = append([]byte{s[i]}, out...)
		count++
	}
	return "Rp" + string(out)
}


func (h *MessageHandler) calculateBill(client *whatsmeow.Client, chatID types.JID, billName string) {
	chatIDStr := chatID.String()
	bills, exists := h.ActiveBills[chatIDStr]
	if !exists || bills[billName] == nil {
		h.sendMessage(client, chatID, fmt.Sprintf("No active bill named '%s' in this chat. Create one with /newbill %s first.", billName, billName))
		return
	}
	bill := bills[billName]
	if len(bill.Participants) == 0 {
		h.sendMessage(client, chatID, "No participants in the bill. Ask people to join with /join first.")
		return
	}
	if len(bill.Items) == 0 {
		h.sendMessage(client, chatID, "No items in the bill. Add items with /add first.")
		return
	}
	summary := bill.GenerateSummary()
	h.sendMessage(client, chatID, summary)
}

func (h *MessageHandler) closeBill(client *whatsmeow.Client, chatID types.JID, billName string) {
	chatIDStr := chatID.String()
	bills, exists := h.ActiveBills[chatIDStr]
	if !exists || bills[billName] == nil {
		h.sendMessage(client, chatID, fmt.Sprintf("No active bill named '%s' in this chat.", billName))
		return
	}
	bill := bills[billName]
	summary := fmt.Sprintf("*Bill Closed: %s*\n\n", bill.Name)
	summary += bill.GenerateSummary()
	// Delete the bill
	delete(bills, billName)
	if len(bills) == 0 {
		delete(h.ActiveBills, chatIDStr)
	}
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
