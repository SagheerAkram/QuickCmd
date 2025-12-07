package refinement

import (
	"fmt"
	"strings"
)

// RefinementLoop provides interactive command refinement
type RefinementLoop struct {
	history []string
}

// NewRefinementLoop creates a new refinement loop
func NewRefinementLoop() *RefinementLoop {
	return &RefinementLoop{
		history: []string{},
	}
}

// Refine refines a command based on user feedback
func (rl *RefinementLoop) Refine(original, feedback string) *RefinementSuggestion {
	rl.history = append(rl.history, original)
	
	suggestion := &RefinementSuggestion{
		Original:    original,
		Refined:     applyFeedback(original, feedback),
		Explanation: fmt.Sprintf("Applied: %s", feedback),
		Options:     []string{},
	}
	
	// Generate options
	suggestion.Options = append(suggestion.Options,
		"Add --exclude for node_modules",
		"Change to find only in src/ folder",
		"Show file sizes in human-readable format",
	)
	
	return suggestion
}

// RefinementSuggestion represents a refinement suggestion
type RefinementSuggestion struct {
	Original    string
	Refined     string
	Explanation string
	Options     []string
}

func applyFeedback(command, feedback string) string {
	// Simplified feedback application
	if strings.Contains(feedback, "exclude") {
		return command + " --exclude node_modules"
	}
	return command
}

// InteractiveSession manages an interactive refinement session
type InteractiveSession struct {
	turns []Turn
}

// Turn represents one turn in the conversation
type Turn struct {
	UserInput string
	Response  string
}

// AddTurn adds a turn to the session
func (is *InteractiveSession) AddTurn(userInput, response string) {
	is.turns = append(is.turns, Turn{
		UserInput: userInput,
		Response:  response,
	})
}
