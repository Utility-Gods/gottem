package setup

import (
	"fmt"

	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/manifoldco/promptui"
)

func RunSetup() error {
	for {
		prompt := promptui.Select{
			Label: "Select action",
			Items: []string{"Set Claude API Key", "Set OpenAI API Key", "Set Groq API Key", "Exit Setup"},
		}

		_, result, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		switch result {
		case "Set Claude API Key":
			if err := setAPIKey("claude"); err != nil {
				return err
			}
		case "Set OpenAI API Key":
			if err := setAPIKey("openai"); err != nil {
				return err
			}
		case "Set Groq API Key":
			if err := setAPIKey("groq"); err != nil {
				return err
			}
		case "Exit Setup":
			return nil
		}
	}
}

func setAPIKey(apiName string) error {
	prompt := promptui.Prompt{
		Label: fmt.Sprintf("Enter %s API Key", apiName),
		Mask:  '*',
	}

	apiKey, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	if err := db.SetAPIKey(apiName, apiKey); err != nil {
		return fmt.Errorf("failed to set API key: %w", err)
	}

	fmt.Printf("%s API Key set successfully.\n", apiName)
	return nil
}

func setOtherAPIKey() error {
	namePrompt := promptui.Prompt{
		Label: "Enter API Name",
	}

	apiName, err := namePrompt.Run()
	if err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	return setAPIKey(apiName)
}
