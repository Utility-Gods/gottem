package menu

import (
	"fmt"
	"strings"

	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/manifoldco/promptui"
)

func SettingsMenu() {
	for {
		prompt := promptui.Select{
			Label: "Settings Menu",
			Items: []string{"API Keys", "View API Keys", "Delete API Key", "Back to Main Menu"},
		}

		_, result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result {
		case "API Keys":
			APIKeyMenu()
		case "View API Keys":
			ViewAPIKeys()
		case "Delete API Key":
			DeleteAPIKey()
		case "Back to Main Menu":
			return
		}
	}
}

func APIKeyMenu() {
	for {
		prompt := promptui.Select{
			Label: "API Key Settings",
			Items: []string{"Set Claude API Key", "Set OpenAI API Key", "Set Groq API Key", "Back to Settings"},
		}

		_, result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result {
		case "Set Claude API Key":
			setAPIKey("claude")
		case "Set OpenAI API Key":
			setAPIKey("openai")
		case "Set Groq API Key":
			setAPIKey("groq")
		case "Back to Settings":
			return
		}
	}
}

func ViewAPIKeys() {
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

func DeleteAPIKey() {
	apiKeys, err := db.GetAllAPIKeys()
	if err != nil {
		fmt.Printf("Error retrieving API keys: %v\n", err)
		return
	}

	if len(apiKeys) == 0 {
		fmt.Println("No API keys found.")
		return
	}

	var items []string
	for _, key := range apiKeys {
		items = append(items, key.APIName)
	}
	items = append(items, "Cancel")

	prompt := promptui.Select{
		Label: "Select API Key to delete",
		Items: items,
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	if result == "Cancel" {
		return
	}

	confirmPrompt := promptui.Prompt{
		Label:     fmt.Sprintf("Are you sure you want to delete the API key for %s? (y/n)", result),
		IsConfirm: true,
	}

	confirmResult, err := confirmPrompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	if strings.ToLower(confirmResult) == "y" {
		err = db.DeleteAPIKey(result)
		if err != nil {
			fmt.Printf("Error deleting API key: %v\n", err)
		} else {
			fmt.Printf("API key for %s deleted successfully.\n", result)
		}
	} else {
		fmt.Println("Deletion cancelled.")
	}
}

func setAPIKey(apiName string) {
	prompt := promptui.Prompt{
		Label: fmt.Sprintf("Enter %s API Key", apiName),
		Mask:  '*',
	}

	apiKey, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	if err := db.SetAPIKey(apiName, apiKey); err != nil {
		fmt.Printf("Failed to set API key: %v\n", err)
		return
	}

	fmt.Printf("%s API Key set successfully.\n", apiName)
}

func setOtherAPIKey() {
	namePrompt := promptui.Prompt{
		Label: "Enter API Name",
	}

	apiName, err := namePrompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	setAPIKey(apiName)
}

func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}
