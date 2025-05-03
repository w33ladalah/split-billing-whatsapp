package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/w33ladalah/split-billing-whatsapp/internal/models"
	"github.com/w33ladalah/split-billing-whatsapp/internal/processor"
	"github.com/w33ladalah/split-billing-whatsapp/internal/translations"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	waProto "go.mau.fi/whatsmeow/binary/proto"
)

type MessageHandler struct {
	// ActiveBill[chatID] = *Bill
	ActiveBill map[string]*models.Bill
	// ExpectingContacts[chatID] = true if waiting for contact attachments
	ExpectingContacts map[string]bool
	// LanguagePrefs[chatID] = "indonesia" or "english"
	LanguagePrefs map[string]string
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		ActiveBill:        make(map[string]*models.Bill),
		ExpectingContacts: make(map[string]bool),
		LanguagePrefs:     make(map[string]string),
	}
}

// getTranslations returns the translation map for the given chatID/user preference
func (h *MessageHandler) getTranslations(chatID string) map[string]string {
	lang := h.LanguagePrefs[chatID]
	if lang == "english" {
		return translations.EN
	}
	return translations.ID // default
}

func (h *MessageHandler) HandleMessage(client *whatsmeow.Client, msg *events.Message) {
	chatIDStr := msg.Info.Chat.String()

	// Handle contact attachment if expecting participants
	if h.ExpectingContacts[chatIDStr] {
		bill := h.ActiveBill[chatIDStr]
		if bill == nil {
			tr := h.getTranslations(chatIDStr)
			h.sendMessage(client, msg.Info.Chat, tr["no_bill"])
			h.ExpectingContacts[chatIDStr] = false
			return
		}
		// Handle single contact
		if msg.Message.GetContactMessage() != nil {
			contact := msg.Message.GetContactMessage()
			name := contact.GetDisplayName()
			jid := ""
			if contact.GetVcard() != "" {
				jid = extractJIDFromVCard(contact.GetVcard())
			}
			added := bill.AddParticipant(name, jid)
			if added {
				tr := h.getTranslations(chatIDStr)
				h.sendMessage(client, msg.Info.Chat, fmt.Sprintf(tr["user_joined"], name, bill.Name))
			} else {
				tr := h.getTranslations(chatIDStr)
				h.sendMessage(client, msg.Info.Chat, fmt.Sprintf(tr["user_already_joined"], name, bill.Name))
			}
			h.ExpectingContacts[chatIDStr] = false
			return
		}
		// Handle multiple contacts (contacts array)
		if msg.Message.GetContactsArrayMessage() != nil {
			contacts := msg.Message.GetContactsArrayMessage().GetContacts()
			var addedNames []string
			for _, c := range contacts {
				name := c.GetDisplayName()
				jid := ""
				if c.GetVcard() != "" {
					jid = extractJIDFromVCard(c.GetVcard())
				}
				added := bill.AddParticipant(name, jid)
				if added {
					addedNames = append(addedNames, name)
				}
			}
			if len(addedNames) > 0 {
				h.sendMessage(client, msg.Info.Chat, "Added participants: "+strings.Join(addedNames, ", "))
			} else {
				h.sendMessage(client, msg.Info.Chat, "All contacts are already participants.")
			}
			h.ExpectingContacts[chatIDStr] = false
			return
		}
	}

	// Skip messages sent by our bot
	if msg.Info.IsFromMe {
		return
	}

	// Get message text
	if msg.Message.GetConversation() == "" && msg.Message.GetExtendedTextMessage() == nil && msg.Message.GetImageMessage() == nil {
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
		return
	}

	// Handle image message with caption as command
	img := msg.Message.GetImageMessage()
	if img != nil {
		caption := strings.TrimSpace(img.GetCaption())
		if strings.HasPrefix(caption, "/") {
			h.handleCommand(client, msg, fmt.Sprintf("%s %s", caption, "with image"))
			return
		}
	}
}

