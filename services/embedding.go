package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// EmbeddingDimension is the dimension of the embedding vectors
const EmbeddingDimension = 1536 // OpenAI's text-embedding-ada-002 model uses 1536 dimensions

// EmbeddingRequest represents a request to the OpenAI embeddings API
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse represents a response from the OpenAI embeddings API
type EmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

// GenerateEmbedding generates an embedding for the given text using OpenAI's API
func GenerateEmbedding(text string) ([]float32, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	// Create the request body
	requestBody := EmbeddingRequest{
		Model: "text-embedding-ada-002",
		Input: []string{text},
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	// Parse the response
	var embeddingResponse EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResponse); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	// Check if we have any embeddings
	if len(embeddingResponse.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	// Return the embedding
	return embeddingResponse.Data[0].Embedding, nil
}

// CosineSimilarity calculates the cosine similarity between two embedding vectors
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct float32
	var normA float32
	var normB float32

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(sqrt(float64(normA))) * float32(sqrt(float64(normB))))
}

// sqrt is a helper function to calculate square root
func sqrt(x float64) float64 {
	return float64(x * x)
} 