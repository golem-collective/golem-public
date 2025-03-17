package main

import (
	"ai-agent-app/database"
	"ai-agent-app/handlers"
	"ai-agent-app/services"
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var scanner = bufio.NewScanner(os.Stdin)

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
}

func main() {
	fmt.Println("AI Agent Application")
	fmt.Println("-------------------")

	// Initialize database connection
	database.InitDB()
	defer database.CloseDB()

	// Create necessary tables
	if err := database.CreateAgentsTable(); err != nil {
		log.Fatalf("Failed to create agents table: %v", err)
	}
	if err := database.CreateChatHistoryTable(); err != nil {
		log.Fatalf("Failed to create chat history table: %v", err)
	}

	// For debugging - print the API key (remove in production)
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Println("Warning: OPENAI_API_KEY is not set")
	} else {
		log.Println("OPENAI_API_KEY is set")
	}

	// Start HTTP server in a goroutine
	go startHTTPServer()

	// Start console interface
	startConsoleInterface()
}

func startHTTPServer() {
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/agents", handlers.GetAgents).Methods("GET")
	api.HandleFunc("/agents/{agentID}/chat", handlers.ChatWithAgent).Methods("POST")
	api.HandleFunc("/agents/{agentID}/history", handlers.ClearAgentHistory).Methods("DELETE")

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("Starting HTTP server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Set up graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal
	<-c

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait until the timeout
	srv.Shutdown(ctx)
	log.Println("HTTP server shutdown gracefully")
}

func startConsoleInterface() {
	// Initialize chat history service
	chatHistory := services.NewChatHistory(10) // Keep last 10 messages

	// Prompt for agent name
	agentName := promptForAgentName()

	// Create default agent with the provided name
	agentID, err := handlers.GetOrCreateDefaultAgent(agentName)
	if err != nil {
		log.Fatalf("Failed to create default agent: %v", err)
	}

	log.Printf("Created agent with ID: %d and name: %s", agentID, agentName)

	fmt.Println("Start chatting with the agent (type 'exit' to quit, 'clear' to clear history):")
	fmt.Println("API server is running in the background.")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		userInput := scanner.Text()
		if strings.ToLower(userInput) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if strings.ToLower(userInput) == "clear" {
			chatHistory.ClearHistory(agentID)
			fmt.Println("Chat history cleared.")
			continue
		}

		// Chat with the agent - the handler will manage the chat history
		response, err := handlers.ConsoleChatWithAgent(agentID, userInput, chatHistory)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Agent: %s\n", response)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}
}

// promptForAgentName asks the user to input a name for the agent
func promptForAgentName() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter a name for your agent (or press Enter for default 'Console Agent'): ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading input: %v. Using default name.", err)
		return ""
	}

	// Trim whitespace and newlines
	name = strings.TrimSpace(name)

	return name
}
