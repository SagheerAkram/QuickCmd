package translator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfidenceBreakdown(t *testing.T) {
	t.Run("creates new breakdown", func(t *testing.T) {
		breakdown := NewConfidenceBreakdown()
		assert.NotNil(t, breakdown)
		assert.Equal(t, 0, breakdown.Overall)
		assert.Empty(t, breakdown.Components)
	})

	t.Run("adds components and calculates overall", func(t *testing.T) {
		breakdown := NewConfidenceBreakdown()
		breakdown.AddComponent(ComponentPattern, 95, "Exact match")
		breakdown.AddComponent(ComponentContext, 85, "Context aware")
		breakdown.AddComponent(ComponentRisk, 100, "Safe operation")
		breakdown.Calculate()

		assert.Equal(t, 93, breakdown.Overall) // (95+85+100)/3 = 93
		assert.Len(t, breakdown.Components, 3)
		assert.Len(t, breakdown.Reasons, 3)
	})

	t.Run("adds warnings and tips", func(t *testing.T) {
		breakdown := NewConfidenceBreakdown()
		breakdown.AddWarning("This is destructive")
		breakdown.AddTip("Use sandbox mode")

		assert.Len(t, breakdown.Warnings, 1)
		assert.Len(t, breakdown.Tips, 1)
	})
}

func TestProgressBar(t *testing.T) {
	tests := []struct {
		percentage int
		expected   int // filled chars
	}{
		{0, 0},
		{50, 10},
		{100, 20},
		{75, 15},
	}

	for _, tt := range tests {
		bar := progressBar(tt.percentage)
		filled := 0
		for _, char := range bar {
			if char == '█' {
				filled++
			}
		}
		assert.Equal(t, tt.expected, filled, "percentage: %d", tt.percentage)
	}
}

func TestCandidateConfidenceBreakdown(t *testing.T) {
	t.Run("safe command gets high scores", func(t *testing.T) {
		candidate := &Candidate{
			Command:     "ls -la",
			Confidence:  95,
			RiskLevel:   RiskSafe,
			Explanation: "List files",
		}

		breakdown := candidate.CalculateConfidenceBreakdown("list files")
		breakdown.Calculate()

		assert.Greater(t, breakdown.Overall, 90)
		assert.Contains(t, breakdown.Components, ComponentPattern)
		assert.Contains(t, breakdown.Components, ComponentRisk)
		assert.Equal(t, 100, breakdown.Components[ComponentRisk])
	})

	t.Run("destructive command gets warnings", func(t *testing.T) {
		candidate := &Candidate{
			Command:         "rm -rf /data",
			Confidence:      90,
			RiskLevel:       RiskHigh,
			Destructiveness: 95,
			AffectedPaths:   []string{"/data"},
		}

		breakdown := candidate.CalculateConfidenceBreakdown("delete data")

		assert.NotEmpty(t, breakdown.Warnings)
		assert.NotEmpty(t, breakdown.Tips)
		assert.Contains(t, breakdown.Warnings[0], "destructive")
	})

	t.Run("low confidence gets warning", func(t *testing.T) {
		candidate := &Candidate{
			Command:    "some-unknown-command",
			Confidence: 60,
			RiskLevel:  RiskMedium,
		}

		breakdown := candidate.CalculateConfidenceBreakdown("do something")

		assert.NotEmpty(t, breakdown.Warnings)
	})
}

func TestExplainChoice(t *testing.T) {
	t.Run("explains high confidence command", func(t *testing.T) {
		candidate := &Candidate{
			Command:       "find . -type f -size +100M",
			Confidence:    95,
			RiskLevel:     RiskSafe,
			Explanation:   "Finds files larger than 100MB",
			AffectedPaths: []string{"."},
		}

		explanation := candidate.ExplainChoice("find large files")

		assert.Contains(t, explanation, "find . -type f -size +100M")
		assert.Contains(t, explanation, "Exact match")
		assert.Contains(t, explanation, "Safe operation")
	})

	t.Run("explains risky command", func(t *testing.T) {
		candidate := &Candidate{
			Command:        "kubectl delete deployment api",
			Confidence:     85,
			RiskLevel:      RiskHigh,
			Explanation:    "Deletes Kubernetes deployment",
			NetworkTargets: []string{"kubernetes-api"},
		}

		explanation := candidate.ExplainChoice("delete api deployment")

		assert.Contains(t, explanation, "High risk")
		assert.Contains(t, explanation, "kubernetes-api")
	})
}

func TestVisualize(t *testing.T) {
	breakdown := NewConfidenceBreakdown()
	breakdown.AddComponent(ComponentPattern, 95, "Exact match")
	breakdown.AddComponent(ComponentContext, 90, "Context aware")
	breakdown.AddComponent(ComponentRisk, 100, "Safe")
	breakdown.AddWarning("Test warning")
	breakdown.AddTip("Test tip")
	breakdown.Calculate()

	visual := breakdown.Visualize()

	assert.Contains(t, visual, "Confidence:")
	assert.Contains(t, visual, "Pattern Match")
	assert.Contains(t, visual, "Context Awareness")
	assert.Contains(t, visual, "Risk Assessment")
	assert.Contains(t, visual, "Why this command?")
	assert.Contains(t, visual, "Warnings:")
	assert.Contains(t, visual, "Tips:")
	assert.Contains(t, visual, "█") // Progress bar
}
