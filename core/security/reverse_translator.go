package security

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/SagheerAkram/QuickCmd/core/policy"
	"github.com/SagheerAkram/QuickCmd/core/translator"
)

// ReverseTranslator converts commands back to natural language prompts
type ReverseTranslator struct {
	dangerousCommands map[string]*DangerousCommand
}

// DangerousCommand represents a known dangerous command pattern
type DangerousCommand struct {
	Command          string
	RiskLevel        string
	NaturalPrompts   []string
	BlockedByPolicy  bool
	SuggestedRule    string
	ImpactDescription string
}

// SecurityGap represents a gap in policy coverage
type SecurityGap struct {
	Command       string
	RiskLevel     string
	Prompts       []string
	CurrentStatus string // "blocked", "allowed", "requires_approval"
	Recommendation string
	SuggestedRule  *policy.Pattern
}

// NewReverseTranslator creates a new reverse translator
func NewReverseTranslator() *ReverseTranslator {
	rt := &ReverseTranslator{
		dangerousCommands: make(map[string]*DangerousCommand),
	}
	rt.loadDangerousCommands()
	return rt
}

// loadDangerousCommands loads known dangerous command patterns
func (rt *ReverseTranslator) loadDangerousCommands() {
	dangerous := []DangerousCommand{
		{
			Command:   "rm -rf /",
			RiskLevel: "critical",
			NaturalPrompts: []string{
				"delete everything",
				"remove all files",
				"clean entire system",
				"wipe root directory",
			},
			ImpactDescription: "Deletes entire filesystem - system destruction",
		},
		{
			Command:   "rm -rf /var/lib/production/*",
			RiskLevel: "high",
			NaturalPrompts: []string{
				"clean up production data",
				"remove old production files",
				"delete everything in production directory",
				"clear production storage",
			},
			ImpactDescription: "Deletes all production data",
		},
		{
			Command:   "kubectl delete namespace production",
			RiskLevel: "high",
			NaturalPrompts: []string{
				"delete production namespace",
				"remove production environment",
				"clean up production cluster",
			},
			ImpactDescription: "Destroys entire production namespace",
		},
		{
			Command:   "git push --force origin main",
			RiskLevel: "high",
			NaturalPrompts: []string{
				"force push to main",
				"override main branch",
				"push changes forcefully",
			},
			ImpactDescription: "Overwrites main branch history",
		},
		{
			Command:   "aws s3 rb s3://production-data --force",
			RiskLevel: "high",
			NaturalPrompts: []string{
				"delete production s3 bucket",
				"remove production storage",
				"clean up s3 bucket",
			},
			ImpactDescription: "Permanently deletes S3 bucket and all data",
		},
		{
			Command:   "DROP DATABASE production;",
			RiskLevel: "critical",
			NaturalPrompts: []string{
				"delete production database",
				"remove database",
				"drop prod db",
			},
			ImpactDescription: "Permanently deletes entire database",
		},
		{
			Command:   "chmod 777 -R /",
			RiskLevel: "high",
			NaturalPrompts: []string{
				"make everything writable",
				"fix permissions",
				"allow all access",
			},
			ImpactDescription: "Removes all security permissions",
		},
		{
			Command:   "dd if=/dev/zero of=/dev/sda",
			RiskLevel: "critical",
			NaturalPrompts: []string{
				"wipe disk",
				"erase hard drive",
				"format drive",
			},
			ImpactDescription: "Destroys all data on primary disk",
		},
	}

	for _, cmd := range dangerous {
		rt.dangerousCommands[cmd.Command] = &cmd
	}
}

