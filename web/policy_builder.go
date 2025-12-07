package web

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/SagheerAkram/QuickCmd/core/policy"
	"gopkg.in/yaml.v3"
)

// VisualRule represents a policy rule in a visual/form-friendly format
type VisualRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	RuleType    string   `json:"rule_type"` // "allowlist", "denylist", "approval"
	Pattern     string   `json:"pattern"`
	IsRegex     bool     `json:"is_regex"`
	Action      string   `json:"action"` // "allow", "block", "approve"
	AppliesTo   []string `json:"applies_to"` // roles
	Priority    int      `json:"priority"`
	Enabled     bool     `json:"enabled"`
	Examples    []string `json:"examples,omitempty"`
}

// PolicyBuilder helps build and test policy rules visually
type PolicyBuilder struct {
	rules map[string]*VisualRule
}

// NewPolicyBuilder creates a new policy builder
func NewPolicyBuilder() *PolicyBuilder {
	return &PolicyBuilder{
		rules: make(map[string]*VisualRule),
	}
}

// AddRule adds a new visual rule
func (pb *PolicyBuilder) AddRule(rule *VisualRule) error {
	// Validate rule
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if rule.Pattern == "" {
		return fmt.Errorf("pattern is required")
	}
	if rule.RuleType == "" {
		return fmt.Errorf("rule type is required")
	}

	// Validate regex if specified
	if rule.IsRegex {
		if _, err := regexp.Compile(rule.Pattern); err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
	}

	// Generate ID if not provided
	if rule.ID == "" {
		rule.ID = generateRuleID(rule.Name)
	}

	pb.rules[rule.ID] = rule
	return nil
}

// RemoveRule removes a rule by ID
func (pb *PolicyBuilder) RemoveRule(id string) {
	delete(pb.rules, id)
}

// GetRule gets a rule by ID
func (pb *PolicyBuilder) GetRule(id string) *VisualRule {
	return pb.rules[id]
}

// ListRules returns all rules
func (pb *PolicyBuilder) ListRules() []*VisualRule {
	rules := make([]*VisualRule, 0, len(pb.rules))
	for _, rule := range pb.rules {
		rules = append(rules, rule)
	}
	return rules
}

// TestRule tests a rule against a command
func (pb *PolicyBuilder) TestRule(ruleID string, command string) (*TestResult, error) {
	rule := pb.rules[ruleID]
	if rule == nil {
		return nil, fmt.Errorf("rule not found: %s", ruleID)
	}

	result := &TestResult{
		RuleID:  ruleID,
		Command: command,
		Matched: false,
		Action:  rule.Action,
	}

	// Test pattern match
	if rule.IsRegex {
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %w", err)
		}
		result.Matched = re.MatchString(command)
	} else {
		// Simple substring match
		result.Matched = contains(command, rule.Pattern)
	}

	if result.Matched {
		result.Message = fmt.Sprintf("Command matches rule '%s'", rule.Name)
		switch rule.Action {
		case "block":
			result.Message += " - BLOCKED"
		case "approve":
			result.Message += " - REQUIRES APPROVAL"
		case "allow":
			result.Message += " - ALLOWED"
		}
	} else {
		result.Message = "Command does not match this rule"
	}

	return result, nil
}

// TestResult represents the result of testing a rule
type TestResult struct {
	RuleID  string `json:"rule_id"`
	Command string `json:"command"`
	Matched bool   `json:"matched"`
	Action  string `json:"action"`
	Message string `json:"message"`
}

// ImpactAnalysis analyzes the impact of a rule on historical commands
type ImpactAnalysis struct {
	RuleID         string   `json:"rule_id"`
	TotalCommands  int      `json:"total_commands"`
	MatchedCount   int      `json:"matched_count"`
	BlockedCount   int      `json:"blocked_count"`
	ApprovalCount  int      `json:"approval_count"`
	MatchedExamples []string `json:"matched_examples"`
}

// AnalyzeImpact analyzes how a rule would affect commands
func (pb *PolicyBuilder) AnalyzeImpact(ruleID string, commands []string) (*ImpactAnalysis, error) {
	rule := pb.rules[ruleID]
	if rule == nil {
		return nil, fmt.Errorf("rule not found: %s", ruleID)
	}

	analysis := &ImpactAnalysis{
		RuleID:          ruleID,
		TotalCommands:   len(commands),
		MatchedExamples: []string{},
	}

	for _, cmd := range commands {
		result, err := pb.TestRule(ruleID, cmd)
		if err != nil {
			continue
		}

		if result.Matched {
			analysis.MatchedCount++
			
			// Track action counts
			switch rule.Action {
			case "block":
				analysis.BlockedCount++
			case "approve":
				analysis.ApprovalCount++
			}

			// Add to examples (max 10)
			if len(analysis.MatchedExamples) < 10 {
				analysis.MatchedExamples = append(analysis.MatchedExamples, cmd)
			}
		}
	}

	return analysis, nil
}

