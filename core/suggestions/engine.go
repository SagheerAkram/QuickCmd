package suggestions

import (
	"fmt"
	"strings"
	"time"

	"github.com/SagheerAkram/QuickCmd/core/translator"
)

// SuggestionEngine analyzes command patterns and provides intelligent suggestions
type SuggestionEngine struct {
	patterns map[string]*CommandPattern
	prefs    map[string]*UserPreferences
}

// CommandPattern represents a detected command pattern
type CommandPattern struct {
	Hash          string
	BaseCommand   string
	Frequency     int
	LastSeen      time.Time
	Variations    []string
	SuggestedAlias string
	Optimization  string
}

// UserPreferences tracks user-specific preferences
type UserPreferences struct {
	UserID              string
	PreferredDirectories []string
	CommonFlags         map[string]int
	IgnoredSuggestions  []string
	CommandHistory      []string
}

// Suggestion represents a command suggestion
type Suggestion struct {
	Type        string `json:"type"` // "alias", "optimization", "pattern", "context"
	Title       string `json:"title"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
	Confidence  int    `json:"confidence"`
	Reason      string `json:"reason"`
}

// NewSuggestionEngine creates a new suggestion engine
func NewSuggestionEngine() *SuggestionEngine {
	return &SuggestionEngine{
		patterns: make(map[string]*CommandPattern),
		prefs:    make(map[string]*UserPreferences),
	}
}

// AnalyzeCommand analyzes a command and updates patterns
func (se *SuggestionEngine) AnalyzeCommand(userID, command string) {
	// Get or create user preferences
	if se.prefs[userID] == nil {
		se.prefs[userID] = &UserPreferences{
			UserID:      userID,
			CommonFlags: make(map[string]int),
		}
	}
	prefs := se.prefs[userID]

	// Add to history
	prefs.CommandHistory = append(prefs.CommandHistory, command)
	if len(prefs.CommandHistory) > 100 {
		prefs.CommandHistory = prefs.CommandHistory[1:]
	}

	// Extract base command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}
	baseCmd := parts[0]

	// Create pattern hash
	hash := generatePatternHash(command)

	// Update or create pattern
	if se.patterns[hash] == nil {
		se.patterns[hash] = &CommandPattern{
			Hash:        hash,
			BaseCommand: baseCmd,
			Frequency:   1,
			LastSeen:    time.Now(),
			Variations:  []string{command},
		}
	} else {
		pattern := se.patterns[hash]
		pattern.Frequency++
		pattern.LastSeen = time.Now()
		
		// Add variation if unique
		if !contains(pattern.Variations, command) {
			pattern.Variations = append(pattern.Variations, command)
		}
	}

	// Track common flags
	for _, part := range parts[1:] {
		if strings.HasPrefix(part, "-") {
			prefs.CommonFlags[part]++
		}
	}

	// Detect preferred directories
	for _, part := range parts {
		if strings.HasPrefix(part, "/") || strings.HasPrefix(part, "./") {
			if !contains(prefs.PreferredDirectories, part) {
				prefs.PreferredDirectories = append(prefs.PreferredDirectories, part)
			}
		}
	}
}

// GetSuggestions returns personalized suggestions for a user
func (se *SuggestionEngine) GetSuggestions(userID, currentPrompt string) []Suggestion {
	suggestions := []Suggestion{}

	prefs := se.prefs[userID]
	if prefs == nil {
		return suggestions
	}

	// 1. Alias suggestions based on frequency
	for _, pattern := range se.patterns {
		if pattern.Frequency >= 5 && pattern.SuggestedAlias == "" {
			suggestions = append(suggestions, Suggestion{
				Type:        "alias",
				Title:       "Create Alias",
				Description: fmt.Sprintf("You've run similar commands %d times", pattern.Frequency),
				Command:     fmt.Sprintf("alias %s='%s'", generateAliasName(pattern.BaseCommand), pattern.Variations[0]),
				Confidence:  calculateAliasConfidence(pattern.Frequency),
				Reason:      "Frequently used command pattern",
			})
		}
	}

	// 2. Context-based suggestions
	if currentPrompt != "" {
		contextSuggestions := se.getContextSuggestions(userID, currentPrompt)
		suggestions = append(suggestions, contextSuggestions...)
	}

	// 3. Optimization suggestions
	optSuggestions := se.getOptimizationSuggestions(userID)
	suggestions = append(suggestions, optSuggestions...)

	// 4. Pattern-based predictions
	if len(prefs.CommandHistory) > 0 {
		predicted := se.predictNextCommand(prefs.CommandHistory)
		if predicted != "" {
			suggestions = append(suggestions, Suggestion{
				Type:        "pattern",
				Title:       "Predicted Next Command",
				Description: "Based on your command history",
				Command:     predicted,
				Confidence:  70,
				Reason:      "Common command sequence detected",
			})
		}
	}

	return suggestions
}

// getContextSuggestions provides context-aware suggestions
func (se *SuggestionEngine) getContextSuggestions(userID, prompt string) []Suggestion {
	suggestions := []Suggestion{}
	prefs := se.prefs[userID]

	// Suggest preferred directories
	if strings.Contains(prompt, "find") || strings.Contains(prompt, "search") {
		for _, dir := range prefs.PreferredDirectories {
			if len(suggestions) < 3 {
				suggestions = append(suggestions, Suggestion{
					Type:        "context",
					Title:       "Use Preferred Directory",
					Description: fmt.Sprintf("You often search in %s", dir),
					Command:     fmt.Sprintf("find %s ...", dir),
					Confidence:  80,
					Reason:      "Frequently accessed directory",
				})
			}
		}
	}

	// Suggest common flags
	baseCmd := extractBaseCommand(prompt)
	if flags, ok := prefs.CommonFlags[baseCmd]; ok && flags > 3 {
		suggestions = append(suggestions, Suggestion{
			Type:        "context",
			Title:       "Add Common Flags",
			Description: fmt.Sprintf("You usually use these flags with %s", baseCmd),
			Confidence:  75,
			Reason:      "Commonly used flags",
		})
	}

	return suggestions
}

// getOptimizationSuggestions suggests command optimizations
func (se *SuggestionEngine) getOptimizationSuggestions(userID string) []Suggestion {
	suggestions := []Suggestion{}
	prefs := se.prefs[userID]

	// Analyze recent commands for optimization opportunities
	for _, cmd := range prefs.CommandHistory {
		// Suggest using grep instead of find + grep
		if strings.Contains(cmd, "find") && strings.Contains(cmd, "grep") {
			suggestions = append(suggestions, Suggestion{
				Type:        "optimization",
				Title:       "Optimize Search",
				Description: "Use grep -r for faster recursive search",
				Command:     optimizeFindGrep(cmd),
				Confidence:  85,
				Reason:      "grep -r is faster than find | grep",
			})
		}

		// Suggest parallel execution
		if strings.Contains(cmd, "for") && strings.Contains(cmd, "do") {
			suggestions = append(suggestions, Suggestion{
				Type:        "optimization",
				Title:       "Parallelize Loop",
				Description: "Use xargs -P for parallel execution",
				Confidence:  75,
				Reason:      "Parallel execution is faster",
			})
		}
	}

	return suggestions
}

// predictNextCommand predicts the next command based on history
func (se *SuggestionEngine) predictNextCommand(history []string) string {
	if len(history) < 2 {
		return ""
	}

	// Look for common sequences
	lastCmd := history[len(history)-1]
	
	// Common git workflows
	if strings.HasPrefix(lastCmd, "git add") {
		return "git commit -m '...'"
	}
	if strings.HasPrefix(lastCmd, "git commit") {
		return "git push"
	}

	// Common kubectl workflows
	if strings.Contains(lastCmd, "kubectl apply") {
		return "kubectl get pods"
	}

	// Common file operations
	if strings.HasPrefix(lastCmd, "mkdir") {
		return "cd " + extractDirName(lastCmd)
	}

	return ""
}

// SuggestAlias suggests an alias for a command pattern
func (se *SuggestionEngine) SuggestAlias(pattern *CommandPattern) string {
	if pattern.Frequency < 5 {
		return ""
	}

	aliasName := generateAliasName(pattern.BaseCommand)
	command := pattern.Variations[0]

	return fmt.Sprintf("alias %s='%s'", aliasName, command)
}

// Helper functions

func generatePatternHash(command string) string {
	// Simple hash based on base command and structure
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ""
	}
	
	// Use base command + number of args as hash
	return fmt.Sprintf("%s-%d", parts[0], len(parts))
}

func generateAliasName(baseCmd string) string {
	// Generate a short alias name
	if len(baseCmd) <= 3 {
		return baseCmd
	}
	return baseCmd[:3]
}

func calculateAliasConfidence(frequency int) int {
	// Higher frequency = higher confidence
	if frequency >= 10 {
		return 95
	}
	if frequency >= 7 {
		return 85
	}
	if frequency >= 5 {
		return 75
	}
	return 60
}

func extractBaseCommand(prompt string) string {
	words := strings.Fields(prompt)
	if len(words) > 0 {
		return words[0]
	}
	return ""
}

func extractDirName(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func optimizeFindGrep(cmd string) string {
	// Convert "find . -name '*.log' | grep error" to "grep -r error --include='*.log' ."
	// Simplified version
	return "grep -r [pattern] --include='*.log' ."
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// FeedbackType represents user feedback on suggestions
type FeedbackType string

const (
	FeedbackAccepted FeedbackType = "accepted"
	FeedbackRejected FeedbackType = "rejected"
	FeedbackIgnored  FeedbackType = "ignored"
)

// RecordFeedback records user feedback on a suggestion
func (se *SuggestionEngine) RecordFeedback(userID, suggestionType string, feedback FeedbackType) {
	prefs := se.prefs[userID]
	if prefs == nil {
		return
	}

	// If rejected multiple times, add to ignored list
	if feedback == FeedbackRejected {
		key := fmt.Sprintf("%s-%s", suggestionType, time.Now().Format("2006-01-02"))
		prefs.IgnoredSuggestions = append(prefs.IgnoredSuggestions, key)
	}
}

// ShouldSuggest checks if a suggestion should be shown based on feedback
func (se *SuggestionEngine) ShouldSuggest(userID, suggestionType string) bool {
	prefs := se.prefs[userID]
	if prefs == nil {
		return true
	}

	// Check if recently ignored
	key := fmt.Sprintf("%s-%s", suggestionType, time.Now().Format("2006-01-02"))
	return !contains(prefs.IgnoredSuggestions, key)
}
