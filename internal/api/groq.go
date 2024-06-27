package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/briandowns/spinner"
)

const (
	groqAPIURL = "https://api.groq.com/openai/v1/chat/completions"
)

// GroqAPI implements the APIHandler interface for Groq API
type GroqAPI struct {
	apiKey string
	client *http.Client
}

// NewGroqAPI creates a new instance of GroqAPI
func NewGroqAPI() (*GroqAPI, error) {
	apiKey, err := db.GetAPIKey("groq")
	if err != nil {
		return nil, fmt.Errorf("error getting Groq API key: %w", err)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("Groq API key not set. Please run setup")
	}

	return &GroqAPI{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *GroqAPI) HandleQuery(query string) string {
	if c.client == nil {
		// log.Println("HTTP client is nil")
		return "Error: HTTP client not initialized"
	}

	// Create a new spinner
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Suffix = " Querying Groq API..."
	s.Start()

	// Defer stopping the spinner
	defer s.Stop()

	// Prepare the messages for the API request
	messages := []map[string]string{
		{"role": "user", "content": query},
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"max_tokens": 1000,
		"messages":   messages,
	})
	if err != nil {
		// log.Printf("Error creating request body: %v", err)
		return fmt.Sprintf("Error creating request body: %v", err)
	}

	req, err := http.NewRequest("POST", groqAPIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		// log.Printf("Error creating request: %v", err)
		return fmt.Sprintf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		// log.Printf("Error making request to Groq API: %v", err)
		return fmt.Sprintf("Error making request to Groq API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// log.Printf("Error reading response body: %v", err)
		return fmt.Sprintf("Error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		// log.Printf("Error from Groq API: %s", body)
		return fmt.Sprintf("Error from Groq API: %s", body)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		// log.Printf("Error parsing response: %v", err)
		return fmt.Sprintf("Error parsing response: %v", err)
	}

	content, ok := result["content"].([]interface{})
	if !ok || len(content) == 0 {
		// log.Println("Unexpected response format from Groq API")
		return "Unexpected response format from Groq API"
	}

	firstContent, ok := content[0].(map[string]interface{})
	if !ok {
		// log.Println("Unexpected content format from Groq API")
		return "Unexpected content format from Groq API"
	}

	text, ok := firstContent["text"].(string)
	if !ok {
		// log.Println("Unable to extract text from Groq API response")
		return "Unable to extract text from Groq API response"
	}

	return text
}