// ConvertToYAML converts visual rules to policy YAML
func (pb *PolicyBuilder) ConvertToYAML() (string, error) {
	policyConfig := &policy.Policy{
		Allowlist:        []policy.Pattern{},
		Denylist:         []policy.Pattern{},
		ApprovalRequired: []policy.Pattern{},
	}

	// Group rules by type
	for _, rule := range pb.rules {
		if !rule.Enabled {
			continue
		}

		pattern := policy.Pattern{
			Pattern: rule.Pattern,
			Reason:  rule.Description,
		}

		switch rule.RuleType {
		case "allowlist":
			policyConfig.Allowlist = append(policyConfig.Allowlist, pattern)
		case "denylist":
			policyConfig.Denylist = append(policyConfig.Denylist, pattern)
		case "approval":
			policyConfig.ApprovalRequired = append(policyConfig.ApprovalRequired, pattern)
		}
	}

	// Convert to YAML
	yamlBytes, err := yaml.Marshal(policyConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return string(yamlBytes), nil
}

// ImportFromYAML imports rules from policy YAML
func (pb *PolicyBuilder) ImportFromYAML(yamlContent string) error {
	var policyConfig policy.Policy
	if err := yaml.Unmarshal([]byte(yamlContent), &policyConfig); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Import allowlist rules
	for i, pattern := range policyConfig.Allowlist {
		rule := &VisualRule{
			ID:          fmt.Sprintf("allowlist-%d", i),
			Name:        fmt.Sprintf("Allowlist Rule %d", i+1),
			Description: pattern.Reason,
			RuleType:    "allowlist",
			Pattern:     pattern.Pattern,
			IsRegex:     true,
			Action:      "allow",
			Enabled:     true,
		}
		pb.AddRule(rule)
	}

	// Import denylist rules
	for i, pattern := range policyConfig.Denylist {
		rule := &VisualRule{
			ID:          fmt.Sprintf("denylist-%d", i),
			Name:        fmt.Sprintf("Denylist Rule %d", i+1),
			Description: pattern.Reason,
			RuleType:    "denylist",
			Pattern:     pattern.Pattern,
			IsRegex:     true,
			Action:      "block",
			Enabled:     true,
		}
		pb.AddRule(rule)
	}

	// Import approval rules
	for i, pattern := range policyConfig.ApprovalRequired {
		rule := &VisualRule{
			ID:          fmt.Sprintf("approval-%d", i),
			Name:        fmt.Sprintf("Approval Rule %d", i+1),
			Description: pattern.Reason,
			RuleType:    "approval",
			Pattern:     pattern.Pattern,
			IsRegex:     true,
			Action:      "approve",
			Enabled:     true,
		}
		pb.AddRule(rule)
	}

	return nil
}

// RuleTemplate represents a pre-built rule template
type RuleTemplate struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Rules       []VisualRule `json:"rules"`
}

// GetTemplates returns common rule templates
func GetTemplates() []RuleTemplate {
	return []RuleTemplate{
		{
			Name:        "Block Production Deletions",
			Description: "Prevent deletion commands in production environment",
			Category:    "Safety",
			Rules: []VisualRule{
				{
					Name:        "Block rm in production",
					Description: "Blocks rm commands with production paths",
					RuleType:    "denylist",
					Pattern:     "rm.*production",
					IsRegex:     true,
					Action:      "block",
					Enabled:     true,
				},
				{
					Name:        "Block kubectl delete in production",
					Description: "Blocks kubectl delete in production namespace",
					RuleType:    "denylist",
					Pattern:     "kubectl delete.*-n production",
					IsRegex:     true,
					Action:      "block",
					Enabled:     true,
				},
			},
		},
		{
			Name:        "Require Approval for High-Risk Operations",
			Description: "Require approval for destructive operations",
			Category:    "Approval",
			Rules: []VisualRule{
				{
					Name:        "Approve force operations",
					Description: "Require approval for git force push",
					RuleType:    "approval",
					Pattern:     "git push.*--force",
					IsRegex:     true,
					Action:      "approve",
					Enabled:     true,
				},
				{
					Name:        "Approve AWS deletions",
					Description: "Require approval for AWS delete operations",
					RuleType:    "approval",
					Pattern:     "aws.*delete",
					IsRegex:     true,
					Action:      "approve",
					Enabled:     true,
				},
			},
		},
		{
			Name:        "Allow Safe Read Operations",
			Description: "Explicitly allow safe read-only commands",
			Category:    "Allowlist",
			Rules: []VisualRule{
				{
					Name:        "Allow ls commands",
					Description: "Allow directory listing",
					RuleType:    "allowlist",
					Pattern:     "^ls",
					IsRegex:     true,
					Action:      "allow",
					Enabled:     true,
				},
				{
					Name:        "Allow cat commands",
					Description: "Allow file reading",
					RuleType:    "allowlist",
					Pattern:     "^cat",
					IsRegex:     true,
					Action:      "allow",
					Enabled:     true,
				},
			},
		},
	}
}

// Helper functions

func generateRuleID(name string) string {
	// Simple ID generation - in production, use UUID
	return fmt.Sprintf("rule-%d", len(name))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

// ToJSON converts visual rule to JSON
func (vr *VisualRule) ToJSON() (string, error) {
	bytes, err := json.MarshalIndent(vr, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
