package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
	
	"github.com/yourusername/quickcmd/core/plugins"
)

// GitPlugin handles Git-related command translations
type GitPlugin struct{}

func init() {
	plugin := &GitPlugin{}
	metadata := &plugins.PluginMetadata{
		Name:        "git",
		Version:     "1.0.0",
		Description: "Git operations with automatic backup branches and safety checks",
		Author:      "QuickCMD Team",
		Scopes:      []string{"git:read", "git:write"},
		Enabled:     true,
	}
	
	plugins.Register(plugin, metadata)
}

// Name returns the plugin name
func (p *GitPlugin) Name() string {
	return "git"
}

// Translate translates Git-related prompts into commands
func (p *GitPlugin) Translate(ctx plugins.Context, prompt string) ([]*plugins.Candidate, error) {
	promptLower := strings.ToLower(prompt)
	
	var candidates []*plugins.Candidate
	
	// Pattern: create backup branch and commit
	if matched, _ := regexp.MatchString(`(?i)(create|make)\s+(backup\s+)?branch.*commit`, promptLower); matched {
		timestamp := time.Now().Format("20060102-150405")
		branchName := fmt.Sprintf("backup/%s", timestamp)
		
		candidates = append(candidates, &plugins.Candidate{
			Command:     fmt.Sprintf("git checkout -b %s && git add -A && git commit -m 'Backup commit'", branchName),
			Explanation: fmt.Sprintf("Creates a new backup branch '%s' and commits all changes", branchName),
			Breakdown: []plugins.Step{
				{Description: "Create and switch to new branch", Command: fmt.Sprintf("git checkout -b %s", branchName)},
				{Description: "Stage all changes", Command: "git add -A"},
				{Description: "Commit changes", Command: "git commit -m 'Backup commit'"},
			},
			Confidence:  90,
			RiskLevel:   plugins.RiskMedium,
			Destructive: false,
			DocLinks:    []string{"https://git-scm.com/docs/git-checkout", "https://git-scm.com/docs/git-commit"},
			UndoStrategy: &plugins.UndoStrategy{
				Type:        "git",
				Description: "Switch back to previous branch and delete backup branch",
				Command:     fmt.Sprintf("git checkout - && git branch -D %s", branchName),
			},
		})
	}
	
	// Pattern: commit all changes
	if matched, _ := regexp.MatchString(`(?i)commit\s+(all\s+)?changes?`, promptLower); matched {
		// Extract commit message if present
		messagePattern := regexp.MustCompile(`(?i)(?:with\s+message|message)\s+["\'](.+?)["\']`)
		matches := messagePattern.FindStringSubmatch(prompt)
		message := "Update files"
		if len(matches) > 1 {
			message = matches[1]
		}
		
		candidates = append(candidates, &plugins.Candidate{
			Command:     fmt.Sprintf("git add -A && git commit -m \"%s\"", message),
			Explanation: fmt.Sprintf("Stages all changes and commits with message: '%s'", message),
			Breakdown: []plugins.Step{
				{Description: "Stage all changes", Command: "git add -A"},
				{Description: "Commit changes", Command: fmt.Sprintf("git commit -m \"%s\"", message)},
			},
			Confidence:  95,
			RiskLevel:   plugins.RiskMedium,
			Destructive: false,
			DocLinks:    []string{"https://git-scm.com/docs/git-commit"},
		})
	}
	
	// Pattern: create branch
	if matched, _ := regexp.MatchString(`(?i)create\s+(?:new\s+)?branch`, promptLower); matched {
		branchPattern := regexp.MustCompile(`(?i)branch\s+(\S+)`)
		matches := branchPattern.FindStringSubmatch(prompt)
		branchName := "new-feature"
		if len(matches) > 1 {
			branchName = matches[1]
		}
		
		candidates = append(candidates, &plugins.Candidate{
			Command:     fmt.Sprintf("git checkout -b %s", branchName),
			Explanation: fmt.Sprintf("Creates and switches to new branch '%s'", branchName),
			Breakdown: []plugins.Step{
				{Description: "Create and checkout new branch", Command: fmt.Sprintf("git checkout -b %s", branchName)},
			},
			Confidence:  92,
			RiskLevel:   plugins.RiskSafe,
			Destructive: false,
			DocLinks:    []string{"https://git-scm.com/docs/git-checkout"},
		})
	}
	
	// Pattern: revert last commit
	if matched, _ := regexp.MatchString(`(?i)revert\s+last\s+commit`, promptLower); matched {
		candidates = append(candidates, &plugins.Candidate{
			Command:     "git reset --soft HEAD~1",
			Explanation: "Reverts the last commit but keeps changes staged",
			Breakdown: []plugins.Step{
				{Description: "Reset to previous commit, keep changes", Command: "git reset --soft HEAD~1"},
			},
			Confidence:      88,
			RiskLevel:       plugins.RiskMedium,
			Destructive:     false,
			RequiresConfirm: true,
			DocLinks:        []string{"https://git-scm.com/docs/git-reset"},
		})
	}
	
	// Pattern: delete branch
	if matched, _ := regexp.MatchString(`(?i)delete\s+branch`, promptLower); matched {
		branchPattern := regexp.MustCompile(`(?i)branch\s+(\S+)`)
		matches := branchPattern.FindStringSubmatch(prompt)
		branchName := "branch-name"
		if len(matches) > 1 {
			branchName = matches[1]
		}
		
		candidates = append(candidates, &plugins.Candidate{
			Command:         fmt.Sprintf("git branch -D %s", branchName),
			Explanation:     fmt.Sprintf("Force deletes branch '%s'", branchName),
			Breakdown:       []plugins.Step{{Description: "Delete branch", Command: fmt.Sprintf("git branch -D %s", branchName)}},
			Confidence:      85,
			RiskLevel:       plugins.RiskHigh,
			Destructive:     true,
			RequiresConfirm: true,
			DocLinks:        []string{"https://git-scm.com/docs/git-branch"},
		})
	}
	
	return candidates, nil
}

