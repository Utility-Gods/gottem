package api

import (
	"github.com/Utility-Gods/gottem/pkg/types"
)

// GetAPIHandlers returns a map of all available API handlers
func GetAPIHandlers() map[string]types.API {
	return map[string]types.API{
		"1": {Name: "Claude API", Shortcut: "1", Handler: &ClaudeAPI{}},
		"2": {Name: "Groq", Shortcut: "2", Handler: &ClaudeAPI{}},
		"w": {Name: "OpenAI API", Shortcut: "w", Handler: &OpenAiAPI{}},
		// Add more APIs here
	}
}
