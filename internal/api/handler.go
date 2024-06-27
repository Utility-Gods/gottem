package api

import (
	"log"

	"github.com/Utility-Gods/gottem/pkg/types"
)

// GetAPIHandlers returns a map of all available API handlers
func GetAPIHandlers() map[string]types.APIInfo {
	handlers := make(map[string]types.APIInfo)

	claudeAPI, err := NewClaudeAPI()
	if err != nil {
		log.Printf("Failed to initialize Claude API: %v", err)
		handlers["c"] = types.APIInfo{Name: "Claude API (Not Configured)", Shortcut: "c", Handler: &ErrorAPI{Err: err}}
	} else {
		handlers["c"] = types.APIInfo{Name: "Claude API", Shortcut: "c", Handler: claudeAPI}
	}

	openAIAPI, err := NewOpenAIAPI()
	if err != nil {
		log.Printf("Failed to initialize OpenAI API: %v", err)
		handlers["o"] = types.APIInfo{Name: "OpenAI API (Not Configured)", Shortcut: "o", Handler: &ErrorAPI{Err: err}}
	} else {
		handlers["o"] = types.APIInfo{Name: "OpenAI API", Shortcut: "o", Handler: openAIAPI}
	}

	groqAPI, err := NewGroqAPI()
	if err != nil {
		log.Printf("Failed to initialize Groq API: %v", err)
		handlers["g"] = types.APIInfo{Name: "Groq API (Not Configured)", Shortcut: "g", Handler: &ErrorAPI{Err: err}}
	} else {
		handlers["g"] = types.APIInfo{Name: "Groq API", Shortcut: "g", Handler: groqAPI}
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
