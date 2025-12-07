package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolicyBuilder(t *testing.T) {
	t.Run("creates new builder", func(t *testing.T) {
		pb := NewPolicyBuilder()
		assert.NotNil(t, pb)
		assert.Empty(t, pb.ListRules())
	})

	t.Run("adds and retrieves rules", func(t *testing.T) {
		pb := NewPolicyBuilder()
		
		rule := &VisualRule{
			ID:       "test-1",
			Name:     "Test Rule",
			Pattern:  "rm -rf",
			RuleType: "denylist",
			Action:   "block",
			IsRegex:  false,
			Enabled:  true,
		}

		err := pb.AddRule(rule)
		assert.NoError(t, err)

		retrieved := pb.GetRule("test-1")
		assert.NotNil(t, retrieved)
		assert.Equal(t, "Test Rule", retrieved.Name)
	})

	t.Run("validates rule requirements", func(t *testing.T) {
		pb := NewPolicyBuilder()

		// Missing name
		err := pb.AddRule(&VisualRule{Pattern: "test", RuleType: "denylist"})
		assert.Error(t, err)

		// Missing pattern
		err = pb.AddRule(&VisualRule{Name: "test", RuleType: "denylist"})
		assert.Error(t, err)

		// Missing rule type
		err = pb.AddRule(&VisualRule{Name: "test", Pattern: "test"})
		assert.Error(t, err)
	})

	t.Run("validates regex patterns", func(t *testing.T) {
		pb := NewPolicyBuilder()

		// Invalid regex
		err := pb.AddRule(&VisualRule{
			Name:     "Bad Regex",
			Pattern:  "[invalid",
			RuleType: "denylist",
			IsRegex:  true,
		})
		assert.Error(t, err)

		// Valid regex
		err = pb.AddRule(&VisualRule{
			Name:     "Good Regex",
			Pattern:  "^rm.*-rf",
			RuleType: "denylist",
			IsRegex:  true,
		})
		assert.NoError(t, err)
	})

	t.Run("removes rules", func(t *testing.T) {
		pb := NewPolicyBuilder()
		
		rule := &VisualRule{
			ID:       "test-1",
			Name:     "Test",
			Pattern:  "test",
			RuleType: "denylist",
		}
		pb.AddRule(rule)

		pb.RemoveRule("test-1")
		assert.Nil(t, pb.GetRule("test-1"))
	})
}

func TestRuleTesting(t *testing.T) {
	pb := NewPolicyBuilder()

	rule := &VisualRule{
		ID:       "test-1",
		Name:     "Block rm -rf",
		Pattern:  "rm.*-rf",
		RuleType: "denylist",
		Action:   "block",
		IsRegex:  true,
		Enabled:  true,
	}
	pb.AddRule(rule)

	t.Run("matches dangerous command", func(t *testing.T) {
		result, err := pb.TestRule("test-1", "rm -rf /data")
		assert.NoError(t, err)
		assert.True(t, result.Matched)
		assert.Equal(t, "block", result.Action)
		assert.Contains(t, result.Message, "BLOCKED")
	})

	t.Run("does not match safe command", func(t *testing.T) {
		result, err := pb.TestRule("test-1", "ls -la")
		assert.NoError(t, err)
		assert.False(t, result.Matched)
	})

	t.Run("handles non-existent rule", func(t *testing.T) {
		_, err := pb.TestRule("non-existent", "test")
		assert.Error(t, err)
	})
}

func TestImpactAnalysis(t *testing.T) {
	pb := NewPolicyBuilder()

	rule := &VisualRule{
		ID:       "test-1",
		Name:     "Block deletions",
		Pattern:  "delete",
		RuleType: "denylist",
		Action:   "block",
		IsRegex:  false,
		Enabled:  true,
	}
	pb.AddRule(rule)

	commands := []string{
		"kubectl delete pod api",
		"aws s3 delete bucket",
		"ls -la",
		"cat file.txt",
		"rm delete.txt",
	}

	t.Run("analyzes impact correctly", func(t *testing.T) {
		analysis, err := pb.AnalyzeImpact("test-1", commands)
		assert.NoError(t, err)
		assert.Equal(t, 5, analysis.TotalCommands)
		assert.Equal(t, 3, analysis.MatchedCount) // 3 commands contain "delete"
		assert.Equal(t, 3, analysis.BlockedCount)
		assert.Len(t, analysis.MatchedExamples, 3)
	})
}

func TestYAMLConversion(t *testing.T) {
	pb := NewPolicyBuilder()

	// Add some rules
	pb.AddRule(&VisualRule{
		ID:       "allow-1",
		Name:     "Allow ls",
		Pattern:  "^ls",
		RuleType: "allowlist",
		Action:   "allow",
		IsRegex:  true,
		Enabled:  true,
	})

	pb.AddRule(&VisualRule{
		ID:       "deny-1",
		Name:     "Block rm -rf",
		Pattern:  "rm.*-rf",
		RuleType: "denylist",
		Action:   "block",
		IsRegex:  true,
		Enabled:  true,
	})

	t.Run("converts to YAML", func(t *testing.T) {
		yaml, err := pb.ConvertToYAML()
		assert.NoError(t, err)
		assert.Contains(t, yaml, "allowlist")
		assert.Contains(t, yaml, "denylist")
		assert.Contains(t, yaml, "^ls")
		assert.Contains(t, yaml, "rm.*-rf")
	})

	t.Run("imports from YAML", func(t *testing.T) {
		yamlContent := `
allowlist:
  - pattern: "^cat"
    reason: "Allow cat commands"
denylist:
  - pattern: "rm -rf /"
    reason: "Prevent root deletion"
`
		pb2 := NewPolicyBuilder()
		err := pb2.ImportFromYAML(yamlContent)
		assert.NoError(t, err)

		rules := pb2.ListRules()
		assert.NotEmpty(t, rules)
	})
}

func TestRuleTemplates(t *testing.T) {
	t.Run("returns templates", func(t *testing.T) {
		templates := GetTemplates()
		assert.NotEmpty(t, templates)
		assert.Greater(t, len(templates), 0)

		// Check first template
		template := templates[0]
		assert.NotEmpty(t, template.Name)
		assert.NotEmpty(t, template.Description)
		assert.NotEmpty(t, template.Category)
		assert.NotEmpty(t, template.Rules)
	})

	t.Run("templates have valid rules", func(t *testing.T) {
		templates := GetTemplates()
		
		for _, template := range templates {
			for _, rule := range template.Rules {
				assert.NotEmpty(t, rule.Name)
				assert.NotEmpty(t, rule.Pattern)
				assert.NotEmpty(t, rule.RuleType)
				assert.NotEmpty(t, rule.Action)
			}
		}
	})
}

func TestVisualRuleJSON(t *testing.T) {
	rule := &VisualRule{
		ID:          "test-1",
		Name:        "Test Rule",
		Description: "Test description",
		Pattern:     "test",
		RuleType:    "denylist",
		Action:      "block",
		IsRegex:     false,
		Enabled:     true,
	}

	t.Run("converts to JSON", func(t *testing.T) {
		json, err := rule.ToJSON()
		assert.NoError(t, err)
		assert.Contains(t, json, "test-1")
		assert.Contains(t, json, "Test Rule")
		assert.Contains(t, json, "denylist")
	})
}
