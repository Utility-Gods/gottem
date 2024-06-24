package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Utility-Gods/gottem/internal/app"
	"github.com/Utility-Gods/gottem/internal/cli"
	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/Utility-Gods/gottem/internal/setup"
	"github.com/manifoldco/promptui"
)

func main() {
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	log.Print("Database initialized")

	prompt := promptui.Select{
		Label: "Select action",
		Items: []string{"Run CLI", "Setup", "Exit"},
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		os.Exit(1)
	}

	switch result {
	case "Run CLI":
		myApp := app.NewApp()
		cli.RunCLI(myApp)
	case "Setup":
		if err := setup.RunSetup(); err != nil {
			fmt.Printf("Setup failed: %v\n", err)
			os.Exit(1)
		}
	case "Exit":
		fmt.Println("Goodbye!")
	}
}
