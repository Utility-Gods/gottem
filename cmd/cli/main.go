package main

import (
	"log"

	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/Utility-Gods/gottem/internal/menu"
)

func main() {
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	menu.MainMenu()
}
