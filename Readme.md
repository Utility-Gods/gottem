# Gottem CLI

Gottem is a command-line interface (CLI) application that allows users to interact with various AI APIs, such as Claude and OpenAI, through a chat-like interface. It provides a user-friendly way to send queries to the APIs and view the responses in a structured manner.

## Features

- Start new chats or continue previous chats
- Select from multiple AI APIs to send queries
- Persistent storage of chat history and API keys
- Intuitive text editor for composing queries and viewing responses
- Customizable settings for API keys and preferences

## Installation

1. Make sure you have Go installed on your system. You can download it from the official website: [https://golang.org/dl/](https://golang.org/dl/)

2. Clone the repository:
   ```
   git clone https://github.com/Utility-Gods/gottem.git
   ```

3. Navigate to the project directory:
   ```
   cd gottem
   ```

4. Build the application:
   ```
   go build -o gottem ./cmd/cli
   ```

## Usage

To run the Gottem CLI, simply execute the built binary:
```
./gottem
```

### Main Menu

Upon running the application, you will be presented with the main menu. From here, you can choose to:

- Start a new chat
- Continue a previous chat
- View API keys
- Exit the application

### Settings

Before using the application, you need to set up your API keys for the AI services you want to use. You can do this through the settings menu, which can be accessed from the main menu.

In the settings menu, you can:

- Set the Claude API key
- Set the OpenAI API key
- Set other API keys

Make sure to obtain the necessary API keys from the respective service providers and enter them accurately.

### Editor

When starting a new chat or continuing a previous one, you will enter the editor mode. The editor provides a full-screen text editing interface where you can compose your queries and view the responses from the AI.

#### Editor Controls

- Ctrl+E: Send the current query to the selected API
- Ctrl+J: Select the API to send the query to
- Ctrl+Q: Quit the editor and return to the main menu
- Arrow Keys: Move the cursor around the text
- Enter: Insert a new line
- Backspace: Delete the character before the cursor

#### Editor Status Bar

The editor status bar at the bottom of the screen provides useful information:

- Current API: Displays the name of the currently selected API
- Cursor Position: Shows the current line and column position of the cursor
- Status Message: Displays relevant status messages and prompts

### Chat History

You can view the history of your chats from the main menu. When viewing chat history, you will be prompted to select a specific chat. Once selected, the full history of the chat will be displayed, showing the queries and responses along with their respective timestamps and API names.

## Configuration

The application stores its configuration and data in the following directory:
```
~/.config/gottem/
```

This directory contains:

- `gottem.db`: The SQLite database file that stores chat history and API keys
- `logs/`: A directory containing log files for debugging purposes

## Dependencies

The Gottem CLI relies on the following dependencies:

- [go-sqlite3](https://github.com/mattn/go-sqlite3): SQLite driver for Go
- [promptui](https://github.com/manifoldco/promptui): Interactive prompt library for Go
- [tcell](https://github.com/gdamore/tcell): Terminal UI library for Go

These dependencies will be automatically downloaded and installed when you build the application.

## Contributing

Contributions to the Gottem CLI are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on the [GitHub repository](https://github.com/Utility-Gods/gottem).

## License

The Gottem CLI is open-source software licensed under the [MIT License](https://opensource.org/licenses/MIT).
