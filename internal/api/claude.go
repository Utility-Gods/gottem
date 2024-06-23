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
		return nil, fmt.Errorf("CLAUDE_API_KEY environment variable is not set")
	}

	log.Println("Claude API key loaded successfully", apiKey)

	return &ClaudeAPI{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *ClaudeAPI) HandleQuery(query string) string {
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
		return fmt.Sprintf("Error creating request body: %v", err)
	}

	req, err := http.NewRequest("POST", claudeAPIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Sprintf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Sprintf("Error making request to Claude API: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Error from Claude API: %s", body)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Sprintf("Error parsing response: %v", err)
	}

	content, ok := result["content"].([]interface{})
	if !ok || len(content) == 0 {
		return "Unexpected response format from Claude API"
	}

	firstContent, ok := content[0].(map[string]interface{})
	if !ok {
		return "Unexpected content format from Claude API"
	}

	text, ok := firstContent["text"].(string)
	if !ok {
		return "Unable to extract text from Claude API response"
	}

	return text
}
