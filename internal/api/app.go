package api

import (
	"fmt"

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

	fullQuery := context + "\n\nHuman: " + query + "\n\nAssistant:"
	response := api.Handler.HandleQuery(fullQuery)

	return response, nil
}

// GetAvailableAPIs returns a list of available APIs
func (a *App) GetAvailableAPIs() []types.APIInfo {
	apis := make([]types.APIInfo, 0, len(a.APIs))
	for _, api := range a.APIs {
		apis = append(apis, api)
	}
	return apis
}