// ReverseTranslate converts a command to possible natural language prompts
func (rt *ReverseTranslator) ReverseTranslate(command string) []string {
	// Check if it's a known dangerous command
	if dangerous, exists := rt.dangerousCommands[command]; exists {
		return dangerous.NaturalPrompts
	}

	// Generate prompts based on command structure
	prompts := []string{}

	// Analyze command components
	if strings.Contains(command, "rm") && strings.Contains(command, "-rf") {
		prompts = append(prompts, "delete files recursively")
		prompts = append(prompts, "remove directory and contents")
	}

	if strings.Contains(command, "kubectl delete") {
		prompts = append(prompts, "delete kubernetes resource")
		prompts = append(prompts, "remove k8s object")
	}

	if strings.Contains(command, "git push") && strings.Contains(command, "--force") {
		prompts = append(prompts, "force push changes")
		prompts = append(prompts, "override remote branch")
	}

	if strings.Contains(command, "aws") && strings.Contains(command, "delete") {
		prompts = append(prompts, "delete aws resource")
		prompts = append(prompts, "remove cloud infrastructure")
	}

	// Generic prompts if nothing specific found
	if len(prompts) == 0 {
		prompts = append(prompts, "execute: "+command)
	}

	return prompts
}

// FindPolicyGaps identifies commands not covered by current policy
func (rt *ReverseTranslator) FindPolicyGaps(policyEngine *policy.Engine) []*SecurityGap {
	gaps := []*SecurityGap{}

	for cmd, dangerous := range rt.dangerousCommands {
		// Test if command is blocked by policy
		result := policyEngine.Validate(cmd, dangerous.RiskLevel, true)

		gap := &SecurityGap{
			Command:   cmd,
			RiskLevel: dangerous.RiskLevel,
			Prompts:   dangerous.NaturalPrompts,
		}

		if !result.Allowed {
			gap.CurrentStatus = "blocked"
			gap.Recommendation = "✓ Already blocked by policy"
		} else if result.RequiresApproval {
			gap.CurrentStatus = "requires_approval"
			gap.Recommendation = "⚠ Requires approval - consider blocking entirely"
			gap.SuggestedRule = rt.generateBlockRule(cmd)
		} else {
			gap.CurrentStatus = "allowed"
			gap.Recommendation = "🔴 NOT BLOCKED - Add denylist rule immediately!"
			gap.SuggestedRule = rt.generateBlockRule(cmd)
			gaps = append(gaps, gap)
		}
	}

	return gaps
}

// generateBlockRule generates a policy rule to block a command
func (rt *ReverseTranslator) generateBlockRule(command string) *policy.Pattern {
	// Extract key patterns from command
	pattern := escapeRegex(command)
	
	// Make it more general to catch variations
	pattern = strings.ReplaceAll(pattern, " ", "\\s+")
	
	return &policy.Pattern{
		Pattern: pattern,
		Reason:  fmt.Sprintf("Blocks dangerous command: %s", command),
	}
}

// SimulateAttack tests if malicious prompts can bypass policy
func (rt *ReverseTranslator) SimulateAttack(policyEngine *policy.Engine, trans *translator.Translator) *AttackSimulation {
	simulation := &AttackSimulation{
		TotalAttempts: 0,
		Blocked:       0,
		Allowed:       0,
		RequiresApproval: 0,
		Attempts:      []*AttackAttempt{},
	}

	for cmd, dangerous := range rt.dangerousCommands {
		for _, prompt := range dangerous.NaturalPrompts {
			simulation.TotalAttempts++

			// Try to translate prompt
			candidates, err := trans.Translate(prompt)
			if err != nil || len(candidates) == 0 {
				continue
			}

			// Test first candidate against policy
			candidate := candidates[0]
			result := policyEngine.Validate(candidate.Command, string(candidate.RiskLevel), candidate.Destructive)

			attempt := &AttackAttempt{
				Prompt:          prompt,
				GeneratedCommand: candidate.Command,
				TargetCommand:   cmd,
				Matched:         candidate.Command == cmd,
			}

			if !result.Allowed {
				simulation.Blocked++
				attempt.Outcome = "blocked"
			} else if result.RequiresApproval {
				simulation.RequiresApproval++
				attempt.Outcome = "requires_approval"
			} else {
				simulation.Allowed++
				attempt.Outcome = "allowed"
			}

			simulation.Attempts = append(simulation.Attempts, attempt)
		}
	}

	// Calculate security score
	if simulation.TotalAttempts > 0 {
		blockedPercent := float64(simulation.Blocked) / float64(simulation.TotalAttempts) * 100
		approvalPercent := float64(simulation.RequiresApproval) / float64(simulation.TotalAttempts) * 100
		simulation.SecurityScore = int(blockedPercent + (approvalPercent * 0.5))
	}

	return simulation
}

