package menu

import (
	"fmt"

	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/manifoldco/promptui"
)

func SettingsMenu() {
	for {
		prompt := promptui.Select{
			Label: "Settings Menu",
			Items: []string{"API Keys", "Back to Main Menu"},
		}

		_, result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed: %v\n", err)
			return
		}

		switch result {
		case "API Keys":
			APIKeyMenu()
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
			fmt.Printf("Prompt failed: %v\n", err)
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

func setAPIKey(apiName string) {
	prompt := promptui.Prompt{
		Label: fmt.Sprintf("Enter %s API Key", apiName),
		Mask:  '*',
	}

	apiKey, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
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
		fmt.Printf("Prompt failed: %v\n", err)
		return
	}

	setAPIKey(apiName)
}
