# Split Billing WhatsApp Bot

A WhatsApp bot built with Go and whatsmeow that helps friends split bills easily.

## Features

- Create and manage bills in WhatsApp groups
- Add participants and items
- Calculate how much each person owes

## Installation

1. Ensure you have Go 1.16+ installed
2. Clone the repository
3. Install dependencies:

   ```shell
   go mod tidy
   ```

4. Build the application:

   ```shell
   go build -o build/split-billing-bot cmd/main.go
   ```

## Web QR Code UI

- When you start the bot, it also runs a web server at [http://localhost:8080/](http://localhost:8080/).
- Open this page in your browser to view and scan the WhatsApp QR code for login.

## Usage

1. Run the application:

   ```shell
   ./build/split-billing-bot
   ```

2. Open your browser and go to [http://localhost:8080/](http://localhost:8080/) to view the QR code.
3. Scan the QR code with WhatsApp to log in.
4. The bot will now process commands in your chats.

## Commands

- `/newbill [name]` - Create a new bill
- `/add [bill_name] [item] [amount]` - Add an item to a bill
- `/join [bill_name]` - Join a bill as a participant
- `/calculate [bill_name]` - Calculate and show how much each person owes in a bill
- `/close [bill_name]` - Close a bill
- `/help` - Show available commands

## Example

1. `/newbill Sarapan`
2. Each person types `/join Sarapan` to participate
3. `/add Sarapan Nasi Goreng 25000`
4. `/add Sarapan Ayam Goreng 15000`
5. `/calculate Sarapan`
6. `/close Sarapan` when done

## TODO

- [ ] Add image processing for bill images
- [ ] Add support for different bill types
- [ ] Add support for different languages
- [ ] Add support for different timezones
- [ ] Send message to the participants WhatsApp number when bill is closed
- [ ] When user types `/join`, send message to the user that they have joined the bill
- [ ] Add bill name to the `/join` command
- [ ] Add bill name to the `/calculate` command
- [ ] Add bill name to the `/close` command
- [x] Add bill name to the `/help` command
- [ ] Change the database server to use PostgreSQL
- [ ] Use Docker to containerize the application
- [x] Use GitHub Actions to deploy the application to a VPS

## License

MIT
