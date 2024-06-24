package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/briandowns/spinner"
)

const (
	claudeAPIURL = "https://api.anthropic.com/v1/messages"
)

// ClaudeAPI implements the APIHandler interface for Claude API
type ClaudeAPI struct {
	apiKey string
	client *http.Client
}

// NewClaudeAPI creates a new instance of ClaudeAPI
func NewClaudeAPI() (*ClaudeAPI, error) {
	apiKey, err := db.GetAPIKey("claude")
	if err != nil {
		return nil, fmt.Errorf("error getting Claude API key: %w", err)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("Claude API key not set. Please run setup")
	}

	return &ClaudeAPI{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *ClaudeAPI) HandleQuery(query string) string {
	if c.client == nil {
		log.Println("HTTP client is nil")
		return "Error: HTTP client not initialized"
	}

	// Create a new spinner
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Suffix = " Querying Claude API..."
	s.Start()

	// Defer stopping the spinner
	defer s.Stop()

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":      "claude-3-opus-20240229",
		"max_tokens": 1000,
		"messages": []map[string]string{
			{"role": "user", "content": query},
		},
	})
	if err != nil {
		log.Printf("Error creating request body: %v", err)
		return fmt.Sprintf("Error creating request body: %v", err)
	}

	req, err := http.NewRequest("POST", claudeAPIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return fmt.Sprintf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// log.Printf("Claude API config: %+v", c)
	// log.Printf("Request: %+v", req)

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Error making request to Claude API: %v", err)
		return fmt.Sprintf("Error making request to Claude API: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return fmt.Sprintf("Error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error from Claude API: %s", body)
		return fmt.Sprintf("Error from Claude API: %s", body)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error parsing response: %v", err)
		return fmt.Sprintf("Error parsing response: %v", err)
	}

	content, ok := result["content"].([]interface{})
	if !ok || len(content) == 0 {
		log.Println("Unexpected response format from Claude API")
		return "Unexpected response format from Claude API"
	}

	firstContent, ok := content[0].(map[string]interface{})
	if !ok {
		log.Println("Unexpected content format from Claude API")
		return "Unexpected content format from Claude API"
	}

	text, ok := firstContent["text"].(string)
	if !ok {
		log.Println("Unable to extract text from Claude API response")
		return "Unable to extract text from Claude API response"
	}

	return text
}
