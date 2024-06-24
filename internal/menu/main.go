package menu

import (
	"fmt"

	"github.com/Utility-Gods/gottem/internal/api"
	"github.com/Utility-Gods/gottem/internal/cli"
	"github.com/manifoldco/promptui"
)

// MainMenu starts the main menu loop
func MainMenu() {
	for {
		prompt := promptui.Select{
			Label: "Main Menu",
			Items: []string{"Run CLI", "Settings", "Exit"},
		}

		_, result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed: %v\n", err)
			return
		}

		switch result {
		case "Run CLI":
			app := api.NewApp()
			cli.RunCLI(app)
		case "Settings":
			SettingsMenu()
		case "Exit":
			fmt.Println("Goodbye!")
			return
		}
	}
}
