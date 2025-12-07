package policy

import (
	"fmt"
	"regexp"
)

// Policy represents the security policy configuration
type Policy struct {
	Allowlist []Pattern        `yaml:"allowlist"`
	Denylist  []Pattern        `yaml:"denylist"`
	Approval  ApprovalConfig   `yaml:"approval"`
	Secrets   SecretsConfig    `yaml:"secrets"`
	Sandbox   SandboxConfig    `yaml:"sandbox"`
}

// Pattern represents a command pattern for matching
type Pattern struct {
	Pattern     string `yaml:"pattern"`
	Description string `yaml:"description"`
	compiled    *regexp.Regexp
}

// ApprovalConfig defines approval requirements
type ApprovalConfig struct {
	HighRisk          bool     `yaml:"high_risk"`
	RequireConfirm    bool     `yaml:"require_confirmation"`
	DestructiveOps    bool     `yaml:"destructive_ops"`
	AllowedUsers      []string `yaml:"allowed_users"`
	RequireMultiParty bool     `yaml:"require_multi_party"`
}

// SecretsConfig defines secrets handling
type SecretsConfig struct {
	RedactEnvVars bool     `yaml:"redact_env_vars"`
	RedactPatterns []string `yaml:"redact_patterns"`
	VaultIntegration bool   `yaml:"vault_integration"`
}

// SandboxConfig defines sandbox behavior
type SandboxConfig struct {
	Enabled       bool   `yaml:"enabled"`
	DefaultMode   string `yaml:"default_mode"` // "dry-run", "sandbox", "direct"
	NetworkAccess bool   `yaml:"network_access"`
	MaxCPU        string `yaml:"max_cpu"`
	MaxMemory     string `yaml:"max_memory"`
	MaxTime       int    `yaml:"max_time_seconds"`
}

// ValidationResult represents the result of policy validation
type ValidationResult struct {
	Allowed         bool
	Reason          string
	RequiresConfirm bool
	ConfirmMessage  string
	MatchedRule     string
}

// Compile compiles the regex pattern
func (p *Pattern) Compile() error {
	if p.compiled != nil {
		return nil
	}
	
	regex, err := regexp.Compile(p.Pattern)
	if err != nil {
		return fmt.Errorf("failed to compile pattern %q: %w", p.Pattern, err)
	}
	
	p.compiled = regex
	return nil
}

// Matches checks if the command matches this pattern
func (p *Pattern) Matches(command string) bool {
	if p.compiled == nil {
		if err := p.Compile(); err != nil {
			return false
		}
	}
	
	return p.compiled.MatchString(command)
}

// DefaultPolicy returns a sensible default policy
func DefaultPolicy() *Policy {
	return &Policy{
		Allowlist: []Pattern{},
		Denylist: []Pattern{
			{Pattern: `rm\s+-rf\s+/`, Description: "Prevent deletion of root directory"},
			{Pattern: `rm\s+-rf\s+/\*`, Description: "Prevent deletion of all root contents"},
			{Pattern: `:\(\)\s*\{\s*:\|:&\s*\};:`, Description: "Prevent fork bomb"},
			{Pattern: `shutdown`, Description: "Prevent system shutdown"},
			{Pattern: `reboot`, Description: "Prevent system reboot"},
			{Pattern: `mkfs`, Description: "Prevent filesystem formatting"},
			{Pattern: `dd\s+if=.*of=/dev/(sd|hd|nvme)`, Description: "Prevent disk overwrite"},
			{Pattern: `chmod\s+777\s+/`, Description: "Prevent dangerous permission changes"},
			{Pattern: `curl.*\|\s*bash`, Description: "Prevent piping curl to bash"},
			{Pattern: `wget.*\|\s*sh`, Description: "Prevent piping wget to shell"},
		},
		Approval: ApprovalConfig{
			HighRisk:       true,
			RequireConfirm: true,
			DestructiveOps: true,
		},
		Secrets: SecretsConfig{
			RedactEnvVars: true,
			RedactPatterns: []string{
				`password\s*=\s*[^\s]+`,
				`token\s*=\s*[^\s]+`,
				`api[_-]?key\s*=\s*[^\s]+`,
				`secret\s*=\s*[^\s]+`,
			},
		},
		Sandbox: SandboxConfig{
			Enabled:       true,
			DefaultMode:   "dry-run",
			NetworkAccess: false,
			MaxCPU:        "1.0",
			MaxMemory:     "512m",
			MaxTime:       300, // 5 minutes
		},
	}
}
