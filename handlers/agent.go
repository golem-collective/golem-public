package handlers

import (
	"ai-agent-app/models"
	"ai-agent-app/services"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// CreateAgentResponse represents the structure of the create agent response
type CreateAgentResponse struct {
	Agent models.Agent `json:"agent"`
}

// CreateAgent handles the creation of a new agent
func CreateAgent(w http.ResponseWriter, r *http.Request) {
	var agent models.Agent
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		log.Printf("Error decoding request body: %v", err)
		return
	}

	// Basic validation (you can expand this as needed)
	if agent.Name == "" {
		http.Error(w, "Agent name is required", http.StatusBadRequest)
		return
	}

	// Call the service to save the agent to the database
	if err := services.CreateAgent(&agent); err != nil {
		http.Error(w, "Error saving agent to database", http.StatusInternalServerError)
		log.Printf("Error saving agent: %v", err)
		return
	}

	// Log the creation of the agent
	log.Printf("Agent created: %+v", agent)

	// Send response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateAgentResponse{Agent: agent})
}

// CreateDefaultAgent creates a default agent and returns its ID
func CreateDefaultAgent(agentName string) (int, error) {
	// Create a default agent
	agent := models.Agent{
		Name: agentName,
	}

	// Call the service to save the agent to the database
	if err := services.CreateAgent(&agent); err != nil {
		return 0, fmt.Errorf("failed to create default agent: %w", err)
	}

	return agent.ID, nil
}

func GetOrCreateDefaultAgent(agentName string) (int, error) {
	// Check if the agent already exists
	existingAgent, err := services.GetAgentByName(agentName)
	if err == nil {
		return existingAgent.ID, nil
	}

	// If the agent doesn't exist, create a new one
	return CreateDefaultAgent(agentName)
}
