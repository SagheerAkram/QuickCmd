package translator

import (
	"fmt"
	"strings"
)

// ConfidenceComponent represents a component of the confidence score
type ConfidenceComponent string

const (
	ComponentPattern ConfidenceComponent = "pattern"
	ComponentContext ConfidenceComponent = "context"
	ComponentRisk    ConfidenceComponent = "risk"
	ComponentPlugin  ConfidenceComponent = "plugin"
)

// ConfidenceBreakdown provides detailed scoring information
type ConfidenceBreakdown struct {
	Overall    int                        `json:"overall"`
	Components map[ConfidenceComponent]int `json:"components"`
	Reasons    []string                   `json:"reasons"`
	Warnings   []string                   `json:"warnings"`
	Tips       []string                   `json:"tips"`
}

// NewConfidenceBreakdown creates a new confidence breakdown
func NewConfidenceBreakdown() *ConfidenceBreakdown {
	return &ConfidenceBreakdown{
		Components: make(map[ConfidenceComponent]int),
		Reasons:    []string{},
		Warnings:   []string{},
		Tips:       []string{},
	}
}

// AddComponent adds a confidence component score
func (cb *ConfidenceBreakdown) AddComponent(component ConfidenceComponent, score int, reason string) {
	cb.Components[component] = score
	if reason != "" {
		cb.Reasons = append(cb.Reasons, reason)
	}
}

// AddWarning adds a warning message
func (cb *ConfidenceBreakdown) AddWarning(warning string) {
	cb.Warnings = append(cb.Warnings, warning)
}

// AddTip adds a helpful tip
func (cb *ConfidenceBreakdown) AddTip(tip string) {
	cb.Tips = append(cb.Tips, tip)
}

// Calculate computes the overall confidence score
func (cb *ConfidenceBreakdown) Calculate() {
	if len(cb.Components) == 0 {
		cb.Overall = 0
		return
	}

	total := 0
	for _, score := range cb.Components {
		total += score
	}
	cb.Overall = total / len(cb.Components)
}

// Visualize creates a visual representation of the confidence score
func (cb *ConfidenceBreakdown) Visualize() string {
	var sb strings.Builder

	// Overall confidence with progress bar
	sb.WriteString(fmt.Sprintf("\n✨ Confidence: %d%% %s\n\n", cb.Overall, progressBar(cb.Overall)))

	// Component breakdown
	sb.WriteString("Breakdown:\n")
	
	componentNames := map[ConfidenceComponent]string{
		ComponentPattern: "Pattern Match",
		ComponentContext: "Context Awareness",
		ComponentRisk:    "Risk Assessment",
		ComponentPlugin:  "Plugin Analysis",
	}

	for component, score := range cb.Components {
		name := componentNames[component]
		if name == "" {
			name = string(component)
		}
		sb.WriteString(fmt.Sprintf("  %-20s %3d%% %s\n", name+":", score, progressBar(score)))
	}

	// Reasons
	if len(cb.Reasons) > 0 {
		sb.WriteString("\nWhy this command?\n")
		for _, reason := range cb.Reasons {
			sb.WriteString(fmt.Sprintf("  ✓ %s\n", reason))
		}
	}

	// Warnings
	if len(cb.Warnings) > 0 {
		sb.WriteString("\nWarnings:\n")
		for _, warning := range cb.Warnings {
			sb.WriteString(fmt.Sprintf("  ⚠ %s\n", warning))
		}
	}

	// Tips
	if len(cb.Tips) > 0 {
		sb.WriteString("\nTips:\n")
		for _, tip := range cb.Tips {
			sb.WriteString(fmt.Sprintf("  💡 %s\n", tip))
		}
	}

	return sb.String()
}

// progressBar creates a visual progress bar
func progressBar(percentage int) string {
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}

	filled := percentage / 5 // 20 chars total
	empty := 20 - filled

	return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}

// ExplainChoice provides detailed explanation of why a command was chosen
func (c *Candidate) ExplainChoice(prompt string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\nCommand: %s\n\n", c.Command))

	// Pattern matching explanation
	if c.Confidence >= 90 {
		sb.WriteString("✓ Exact match to known pattern\n")
	} else if c.Confidence >= 70 {
		sb.WriteString("✓ Close match to similar patterns\n")
	} else {
		sb.WriteString("⚠ Fuzzy match - verify before executing\n")
	}

	// Risk explanation
	switch c.RiskLevel {
	case RiskSafe:
		sb.WriteString("✓ Safe operation (read-only)\n")
	case RiskMedium:
		sb.WriteString("⚠ Medium risk (modifies files/state)\n")
	case RiskHigh:
		sb.WriteString("🔴 High risk (destructive operation)\n")
	}

	// Command breakdown
	sb.WriteString(fmt.Sprintf("\n%s\n", c.Explanation))

	// Affected resources
	if len(c.AffectedPaths) > 0 {
		sb.WriteString("\nAffected paths:\n")
		for _, path := range c.AffectedPaths {
			sb.WriteString(fmt.Sprintf("  • %s\n", path))
		}
	}

	if len(c.NetworkTargets) > 0 {
		sb.WriteString("\nNetwork targets:\n")
		for _, target := range c.NetworkTargets {
			sb.WriteString(fmt.Sprintf("  • %s\n", target))
		}
	}

	return sb.String()
}

// CalculateConfidenceBreakdown creates a detailed confidence breakdown for a candidate
func (c *Candidate) CalculateConfidenceBreakdown(prompt string) *ConfidenceBreakdown {
	breakdown := NewConfidenceBreakdown()

	// Pattern matching score
	patternScore := c.Confidence
	if patternScore >= 95 {
		breakdown.AddComponent(ComponentPattern, patternScore, "Exact match to template pattern")
	} else if patternScore >= 80 {
		breakdown.AddComponent(ComponentPattern, patternScore, "Strong pattern match")
	} else {
		breakdown.AddComponent(ComponentPattern, patternScore, "Fuzzy pattern match")
		breakdown.AddWarning("Lower confidence - verify command before executing")
	}

	// Context awareness (check if command uses current directory, environment, etc.)
	contextScore := 85
	if strings.Contains(c.Command, ".") || strings.Contains(c.Command, "$PWD") {
		contextScore = 95
		breakdown.AddComponent(ComponentContext, contextScore, "Uses current directory context")
	} else {
		breakdown.AddComponent(ComponentContext, contextScore, "Context-aware command")
	}

	// Risk assessment
	riskScore := 100
	switch c.RiskLevel {
	case RiskSafe:
		riskScore = 100
		breakdown.AddComponent(ComponentRisk, riskScore, "Safe operation (read-only)")
	case RiskMedium:
		riskScore = 75
		breakdown.AddComponent(ComponentRisk, riskScore, "Medium risk (modifies state)")
		breakdown.AddWarning("This command will modify files or system state")
	case RiskHigh:
		riskScore = 50
		breakdown.AddComponent(ComponentRisk, riskScore, "High risk (destructive)")
		breakdown.AddWarning("This is a destructive operation - use with caution")
		breakdown.AddTip("Consider running in sandbox mode first")
	}

	// Performance tips
	if c.Destructiveness > 50 {
		breakdown.AddTip("Create a backup before executing")
	}

	if len(c.AffectedPaths) > 10 {
		breakdown.AddWarning("This command affects many files - could be slow")
	}

	// Calculate overall score
	breakdown.Calculate()

	return breakdown
}
