package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("Welcome to the Uppercase Converter!")
	fmt.Println("Type a string and press Enter. To quit, type 'exit' or press Ctrl+C.")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		scanner.Scan()
		input := scanner.Text()

		if strings.ToLower(input) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		uppercase := strings.ToUpper(input)
		fmt.Println(uppercase)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