func (h *MessageHandler) handleCommand(client *whatsmeow.Client, msg *events.Message, text string) {
	parts := strings.Fields(text)
	command := strings.ToLower(parts[0])

	chatID := msg.Info.Chat
	chatIDStr := chatID.String()
	tr := h.getTranslations(chatIDStr)

	switch command {
	case "/lang":
		lang := ""
		if len(parts) > 1 {
			lang = strings.ToLower(parts[1])
		}
		if lang == "indonesia" || lang == "id" {
			h.LanguagePrefs[chatIDStr] = "indonesia"
			h.sendMessage(client, chatID, tr["language_changed"])
		} else if lang == "english" || lang == "en" {
			h.LanguagePrefs[chatIDStr] = "english"
			tr = h.getTranslations(chatIDStr) // update after change
			h.sendMessage(client, chatID, tr["language_changed"])
		} else if lang == "" {
			current := h.LanguagePrefs[chatIDStr]
			if current == "" {
				current = "indonesia"
			}
			tr = h.getTranslations(chatIDStr)
			h.sendMessage(client, chatID, tr["current_language"])
		} else {
			h.sendMessage(client, chatID, tr["unknown_language"])
		}
		return

	case "/myid":
		h.sendMessage(client, chatID, fmt.Sprintf(tr["your_id"], msg.Info.Sender.String()))
	case "/help":
		h.sendMessage(client, chatID, tr["help_text"])
	case "/participant":
		h.handleParticipantCommand(client, chatID)
	case "/new":
		if len(parts) < 2 {
			h.sendMessage(client, chatID, tr["new_bill_usage"])
			return
		}

		// Try to get image message
		img := msg.Message.GetImageMessage()
		var billName string

		if img != nil {
			caption := img.GetCaption()
			if strings.HasPrefix(caption, "/new") {
				// Extract bill name from caption
				billName = strings.TrimSpace(strings.TrimPrefix(caption, "/new"))
			} else {
				// Fallback: extract from text (for non-caption usage)
				billName = strings.TrimSpace(strings.TrimPrefix(text, "/new"))
			}
		} else {
			// No image: extract from text
			billName = strings.TrimSpace(strings.TrimPrefix(text, "/new"))
		}

		// Check for image data
		var imgData []byte
		if img != nil {
			if img.GetURL() != "" && img.GetDirectPath() != "" {
				data, err := client.Download(img)
				if err == nil {
					imgData = data
				}
			}
		}
		if img != nil {
			if img.GetURL() != "" && img.GetDirectPath() != "" {
				data, err := client.Download(img)
				if err == nil {
					imgData = data
				}
			}
		}

		if imgData != nil {
			proc := processor.NewImageProcessor()
			bill, err := proc.ProcessBillImage(imgData)
			if err != nil {
				h.sendMessage(client, chatID, "Failed to process bill image: "+err.Error())
				return
			}
			chatIDStr := chatID.String()
			if h.ActiveBill[chatIDStr] != nil {
				h.sendMessage(client, chatID, "A bill already exists in this chat. Please close it before creating a new one.")
				return
			}
			bill.Name = billName
			h.ActiveBill[chatIDStr] = bill
			summary := fmt.Sprintf("Created new bill from image: *%s*\n", billName)
			summary += bill.GenerateSummary()
			h.sendMessage(client, chatID, summary)
			h.sendMessage(client, chatID, "Everyone who wants to participate, please type _/join_ or _/participant_ with contact attachments")
			return
		}
		h.createBill(client, chatID, billName)
	case "/add":
		if len(parts) < 3 {
			h.sendMessage(client, chatID, "Please provide item and amount. Example: /add Nasi_Goreng 25000")
			return
		}
		itemName := strings.Join(parts[1:len(parts)-1], " ")
		amount := parts[len(parts)-1]
		h.addItem(client, chatID, itemName, amount)
	case "/join":
		var billName string
		if len(parts) > 1 {
			billName = strings.Join(parts[1:], " ")
		}
		h.joinBill(client, chatID, &msg.Info.Sender, billName)
	case "/calculate":
		h.calculateBill(client, chatID)
	case "/close":
		h.closeBill(client, chatID)
	default:
		h.sendMessage(client, chatID, "Unknown command. Type /help for available commands.")
	}
}

// handleParticipantCommand responds with instructions for adding participants via contact attachments
func (h *MessageHandler) handleParticipantCommand(client *whatsmeow.Client, chatID types.JID) {
	chatIDStr := chatID.String()
	tr := h.getTranslations(chatIDStr)
	h.ExpectingContacts[chatIDStr] = true
	h.sendMessage(client, chatID, tr["add_contact_prompt"])
}

func (h *MessageHandler) createBill(client *whatsmeow.Client, chatID types.JID, name string) {
	chatIDStr := chatID.String()
	tr := h.getTranslations(chatIDStr)

	if h.ActiveBill[chatIDStr] != nil {
		h.sendMessage(client, chatID, tr["bill_exists"])
		return
	}

	bill := models.NewBill(name)
	h.ActiveBill[chatIDStr] = bill

	h.sendMessage(client, chatID, fmt.Sprintf(tr["bill_created"], name))
}