// PreRunCheck performs safety checks before Git command execution
func (p *GitPlugin) PreRunCheck(ctx plugins.Context, candidate *plugins.Candidate) (*plugins.CheckResult, error) {
	result := &plugins.CheckResult{
		Allowed:  true,
		Metadata: make(map[string]interface{}),
	}
	
	// Check if we're in a Git repository
	if !isGitRepo(ctx.WorkingDir) {
		result.Allowed = false
		result.Reason = "Not in a Git repository"
		return result, nil
	}
	
	// Check for uncommitted changes for destructive operations
	if candidate.Destructive {
		hasChanges, err := hasUncommittedChanges(ctx.WorkingDir)
		if err == nil && hasChanges {
			result.RequiresApproval = true
			result.ApprovalMessage = "Workspace has uncommitted changes. Type 'PROCEED WITH CHANGES' to continue"
			result.AdditionalChecks = append(result.AdditionalChecks, "uncommitted_changes")
			result.Metadata["uncommitted_changes"] = true
		}
	}
	
	// Check for force push or destructive Git operations
	if strings.Contains(candidate.Command, "push -f") || strings.Contains(candidate.Command, "push --force") {
		result.RequiresApproval = true
		result.ApprovalMessage = "Force push detected. Type 'FORCE PUSH' to confirm"
		result.AdditionalChecks = append(result.AdditionalChecks, "force_push")
	}
	
	// Add current branch to metadata
	if branch, err := getCurrentBranch(ctx.WorkingDir); err == nil {
		result.Metadata["current_branch"] = branch
	}
	
	return result, nil
}

// RequiresApproval checks if the candidate requires approval
func (p *GitPlugin) RequiresApproval(candidate *plugins.Candidate) bool {
	// Destructive Git operations require approval
	if candidate.Destructive {
		return true
	}
	
	// Force operations require approval
	if strings.Contains(candidate.Command, "-f") || strings.Contains(candidate.Command, "--force") {
		return true
	}
	
	return false
}

// Scopes returns required scopes
func (p *GitPlugin) Scopes() []string {
	return []string{"git:read", "git:write"}
}

// Helper functions

func isGitRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	return cmd.Run() == nil
}

func hasUncommittedChanges(dir string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

func getCurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
