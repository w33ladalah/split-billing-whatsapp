# Split Billing WhatsApp Bot

A WhatsApp bot built with Go and whatsmeow that helps friends split bills easily.

## Features

- Create and manage bills in WhatsApp groups
- Add participants and items
- Calculate how much each person owes
- Support for different bill types

## Installation

1. Ensure you have Go 1.16+ installed
2. Clone the repository
3. Install dependencies:
   ```
   go mod tidy
   ```
4. Build the application:
   ```
   go build -o build/split-billing-bot cmd/main.go
   ```

## Usage

1. Run the application:
   ```
   ./build/split-billing-bot
   ```
2. Scan the QR code with WhatsApp to log in
3. The bot will now process commands in your chats

## Commands

- `/newbill [name]` - Create a new bill
- `/add [item] [amount]` - Add an item to the current bill
- `/join` - Join the current bill as a participant
- `/calculate` - Calculate and show how much each person owes
- `/close` - Close the current bill
- `/help` - Show available commands

## Example

1. `/newbill Dinner at Restaurant`
2. Each person types `/join` to participate
3. `/add Pizza 25.50`
4. `/add Drinks 15.75`
5. `/calculate`
6. `/close` when done

## License

MIT
