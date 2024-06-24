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
	openAIAPIURL = "https://api.openai.com/v1/chat/completions"
)

// OpenAIAPI implements the APIHandler interface for OpenAI API
type OpenAIAPI struct {
	apiKey string
	client *http.Client
}

// NewOpenAIAPI creates a new instance of OpenAIAPI
func NewOpenAIAPI() (*OpenAIAPI, error) {
	apiKey, err := db.GetAPIKey("openai")
	if err != nil {
		return nil, fmt.Errorf("error getting OpenAI API key: %w", err)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not set. Please run setup")
	}

	return &OpenAIAPI{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (o *OpenAIAPI) HandleQuery(query string) string {
	if o.client == nil {
		log.Println("HTTP client is nil")
		return "Error: HTTP client not initialized"
	}

	// Create a new spinner
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Suffix = " Querying OpenAI API..."
	s.Start()

	// Defer stopping the spinner
	defer s.Stop()

	requestBody, err := json.Marshal(map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{"role": "user", "content": query},
		},
		"max_tokens": 1000,
	})
	if err != nil {
		log.Printf("Error creating request body: %v", err)
		return fmt.Sprintf("Error creating request body: %v", err)
	}

	req, err := http.NewRequest("POST", openAIAPIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return fmt.Sprintf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		log.Printf("Error making request to OpenAI API: %v", err)
		return fmt.Sprintf("Error making request to OpenAI API: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return fmt.Sprintf("Error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error from OpenAI API: %s", body)
		return fmt.Sprintf("Error from OpenAI API: %s", body)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error parsing response: %v", err)
		return fmt.Sprintf("Error parsing response: %v", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		log.Println("Unexpected response format from OpenAI API")
		return "Unexpected response format from OpenAI API"
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		log.Println("Unexpected choice format from OpenAI API")
		return "Unexpected choice format from OpenAI API"
	}

	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		log.Println("Unexpected message format from OpenAI API")
		return "Unexpected message format from OpenAI API"
	}

	content, ok := message["content"].(string)
	if !ok {
		log.Println("Unable to extract content from OpenAI API response")
		return "Unable to extract content from OpenAI API response"
	}

	return content
}
