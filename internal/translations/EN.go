package translations

// EN contains all English translations for the chatbot
var EN = map[string]string{
	// General messages
	"language_changed": "Bot language has been changed.",
	"current_language": "Current bot language: *English*. Change with /lang indonesia or /lang english.",
	"unknown_language": "Unknown language. Use /lang indonesia or /lang english.",

	// Bill management
	"bill_exists":         "A bill already exists in this chat. Please close it before creating a new one.",
	"bill_created":        "Created new bill: *%s*\nEveryone who wants to participate, please type /join",
	"no_bill":             "No active bill in this chat. Create one with /new first.",
	"bill_name_set":       "Bill name changed to *%s*.",
	"user_joined":         "*%s* joined the bill *%s*.",
	"user_already_joined": "*%s* is already a participant in bill *%s*.",
	"bill_closed":         "Bill *%s* has been closed.",

	// Item management
	"item_added":     "Added item: *%s* with price %s to bill *%s*",
	"invalid_amount": "Invalid amount. Example: /add Fried Rice 25000",

	// Commands
	"new_bill_usage":     "Please provide a bill name. Example: /new Breakfast or send /new Breakfast with a bill photo.",
	"add_contact_prompt": "To add participants, please send one or more WhatsApp contact attachments now. The bot will add those contacts as participants to the current bill.",
	"your_id":            "Your WhatsApp ID: %s",

	// Private message
	"private_message":        "*Bill Calculation Results: %s*\n\n%s\n\nPay to: %s",
	"private_message_failed": "Failed to send private message to %s (%s).",

	// Calculation
	"calculation_result": "*Bill Calculation Results: %s*\n\n%s\n\nTotal: %s\nNumber of participants: %d\nShare per person: %s",
	"no_participants":    "There are no participants in this bill. Please use /join or /participant to add participants.",
	"no_items":           "There are no items in this bill. Please add items with /add or send a bill photo.",

	// Help text
	"help_text": `*Split Bill Bot Help*

How to use the WhatsApp Split Bill Bot:

1. Create a new bill:
   /new <bill_name>
   or
   /new <bill_name> with a bill 📷 (send a photo to auto-extract items)
2. Each participant types /join to participate
3. Add items and amounts:
   /add <item_name> <amount>
   (You can skip this if you send a bill photo)
4. Calculate the split:
   /calculate
5. Close the bill when finished:
   /close (each participant will receive a private message with their share)

*Command List:*
/new [name] - Create a new bill
/add [item] [amount] - Add item to the bill
/join [bill_name] - Join the bill as a participant (optionally set/change bill name)
/participant - Add participants by sending their contact(s)
/calculate - Calculate and show how much each person owes
/close - Close the current bill and send private messages to all participants
/bill - Show bill details and participant list
/help - Show this help message
/myid - Show your WhatsApp ID
/lang [indonesia|english] - Change bot language preference for this chat

Example:
1. /new <bill_name> with a bill 📷 or /new <bill_name>
2. Everyone types /join or /participant with contact attachments
3. /add <item_name> <amount> (skip if you send a bill photo)
4. /calculate
5. /close when finished

About:
Created by Hendro Wibowo (@w33ladalah) and Affandy Fahrizain (@fhrzn)
https://github.com/w33ladalah/split-billing-whatsapp
`,
}
