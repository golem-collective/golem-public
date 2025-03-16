package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// OpenAIAPIURL is the endpoint for the OpenAI API
const OpenAIAPIURL = "https://api.openai.com/v1/chat/completions"

// OpenAIRequest represents the structure of a request to the OpenAI API
type OpenAIRequest struct {
	Model   string `json:"model"`
	Context string `json:"context"`
}

// OpenAIResponse represents the structure of a response from the OpenAI API
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Add this at the package level
var (
	httpClient = &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
)

// SendMessageToOpenAI sends a message to the OpenAI API and returns the response
func SendMessageToOpenAI(apiKey, message, template string, state map[string]string) (string, error) {
	// Start timing
	startTime := time.Now()
	log.Printf("Starting OpenAI API request...")

	// Process the template with the state
	context := template
	for key, value := range state {
		context = strings.Replace(context, "{{"+key+"}}", value, -1)
	}

	// Combine the context and message
	prompt := context + message

	// Create the request payload
	requestBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	}

	// Convert the request payload to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Set the headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the response
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	// Check if there are any choices
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	// Calculate and log the elapsed time
	elapsedTime := time.Since(startTime)
	log.Printf("OpenAI API request completed in %v", elapsedTime)

	// Return the content of the first choice
	return response.Choices[0].Message.Content, nil
}

// AddMessage is a helper function to add a message to the history
func AddMessage(agentID, role, content string) {
	// This function would typically store the message in a database
	// For now, we'll just log it
	fmt.Printf("Adding message to history for agent %s: %s: %s\n", agentID, role, content)
}
