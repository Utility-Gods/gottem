package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Utility-Gods/gottem/internal/app"
)

// RunCLI starts the CLI interface
func RunCLI(app *app.App) {
	fmt.Println("Welcome to the Multi-API CLI!")
	fmt.Println("Available APIs:")
	for _, api := range app.GetAvailableAPIs() {
		fmt.Printf("%s: %s\n", api.Shortcut, api.Name)
	}
	fmt.Println("Enter your query in the format: <API_SHORTCUT> <QUERY>")
	fmt.Println("To quit, type 'exit'")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		scanner.Scan()
		input := scanner.Text()

		if strings.ToLower(input) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		parts := strings.SplitN(input, " ", 2)
		if len(parts) != 2 {
			fmt.Println("Invalid input. Please use the format: <API_SHORTCUT> <QUERY>")
			continue
		}

		apiShortcut, query := parts[0], parts[1]

		response, err := app.HandleQuery(apiShortcut, query)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		fmt.Println(response)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