// AttackSimulation represents results of attack simulation
type AttackSimulation struct {
	TotalAttempts    int
	Blocked          int
	Allowed          int
	RequiresApproval int
	SecurityScore    int // 0-100
	Attempts         []*AttackAttempt
}

// AttackAttempt represents a single attack attempt
type AttackAttempt struct {
	Prompt           string
	GeneratedCommand string
	TargetCommand    string
	Matched          bool
	Outcome          string // "blocked", "allowed", "requires_approval"
}

// GenerateSecurityReport generates a comprehensive security report
func (rt *ReverseTranslator) GenerateSecurityReport(policyEngine *policy.Engine) *SecurityReport {
	gaps := rt.FindPolicyGaps(policyEngine)

	report := &SecurityReport{
		TotalDangerousCommands: len(rt.dangerousCommands),
		Blocked:                0,
		RequiresApproval:       0,
		Allowed:                0,
		Gaps:                   gaps,
		Recommendations:        []string{},
	}

	// Count statuses
	for _, gap := range gaps {
		switch gap.CurrentStatus {
		case "blocked":
			report.Blocked++
		case "requires_approval":
			report.RequiresApproval++
		case "allowed":
			report.Allowed++
		}
	}

	// Calculate coverage
	if report.TotalDangerousCommands > 0 {
		report.CoveragePercent = (report.Blocked * 100) / report.TotalDangerousCommands
	}

	// Generate recommendations
	if report.Allowed > 0 {
		report.Recommendations = append(report.Recommendations,
			fmt.Sprintf("🔴 CRITICAL: %d dangerous commands are not blocked", report.Allowed))
	}
	if report.RequiresApproval > 0 {
		report.Recommendations = append(report.Recommendations,
			fmt.Sprintf("⚠ WARNING: %d commands require approval - consider blocking", report.RequiresApproval))
	}
	if report.CoveragePercent >= 90 {
		report.Recommendations = append(report.Recommendations,
			"✓ Good coverage - policy blocks most dangerous commands")
	}

	return report
}

// SecurityReport represents a security analysis report
type SecurityReport struct {
	TotalDangerousCommands int
	Blocked                int
	RequiresApproval       int
	Allowed                int
	CoveragePercent        int
	Gaps                   []*SecurityGap
	Recommendations        []string
}

// Helper functions

func escapeRegex(s string) string {
	// Escape special regex characters
	special := []string{".", "*", "+", "?", "^", "$", "(", ")", "[", "]", "{", "}", "|", "\\"}
	result := s
	for _, char := range special {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// GetDangerousCommandsByRisk returns dangerous commands filtered by risk level
func (rt *ReverseTranslator) GetDangerousCommandsByRisk(riskLevel string) []*DangerousCommand {
	commands := []*DangerousCommand{}
	for _, cmd := range rt.dangerousCommands {
		if cmd.RiskLevel == riskLevel {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// TestPromptSafety tests if a natural language prompt could generate dangerous commands
func (rt *ReverseTranslator) TestPromptSafety(prompt string) *PromptSafetyResult {
	result := &PromptSafetyResult{
		Prompt:            prompt,
		PotentialDangers:  []*DangerousCommand{},
		SafetyScore:       100,
		Recommendation:    "Prompt appears safe",
	}

	// Check if prompt matches any dangerous command prompts
	promptLower := strings.ToLower(prompt)
	for _, dangerous := range rt.dangerousCommands {
		for _, naturalPrompt := range dangerous.NaturalPrompts {
			if strings.Contains(promptLower, strings.ToLower(naturalPrompt)) {
				result.PotentialDangers = append(result.PotentialDangers, dangerous)
				result.SafetyScore -= 30
			}
		}
	}

	if result.SafetyScore < 50 {
		result.Recommendation = "🔴 HIGH RISK - Prompt may generate dangerous commands"
	} else if result.SafetyScore < 80 {
		result.Recommendation = "⚠ MEDIUM RISK - Review generated commands carefully"
	}

	return result
}

// PromptSafetyResult represents the safety analysis of a prompt
type PromptSafetyResult struct {
	Prompt           string
	PotentialDangers []*DangerousCommand
	SafetyScore      int
	Recommendation   string
}