func extractJIDFromVCard(vcard string) string {
	// Looks for waid=XXXXXXXXXX in the vCard and returns WhatsApp JID format
	waidPrefix := "waid="
	idx := strings.Index(vcard, waidPrefix)
	if idx == -1 {
		return ""
	}
	start := idx + len(waidPrefix)
	end := start
	for end < len(vcard) && ((vcard[end] >= '0' && vcard[end] <= '9') || vcard[end] == '+') {
		end++
	}
	waid := vcard[start:end]
	if waid != "" {
		return waid + "@s.whatsapp.net"
	}
	return ""
}

func (h *MessageHandler) joinBill(client *whatsmeow.Client, chatID types.JID, senderJID *types.JID, billName string) {
	chatIDStr := chatID.String()
	tr := h.getTranslations(chatIDStr)
	bill := h.ActiveBill[chatIDStr]
	if bill == nil {
		h.sendMessage(client, chatID, tr["no_bill"])
		return
	}
	if billName != "" {
		bill.Name = billName
		h.sendMessage(client, chatID, fmt.Sprintf(tr["bill_name_set"], bill.Name))
	}
	name := senderJID.User
	jid := senderJID.User + "@s.whatsapp.net"
	added := bill.AddParticipant(name, jid)
	if added {
		h.sendMessage(client, chatID, fmt.Sprintf(tr["user_joined"], name, bill.Name))
	} else {
		h.sendMessage(client, chatID, fmt.Sprintf(tr["user_already_joined"], name, bill.Name))
	}
}

func (h *MessageHandler) addItem(client *whatsmeow.Client, chatID types.JID, itemName, amountStr string) {
	chatIDStr := chatID.String()
	tr := h.getTranslations(chatIDStr)
	bill := h.ActiveBill[chatIDStr]
	if bill == nil {
		h.sendMessage(client, chatID, tr["no_bill"])
		return
	}
	amount, err := bill.AddItem(itemName, amountStr)
	if err != nil {
		h.sendMessage(client, chatID, tr["invalid_amount"])
		return
	}
	h.sendMessage(client, chatID, fmt.Sprintf(tr["item_added"], itemName, formatIDRLocal(amount), bill.Name))
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

func (h *MessageHandler) calculateBill(client *whatsmeow.Client, chatID types.JID) {
	chatIDStr := chatID.String()
	tr := h.getTranslations(chatIDStr)
	bill := h.ActiveBill[chatIDStr]
	if bill == nil {
		h.sendMessage(client, chatID, tr["no_bill"])
		return
	}
	if len(bill.Participants) == 0 {
		h.sendMessage(client, chatID, tr["no_participants"])
		return
	}
	if len(bill.Items) == 0 {
		h.sendMessage(client, chatID, tr["no_items"])
		return
	}
	summary := bill.GenerateSummary()
	h.sendMessage(client, chatID, summary)
}

func (h *MessageHandler) closeBill(client *whatsmeow.Client, chatID types.JID) {
	chatIDStr := chatID.String()
	tr := h.getTranslations(chatIDStr)
	bill := h.ActiveBill[chatIDStr]
	if bill == nil {
		h.sendMessage(client, chatID, tr["no_bill"])
		return
	}
	summary := "*BILL CLOSED*\n\n"
	summary += bill.GenerateSummary()

	// Send private message to each participant
	for _, p := range bill.Participants {
		fmt.Println("[DEBUG] Sending private message to:", p.Name)
		partsJID := strings.Split(p.JID, "@")
		jid := types.NewJID(partsJID[0], "s.whatsapp.net")
		fmt.Println("[DEBUG] JID:", jid)
		receiver := p.JID[:strings.Index(p.JID, "@")]
		personalMsg := fmt.Sprintf(tr["private_message"], bill.Name, summary, types.NewJID(receiver, "s.whatsapp.net").User)
		fmt.Println("[DEBUG] Personal message:", personalMsg)
		if jid == chatID {
			h.sendMessage(client, chatID, personalMsg)
		} else {
			err := h.sendMessageWithError(client, jid, personalMsg)
			if err != nil {
				h.sendMessage(client, chatID, fmt.Sprintf(tr["private_message_failed"], p.Name, p.JID))
			}
		}
	}

	delete(h.ActiveBill, chatIDStr)
	h.sendMessage(client, chatID, summary)
	h.sendMessage(client, chatID, fmt.Sprintf(tr["bill_closed"], bill.Name))
}

func (h *MessageHandler) sendMessageWithError(client *whatsmeow.Client, chatID types.JID, text string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.SendMessage(ctx, chatID, &waProto.Message{
		Conversation: &text,
	})
	return err
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
