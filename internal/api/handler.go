package api

import (
	"log"

	"github.com/Utility-Gods/gottem/pkg/types"
)

// GetAPIHandlers returns a map of all available API handlers
func GetAPIHandlers() map[string]types.API {
	handlers := make(map[string]types.API)

	claudeAPI, err := NewClaudeAPI()
	if err != nil {
		log.Printf("Failed to initialize Claude API: %v", err)
		handlers["c"] = types.API{Name: "Claude API (Not Configured)", Shortcut: "c", Handler: &ErrorAPI{Err: err}}
	} else {
		handlers["c"] = types.API{Name: "Claude API", Shortcut: "c", Handler: claudeAPI}
	}

	openAIAPI, err := NewOpenAIAPI()
	if err != nil {
		log.Printf("Failed to initialize OpenAI API: %v", err)
		handlers["o"] = types.API{Name: "OpenAI API (Not Configured)", Shortcut: "o", Handler: &ErrorAPI{Err: err}}
	} else {
		handlers["o"] = types.API{Name: "OpenAI API", Shortcut: "o", Handler: openAIAPI}
	}

	return handlers
}

// ErrorAPI is a placeholder API that returns an error message
type ErrorAPI struct {
	Err error
}

func (e *ErrorAPI) HandleQuery(query string) string {
	return "API not properly configured: " + e.Err.Error()
}
