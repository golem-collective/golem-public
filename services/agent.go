// services/agent.go
package services

import (
	"ai-agent-app/database"
	"ai-agent-app/models"
	"fmt"
	"log"
)

// CreateAgent saves a new agent to the database and returns its ID
func CreateAgent(agent *models.Agent) error {
	// Prepare the SQL statement with RETURNING clause to get the generated ID
	query := `INSERT INTO agents (name) VALUES ($1) RETURNING id`
	err := database.GetDB().QueryRow(query,
		agent.Name,
	).Scan(&agent.ID)

	if err != nil {
		log.Printf("Error saving agent to database: %v", err)
		return err
	}
	return nil
}

// GetAgentByID retrieves an agent from the database by its ID
func GetAgentByID(id int) (*models.Agent, error) {
	query := `SELECT id, name FROM agents WHERE id = $1`

	var agent models.Agent
	err := database.GetDB().QueryRow(query, id).Scan(
		&agent.ID,
		&agent.Name,
	)

	if err != nil {
		log.Printf("Error retrieving agent with ID %d: %v", id, err)
		return nil, err
	}

	return &agent, nil
}

func GetAgentByName(name string) (*models.Agent, error) {
	query := `SELECT id, name FROM agents WHERE name = $1`

	var agent models.Agent
	err := database.GetDB().QueryRow(query, name).Scan(
		&agent.ID,
		&agent.Name,
	)

	if err != nil {
		log.Printf("Error retrieving agent with name %s: %v", name, err)
		return nil, err
	}

	return &agent, nil
}

// GetAllAgents returns all agents from the database
func GetAllAgents() ([]models.Agent, error) {
	query := `SELECT id, name FROM agents ORDER BY id`

	db := database.GetDB()
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying agents: %w", err)
	}
	defer rows.Close()

	var agents []models.Agent
	for rows.Next() {
		var agent models.Agent
		if err := rows.Scan(&agent.ID, &agent.Name); err != nil {
			return nil, fmt.Errorf("error scanning agent row: %w", err)
		}
		agents = append(agents, agent)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating agent rows: %w", err)
	}

	return agents, nil
}
