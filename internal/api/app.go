package api

import (
	"fmt"

	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/Utility-Gods/gottem/pkg/types"
)

// App represents the main application structure
type App struct {
	APIs map[string]types.APIInfo
}

func NewApp() *App {
	return &App{
		APIs: GetAPIHandlers(),
	}
}

// HandleQuery processes a query for a specific API
func (a *App) HandleQuery(apiShortcut, query string, chatID int, context string) (string, error) {
	api, exists := a.APIs[apiShortcut]
	if !exists {
		return "", fmt.Errorf("no API found for shortcut '%s'", apiShortcut)
	}

	// Prepare full query with context
	fullQuery := context + "\n\nHuman: " + query + "\n\nAssistant:"
	response := api.Handler.HandleQuery(fullQuery)

	return response, nil
}

// prepareContext creates a context string from previous messages
func prepareContext(messages []db.Message) string {
	var context string
	for _, msg := range messages {
		if msg.Role == "user" {
			context += "Human: " + msg.Content + "\n\n"
		} else if msg.Role == "assistant" {
			context += "Assistant: " + msg.Content + "\n\n"
		}
	}
	return context
}

// GetAvailableAPIs returns a list of available APIs
func (a *App) GetAvailableAPIs() []types.APIInfo {
	apis := make([]types.APIInfo, 0, len(a.APIs))
	for _, api := range a.APIs {
		apis = append(apis, api)
	}
	return apis
}
