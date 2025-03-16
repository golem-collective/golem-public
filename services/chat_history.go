package services

import (
	"ai-agent-app/database"
	"encoding/json"
	"fmt"
	"log"
)

// Message represents a single message in the chat history
type Message struct {
	ID        int       `json:"id"`
	Role      string    `json:"role"`    // "user" or "assistant"
	Content   string    `json:"content"` // The message content
	Embedding []float32 `json:"-"`       // The embedding vector (not included in JSON)
}

// ChatHistory stores conversation history for each agent
type ChatHistory struct {
	contextSize int // Number of messages to include in context
}

// NewChatHistory creates a new chat history manager
func NewChatHistory(contextSize int) *ChatHistory {
	return &ChatHistory{
		contextSize: contextSize,
	}
}

// AddMessage adds a message to the conversation history for a specific agent
func (ch *ChatHistory) AddMessage(agentID int, role, content string) error {
	// Generate embedding for the message
	embedding, err := GenerateEmbedding(content)
	if err != nil {
		log.Printf("Warning: Could not generate embedding for message: %v", err)
		// Continue without embedding
	}

	// Convert embedding to JSON for storage
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		log.Printf("Warning: Could not marshal embedding: %v", err)
		// Continue without embedding
	}

	// Insert the message with embedding
	query := `
		INSERT INTO chat_history (agent_id, role, content, embedding)
		VALUES ($1, $2, $3, $4)`

	_, err = database.Exec(query, agentID, role, content, embeddingJSON)
	if err != nil {
		log.Printf("Error adding message to chat history: %v", err)
		return err
	}

	return nil
}

// GetHistory returns the conversation history for a specific agent
// Limited to the most recent contextSize messages for context building
func (ch *ChatHistory) GetHistory(agentID int) []Message {
	query := `
		SELECT id, role, content 
		FROM chat_history 
		WHERE agent_id = $1 
		ORDER BY created_at DESC
		LIMIT $2`

	db := database.GetDB()
	rows, err := db.Query(query, agentID, ch.contextSize)
	if err != nil {
		log.Printf("Error getting chat history: %v", err)
		return []Message{}
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Role, &msg.Content); err != nil {
			log.Printf("Error scanning chat history row: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating chat history rows: %v", err)
	}

	// Reverse the messages to get chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages
}

// GetFullHistory returns the complete conversation history for a specific agent
func (ch *ChatHistory) GetFullHistory(agentID int) []Message {
	query := `
		SELECT id, role, content 
		FROM chat_history 
		WHERE agent_id = $1 
		ORDER BY created_at ASC`

	db := database.GetDB()
	rows, err := db.Query(query, agentID)
	if err != nil {
		log.Printf("Error getting full chat history: %v", err)
		return []Message{}
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Role, &msg.Content); err != nil {
			log.Printf("Error scanning chat history row: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating chat history rows: %v", err)
	}

	return messages
}

// SearchSimilarMessages finds messages similar to the query using embeddings
func (ch *ChatHistory) SearchSimilarMessages(agentID int, query string, limit int) ([]Message, error) {
	// Generate embedding for the query
	queryEmbedding, err := GenerateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("error generating embedding for query: %v", err)
	}

	// Convert embedding to JSON for the query
	queryEmbeddingJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("error marshaling query embedding: %v", err)
	}

	// Search for similar messages using cosine similarity
	sqlQuery := `
		SELECT id, role, content, embedding, embedding <=> $1 AS similarity
		FROM chat_history
		WHERE agent_id = $2 AND embedding IS NOT NULL
		ORDER BY similarity ASC
		LIMIT $3`

	db := database.GetDB()
	rows, err := db.Query(sqlQuery, queryEmbeddingJSON, agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("error searching similar messages: %v", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var similarity float32
		var embeddingJSON []byte

		if err := rows.Scan(&msg.ID, &msg.Role, &msg.Content, &embeddingJSON, &similarity); err != nil {
			log.Printf("Error scanning search result: %v", err)
			continue
		}

		// Print the message ID to the console
		log.Printf("Found similar message with ID: %d, similarity: %f", msg.ID, similarity)

		// Only try to unmarshal if we have embedding data
		if len(embeddingJSON) > 0 {
			if err := json.Unmarshal(embeddingJSON, &msg.Embedding); err != nil {
				log.Printf("Warning: Could not unmarshal embedding: %v", err)
			}
		}

		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search results: %v", err)
	}

	return messages, nil
}

// ClearHistory clears the conversation history for a specific agent
func (ch *ChatHistory) ClearHistory(agentID int) {
	query := `DELETE FROM chat_history WHERE agent_id = $1`
	_, err := database.Exec(query, agentID)
	if err != nil {
		log.Printf("Error clearing chat history: %v", err)
	}
}
