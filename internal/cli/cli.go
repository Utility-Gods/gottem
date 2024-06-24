package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Utility-Gods/gottem/internal/api"
	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/manifoldco/promptui"
)

func RunCLI(app *api.App) {
	for {
		displayMainMenu()
		choice := getUserInput("Enter your choice: ")

		switch choice {
		case "1":
			startNewChat(app)
		case "2":
			continuePreviousChat(app)
		case "3":
			viewAPIKeys()
		case "4":
			return // Exit to main menu
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

func displayMainMenu() {
	fmt.Println("\n--- Multi-API CLI Menu ---")
	fmt.Println("1. Start a new chat")
	fmt.Println("2. Continue a previous chat")
	fmt.Println("3. View API keys")
	fmt.Println("4. Exit to main menu")
}

func getUserInput(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func startNewChat(app *api.App) {
	fmt.Println("\n--- New Chat ---")
	fmt.Println("Available APIs:")
	for _, api := range app.GetAvailableAPIs() {
		fmt.Printf("%s: %s\n", api.Shortcut, api.Name)
	}
	fmt.Println("Enter your query in the format: <API_SHORTCUT> <QUERY>")
	fmt.Println("To end the chat, type 'exit'")

	chatTitle := getUserInput("Enter a title for this chat: ")
	chatID, err := db.CreateChat(chatTitle)
	if err != nil {
		fmt.Printf("Error creating chat: %v\n", err)
		return
	}

	var messages []db.Message
	for {
		input := getUserInput("> ")
		if strings.ToLower(input) == "exit" {
			break
		}

		parts := strings.SplitN(input, " ", 2)
		if len(parts) != 2 {
			fmt.Println("Invalid input. Please use the format: <API_SHORTCUT> <QUERY>")
			continue
		}

		apiShortcut, query := parts[0], parts[1]
		response, err := app.HandleQuery(apiShortcut, query, chatID, messages)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		fmt.Println("Response:", response)

		// Add the new messages to the messages slice
		messages = append(messages,
			db.Message{Role: "user", APIName: apiShortcut, Content: query},
			db.Message{Role: "assistant", APIName: apiShortcut, Content: response},
		)
	}
}

func viewPreviousChats() {
	fmt.Println("\n--- Previous Chats ---")
	chats, err := db.GetChats()
	if err != nil {
		fmt.Printf("Error retrieving chats: %v\n", err)
		return
	}

	if len(chats) == 0 {
		fmt.Println("No previous chats found.")
		return
	}

	for _, chat := range chats {
		fmt.Printf("ID: %d, Title: %s, Created: %s\n", chat.ID, chat.Title, chat.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	chatIDStr := getUserInput("Enter a chat ID to view messages (or press Enter to go back): ")
	if chatIDStr == "" {
		return
	}

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		fmt.Println("Invalid chat ID.")
		return
	}

	messages, err := db.GetChatMessages(chatID)
	if err != nil {
		fmt.Printf("Error retrieving messages: %v\n", err)
		return
	}

	fmt.Printf("\n--- Messages for Chat ID %d ---\n", chatID)
	for _, msg := range messages {
		fmt.Printf("[%s] %s: %s\n", msg.CreatedAt.Format("2006-01-02 15:04:05"), msg.Role, msg.Content)
	}
}

func viewAPIKeys() {
	fmt.Println("\n--- API Keys ---")
	apiKeys, err := db.GetAllAPIKeys()
	if err != nil {
		fmt.Printf("Error retrieving API keys: %v\n", err)
		return
	}

	if len(apiKeys) == 0 {
		fmt.Println("No API keys found.")
		return
	}

	for _, key := range apiKeys {
		maskedKey := maskAPIKey(key.APIKey)
		fmt.Printf("API Name: %s\nAPI Key: %s\n\n", key.APIName, maskedKey)
	}
}

func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}

func continuePreviousChat(app *api.App) {
	chats, err := db.GetChats()
	if err != nil {
		fmt.Printf("Error retrieving chats: %v\n", err)
		return
	}

	if len(chats) == 0 {
		fmt.Println("No previous chats found.")
		return
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U0001F449 {{ .Title | cyan }} ({{ .CreatedAt | fdate }})",
		Inactive: "  {{ .Title | cyan }} ({{ .CreatedAt | fdate }})",
		Selected: "\U0001F449 {{ .Title | red | cyan }}",
		Details: `
--------- Chat ----------
{{ "ID:" | faint }}	{{ .ID }}
{{ "Title:" | faint }}	{{ .Title }}
{{ "Created:" | faint }}	{{ .CreatedAt | fdate }}
{{ "Updated:" | faint }}	{{ .UpdatedAt | fdate }}`,
	}

	funcMap := promptui.FuncMap
	funcMap["fdate"] = func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05")
	}

	prompt := promptui.Select{
		Label:     "Select a chat to continue",
		Items:     chats,
		Templates: templates,
		Size:      10,
		Searcher: func(input string, index int) bool {
			chat := chats[index]
			title := strings.Replace(strings.ToLower(chat.Title), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)
			return strings.Contains(title, input)
		},
	}

	index, _, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	selectedChat := chats[index]
	messages, err := db.GetChatMessages(selectedChat.ID)
	if err != nil {
		fmt.Printf("Error retrieving messages: %v\n", err)
		return
	}

	displayChatHistory(messages)
	continueChat(app, selectedChat.ID, messages)
}

func continueChat(app *api.App, chatID int, previousMessages []db.Message) {
	if previousMessages == nil {
		var err error
		previousMessages, err = db.GetChatMessages(chatID)
		if err != nil {
			fmt.Printf("Error retrieving previous messages: %v\n", err)
			return
		}
	}

	fmt.Println("\nContinue the conversation. Type 'exit' to end the chat, or 'edit' to edit the chat history.")

	for {
		input := getUserInput("> ")
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "exit" {
			break
		}

		if strings.ToLower(input) == "edit" {
			editedMessages, err := EditChatWithExternalEditor(previousMessages)
			if err != nil {
				fmt.Printf("Error editing chat: %v\n", err)
			} else {
				previousMessages = editedMessages
				fmt.Println("Chat history updated.")
				displayChatHistory(previousMessages)
			}
			continue
		}

		parts := strings.SplitN(input, " ", 2)
		if len(parts) != 2 {
			fmt.Println("Invalid input. Please use the format: <API_SHORTCUT> <QUERY>")
			continue
		}

		apiShortcut, query := parts[0], parts[1]
		response, err := app.HandleQuery(apiShortcut, query, chatID, previousMessages)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		fmt.Println("Response:", response)

		previousMessages = append(previousMessages,
			db.Message{Role: "user", APIName: apiShortcut, Content: query},
			db.Message{Role: "assistant", APIName: apiShortcut, Content: response},
		)
	}
}

func displayChatHistory(messages []db.Message) {
	fmt.Println("\n--- Chat History ---")
	for _, msg := range messages {
		fmt.Printf("[%s] %s (%s): %s\n\n",
			msg.CreatedAt.Format("2006-01-02 15:04:05"),
			msg.Role,
			msg.APIName,
			msg.Content,
		)
	}
	fmt.Println("--------------------")
}
