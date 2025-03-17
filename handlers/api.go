package handlers

import (
	"ai-agent-app/services"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// ChatRequest represents the structure of a chat request
type ChatRequest struct {
	Message string `json:"message"`
}

// WebChatHistory is a global chat history for web requests
var WebChatHistory = services.NewChatHistory(10)

// ChatWithAgent handles API chat requests with the agent
func ChatWithAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentIDStr := vars["agentID"]

	// Validate and convert agentID to integer
	if agentIDStr == "" {
		http.Error(w, "agentID is required", http.StatusBadRequest)
		return
	}

	agentID, err := strconv.Atoi(agentIDStr)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	// Extract the message from the request body
	var requestBody ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		log.Printf("Error decoding request body: %v", err)
		return
	}

	// Validate the message
	if requestBody.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	// Use the same function as the console chat
	responseMessage, err := ConsoleChatWithAgent(agentID, requestBody.Message, WebChatHistory)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error communicating with agent: %v", err), http.StatusInternalServerError)
		return
	}

	// Log the API chat request
	log.Printf("API chat request for agentID: %d, message: %s", agentID, requestBody.Message)

	// Send response
	response := ChatResponse{
		Message: responseMessage,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAgents returns a list of all available agents
func GetAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := services.GetAllAgents()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving agents: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// ClearAgentHistory clears the chat history for a specific agent
func ClearAgentHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentIDStr := vars["agentID"]

	agentID, err := strconv.Atoi(agentIDStr)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	WebChatHistory.ClearHistory(agentID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Chat history cleared"})
}
