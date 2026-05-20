# NovaUserbot

![NovaUserbot Logo](https://files.indrajeeth.in/nova.jpg)

NovaUserbot is a powerful and flexible userbot written in Go. It leverages the Telegram API to automate tasks and enhance your Telegram experience.

## Features

- 🤖 Easy to set up and use
- 🚀 High performance with Go
- 🔒 Secure and reliable
- 📚 Built using the Gogram library
- 🗄️ Uses PostgreSQL for database management

## Requirements

- Go 1.23 or higher

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/TAMILVIP007/NovaUserbot.git
    cd NovaUserbot
    ```

2. Set up your environment variables:
    ```sh
    export API_ID=your_api_id
    export API_HASH=your_api_hash
    export TOKEN=your_bot_token
    export STRING_SESSION=your_string_session
    export DB_URL=your_database_url
    ```

3. Run the bot:
    ```sh
    go run .
    ```

4. Alternatively, build and run the bot:
    ```sh
    go build -o novauserbot
    ./novauserbot
    ```

 ## Generate String Session

 To generate a string session, you can use the following command:

    curl -O https://gist.githubusercontent.com/TAMILVIP007/da9d52f3eb48b6d88f10aa634077bcac/raw/a7e082d9405ba0ba6e329ed5a409e25b56c13c95/sessionge.go && [ ! -f go.mod ] && go mod init GoGramSession || echo "go.mod already exists" && go mod tidy && go run sessionge.go && rm sessionge.go

## Environment Variables

- `API_ID`: Your Telegram API ID (e.g., `123`)
- `API_HASH`: Your Telegram API Hash (e.g., `abcdef1234567890abcdef1234567890`)
- `TOKEN`: Your bot token (e.g., `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11`)
- `STRING_SESSION`: Your string session (e.g., `1BvXWG`)
- `DB_URL`: Your database URL (e.g., `postgresql://username:password@hostname:port/database`)


## Credits

This project is maintained by [TAMILVIP007](https://github.com/tamilvip007).

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

Happy botting! 🚀