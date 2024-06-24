package api

import (
	"log"

	"github.com/Utility-Gods/gottem/pkg/types"
)

// GetAPIHandlers returns a map of all available API handlers
func GetAPIHandlers() map[string]types.API {
	claudeAPI, err := NewClaudeAPI()
	if err != nil {
		log.Printf("Failed to initialize Claude API: %v", err)
		return map[string]types.API{
			"c": {Name: "Claude API (Not Configured)", Shortcut: "c", Handler: &ErrorAPI{Err: err}},
		}
	}

	return map[string]types.API{
		"c": {Name: "Claude API", Shortcut: "c", Handler: claudeAPI},
		// You can keep other APIs if needed
	}
}

// ErrorAPI is a placeholder API that returns an error message
type ErrorAPI struct {
	Err error
}

func (e *ErrorAPI) HandleQuery(query string) string {
	return "API not properly configured: " + e.Err.Error()
}
