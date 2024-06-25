package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/Utility-Gods/gottem/internal/api"
	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/manifoldco/promptui"
)

func RunCLI(app *api.App) {
	for {
		choice := displayMainMenu()

		switch choice {
		case "Start a new chat":
			startNewChat(app)
		case "Continue a previous chat":
			continuePreviousChat(app)
		case "View chat history":
			viewChatHistory()
		case "View API keys":
			viewAPIKeys()
		case "Exit":
			fmt.Println("Goodbye!")
			return
		}
	}
}

func displayMainMenu() string {
	prompt := promptui.Select{
		Label: "Select an option",
		Items: []string{"Start a new chat", "Continue a previous chat", "View chat history", "View API keys", "Exit"},
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return ""
	}

	return result
}

func startNewChat(app *api.App) {
	prompt := promptui.Prompt{
		Label: "Enter a title for the new chat",
	}

	title, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	chatID, err := db.CreateChat(title)
	if err != nil {
		fmt.Printf("Error creating chat: %v\n", err)
		return
	}

	editor, err := NewVimEditor(app, chatID, nil)
	if err != nil {
		fmt.Printf("Error creating Vim editor: %v\n", err)
		return
	}

	if err := editor.Run(); err != nil {
		fmt.Printf("Error running Vim editor: %v\n", err)
	}
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

	selectedChat, err := selectChat(chats)
	if err != nil {
		fmt.Printf("Error selecting chat: %v\n", err)
		return
	}

	messages, err := db.GetChatMessages(selectedChat.ID)
	if err != nil {
		fmt.Printf("Error retrieving messages: %v\n", err)
		return
	}

	editor, err := NewVimEditor(app, selectedChat.ID, messages)
	if err != nil {
		fmt.Printf("Error creating Vim editor: %v\n", err)
		return
	}

	if err := editor.Run(); err != nil {
		fmt.Printf("Error running Vim editor: %v\n", err)
	}
}

func viewChatHistory() {
	chats, err := db.GetChats()
	if err != nil {
		fmt.Printf("Error retrieving chats: %v\n", err)
		return
	}

	if len(chats) == 0 {
		fmt.Println("No chats found.")
		return
	}

	selectedChat, err := selectChat(chats)
	if err != nil {
		fmt.Printf("Error selecting chat: %v\n", err)
		return
	}

	messages, err := db.GetChatMessages(selectedChat.ID)
	if err != nil {
		fmt.Printf("Error retrieving messages: %v\n", err)
		return
	}

	fmt.Printf("\n--- Chat History for '%s' ---\n", selectedChat.Title)
	for _, msg := range messages {
		fmt.Printf("[%s] %s (%s): %s\n\n",
			msg.CreatedAt.Format("2006-01-02 15:04:05"),
			msg.Role,
			msg.APIName,
			msg.Content,
		)
	}
	fmt.Println("--- End of Chat History ---")

	prompt := promptui.Prompt{
		Label:     "Press Enter to continue",
		IsConfirm: true,
	}
	_, err = prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
	}
}

func viewAPIKeys() {
	apiKeys, err := db.GetAllAPIKeys()
	if err != nil {
		fmt.Printf("Error retrieving API keys: %v\n", err)
		return
	}

	if len(apiKeys) == 0 {
		fmt.Println("No API keys found.")
		return
	}

	fmt.Println("\n--- API Keys ---")
	for _, key := range apiKeys {
		maskedKey := maskAPIKey(key.APIKey)
		fmt.Printf("API Name: %s\nAPI Key: %s\n\n", key.APIName, maskedKey)
	}

	prompt := promptui.Prompt{
		Label:     "Press Enter to continue",
		IsConfirm: true,
	}
	_, err = prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
	}
}

func selectChat(chats []db.Chat) (db.Chat, error) {
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
		Label:     "Select a chat",
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
		return db.Chat{}, err
	}

	return chats[index], nil
}

func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}
