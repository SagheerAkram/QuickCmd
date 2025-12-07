package translator

import (
	"fmt"
	"strings"
)

// Risk represents the risk level of a command
type Risk string

const (
	RiskSafe   Risk = "safe"
	RiskMedium Risk = "medium"
	RiskHigh   Risk = "high"
)

// Step represents a single step in command breakdown
type Step struct {
	Description string
	Command     string
}

// Candidate represents a potential command translation
type Candidate struct {
	Command        string   // The actual shell command
	Explanation    string   // Human-friendly explanation
	Breakdown      []Step   // Step-by-step breakdown
	Confidence     int      // 0-100 confidence score
	RiskLevel      Risk     // Risk classification
	AffectedPaths  []string // Paths that will be affected
	NetworkTargets []string // Network endpoints accessed
	Destructive    bool     // Whether this is a destructive operation
	RequiresConfirm bool    // Whether typed confirmation is needed
	DocLinks       []string // Links to documentation
}

// String returns a formatted string representation of the candidate
func (c *Candidate) String() string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("Command: %s\n", c.Command))
	sb.WriteString(fmt.Sprintf("Explanation: %s\n", c.Explanation))
	sb.WriteString(fmt.Sprintf("Confidence: %d%%\n", c.Confidence))
	sb.WriteString(fmt.Sprintf("Risk: %s\n", c.RiskLevel))
	
	if c.Destructive {
		sb.WriteString("⚠️  DESTRUCTIVE OPERATION\n")
	}
	
	if len(c.AffectedPaths) > 0 {
		sb.WriteString(fmt.Sprintf("Affected paths: %s\n", strings.Join(c.AffectedPaths, ", ")))
	}
	
	if len(c.NetworkTargets) > 0 {
		sb.WriteString(fmt.Sprintf("Network targets: %s\n", strings.Join(c.NetworkTargets, ", ")))
	}
	
	if len(c.Breakdown) > 0 {
		sb.WriteString("\nBreakdown:\n")
		for i, step := range c.Breakdown {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, step.Description))
			if step.Command != "" {
				sb.WriteString(fmt.Sprintf("     → %s\n", step.Command))
			}
		}
	}
	
	return sb.String()
}

// RiskIcon returns an emoji icon for the risk level
func (c *Candidate) RiskIcon() string {
	switch c.RiskLevel {
	case RiskSafe:
		return "✅"
	case RiskMedium:
		return "⚠️"
	case RiskHigh:
		return "🔴"
	default:
		return "❓"
	}
}

// RiskColor returns an ANSI color code for the risk level
func (c *Candidate) RiskColor() string {
	switch c.RiskLevel {
	case RiskSafe:
		return "\033[32m" // Green
	case RiskMedium:
		return "\033[33m" // Yellow
	case RiskHigh:
		return "\033[31m" // Red
	default:
		return "\033[0m" // Reset
	}
}
