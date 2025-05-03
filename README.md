# Split Billing WhatsApp Bot

A WhatsApp bot built with Go and whatsmeow that helps friends split bills easily.

## Features

- Create and manage bills in WhatsApp groups
- Add participants and items
- Send a photo of your bill to extract items and amounts automatically
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

- `/new [name]` - Create a new bill
- `/add [item] [amount]` - Add an item to the current bill (optional)
- Send a photo of your bill to add all items at once (optional)
- `/join [bill_name]` - Join the current bill as a participant (optional)
- `/calculate` - Calculate and show how much each person owes in the current bill
- `/close` - Close the current bill
- `/help` - Show available commands

## Example

1. `/new Sarapan`
2. Each person types `/join` to participate
3. Send a photo of your bill to add all items at once
4. `/calculate`
5. `/close` when done

## TODO

- [ ] Add image processing for bill images using LLM
- [ ] Send message to the participants WhatsApp number when bill is closed
- [ ] When user types `/join`, send message to the user that they have joined the bill
- [x] Add bill name to the `/join` command
- [x] Add bill name to the `/calculate` command
- [x] Add bill name to the `/close` command
- [x] Add bill name to the `/help` command
- [ ] Change the database server to use PostgreSQL
- [ ] Use Docker to containerize the application
- [x] Use GitHub Actions to deploy the application to a VPS
- [ ] Add unit tests
- [ ] Add integration tests

## License

MIT
