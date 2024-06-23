package api

import (
	"github.com/Utility-Gods/gottem/pkg/types"
)

// GetAPIHandlers returns a map of all available API handlers
func GetAPIHandlers() map[string]types.API {
	return map[string]types.API{
		"1": {Name: "Clause API", Shortcut: "1", Handler: &MockAPI{Name: "Clause API"}},
		"2": {Name: "Groq", Shortcut: "2", Handler: &MockAPI{Name: "Groq API"}},
		"w": {Name: "OpenAI API", Shortcut: "w", Handler: &OpenAiAPI{}},
		// Add more APIs here
	}
}
