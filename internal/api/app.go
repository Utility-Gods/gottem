package api

import (
	"fmt"

	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/Utility-Gods/gottem/pkg/types"
)

// App represents the main application structure
type App struct {
	APIs map[string]types.API
}

// NewApp creates and initializes a new App instance
func NewApp() *App {
	return &App{
		APIs: GetAPIHandlers(),
	}
}

// HandleQuery processes a query for a specific API
func (a *App) HandleQuery(apiShortcut, query string, chatID int) (string, error) {
	api, exists := a.APIs[apiShortcut]
	if !exists {
		return "", fmt.Errorf("no API found for shortcut '%s'", apiShortcut)
	}

	response := api.Handler.HandleQuery(query)

	// Save the user message
	err := db.AddMessage(chatID, "user", apiShortcut, query)
	if err != nil {
		return "", fmt.Errorf("failed to save user message: %w", err)
	}

	// Save the assistant message
	err = db.AddMessage(chatID, "assistant", apiShortcut, response)
	if err != nil {
		return "", fmt.Errorf("failed to save assistant message: %w", err)
	}

	return response, nil
}

// GetAvailableAPIs returns a list of available APIs
func (a *App) GetAvailableAPIs() []types.API {
	apis := make([]types.API, 0, len(a.APIs))
	for _, api := range a.APIs {
		apis = append(apis, api)
	}
	return apis
}
