package app

import (
	"fmt"

	"github.com/Utility-Gods/gottem/internal/api"
	"github.com/Utility-Gods/gottem/pkg/types"
)

// App represents the main application
type App struct {
	APIs map[string]types.API
}

// NewApp creates a new instance of the application
func NewApp() *App {
	return &App{
		APIs: api.GetAPIHandlers(),
	}
}

// HandleQuery processes a query for a specific API
func (a *App) HandleQuery(apiShortcut, query string) (string, error) {
	api, exists := a.APIs[apiShortcut]
	if !exists {
		return "", fmt.Errorf("no API found for shortcut '%s'", apiShortcut)
	}
	return api.Handler.HandleQuery(query), nil
}

// GetAvailableAPIs returns a list of available APIs
func (a *App) GetAvailableAPIs() []types.API {
	apis := make([]types.API, 0, len(a.APIs))
	for _, api := range a.APIs {
		apis = append(apis, api)
	}
	return apis
}
