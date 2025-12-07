package marketplace

import (
	"fmt"
	"time"
)

// Marketplace provides command sharing and discovery
type Marketplace struct {
	commands map[string]*SharedCommand
}

// SharedCommand represents a shared command
type SharedCommand struct {
	ID          string
	Name        string
	Command     string
	Description string
	Author      string
	Category    string
	Tags        []string
	Rating      float64
	Downloads   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewMarketplace creates a new marketplace
func NewMarketplace() *Marketplace {
	return &Marketplace{
		commands: make(map[string]*SharedCommand),
	}
}

// Publish publishes a command to the marketplace
func (m *Marketplace) Publish(cmd *SharedCommand) error {
	cmd.ID = generateID()
	cmd.CreatedAt = time.Now()
	cmd.UpdatedAt = time.Now()
	m.commands[cmd.ID] = cmd
	return nil
}

// Search searches for commands
func (m *Marketplace) Search(query string) []*SharedCommand {
	results := []*SharedCommand{}
	for _, cmd := range m.commands {
		if matchesQuery(cmd, query) {
			results = append(results, cmd)
		}
	}
	return results
}

// GetPopular returns popular commands
func (m *Marketplace) GetPopular(limit int) []*SharedCommand {
	// Sort by downloads and return top N
	popular := []*SharedCommand{}
	for _, cmd := range m.commands {
		popular = append(popular, cmd)
	}
	return popular[:min(limit, len(popular))]
}

// Rate rates a command
func (m *Marketplace) Rate(id string, rating float64) error {
	cmd := m.commands[id]
	if cmd == nil {
		return fmt.Errorf("command not found")
	}
	cmd.Rating = rating
	return nil
}

func matchesQuery(cmd *SharedCommand, query string) bool {
	// Simple search implementation
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
