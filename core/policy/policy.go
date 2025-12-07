package policy

import (
	"fmt"
	"os"
	
	"gopkg.in/yaml.v3"
)

// Engine handles policy enforcement
type Engine struct {
	policy *Policy
}

// NewEngine creates a new policy engine with default policy
func NewEngine() *Engine {
	return &Engine{
		policy: DefaultPolicy(),
	}
}

// NewEngineFromFile creates a policy engine from a YAML file
func NewEngineFromFile(path string) (*Engine, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}
	
	var policy Policy
	if err := yaml.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse policy YAML: %w", err)
	}
	
	// Compile all patterns
	for i := range policy.Allowlist {
		if err := policy.Allowlist[i].Compile(); err != nil {
			return nil, err
		}
	}
	for i := range policy.Denylist {
		if err := policy.Denylist[i].Compile(); err != nil {
			return nil, err
		}
	}
	
	return &Engine{policy: &policy}, nil
}

// Validate checks if a command is allowed by the policy
func (e *Engine) Validate(command string, riskLevel string, destructive bool) *ValidationResult {
	// Check denylist first (highest priority)
	for _, pattern := range e.policy.Denylist {
		if pattern.Matches(command) {
			return &ValidationResult{
				Allowed:     false,
				Reason:      fmt.Sprintf("Command blocked by denylist: %s", pattern.Description),
				MatchedRule: pattern.Pattern,
			}
		}
	}
	
	// If allowlist is defined and not empty, command must match allowlist
	if len(e.policy.Allowlist) > 0 {
		allowed := false
		var matchedRule string
		
		for _, pattern := range e.policy.Allowlist {
			if pattern.Matches(command) {
				allowed = true
				matchedRule = pattern.Pattern
				break
			}
		}
		
		if !allowed {
			return &ValidationResult{
				Allowed: false,
				Reason:  "Command not in allowlist",
			}
		}
		
		// Command is in allowlist, check if confirmation needed
		result := &ValidationResult{
			Allowed:     true,
			MatchedRule: matchedRule,
		}
		
		e.applyApprovalRules(result, riskLevel, destructive)
		return result
	}
	
	// No allowlist defined, command passes denylist, check approval requirements
	result := &ValidationResult{
		Allowed: true,
	}
	
	e.applyApprovalRules(result, riskLevel, destructive)
	return result
}

// applyApprovalRules applies approval requirements based on risk and destructiveness
func (e *Engine) applyApprovalRules(result *ValidationResult, riskLevel string, destructive bool) {
	// Check if high-risk commands require confirmation
	if e.policy.Approval.HighRisk && riskLevel == "high" {
		result.RequiresConfirm = true
		result.ConfirmMessage = "This is a HIGH RISK operation. Type 'I UNDERSTAND' to proceed"
	}
	
	// Check if destructive operations require confirmation
	if e.policy.Approval.DestructiveOps && destructive {
		result.RequiresConfirm = true
		if result.ConfirmMessage == "" {
			result.ConfirmMessage = "This is a DESTRUCTIVE operation. Type 'CONFIRM DELETE' to proceed"
		}
	}
	
	// General confirmation requirement
	if e.policy.Approval.RequireConfirm && result.ConfirmMessage == "" {
		result.RequiresConfirm = true
		result.ConfirmMessage = "Type 'CONFIRM' to proceed"
	}
}

// GetPolicy returns the current policy
func (e *Engine) GetPolicy() *Policy {
	return e.policy
}

// SetPolicy sets a new policy
func (e *Engine) SetPolicy(policy *Policy) {
	e.policy = policy
}

// SavePolicy saves the current policy to a YAML file
func (e *Engine) SavePolicy(path string) error {
	data, err := yaml.Marshal(e.policy)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write policy file: %w", err)
	}
	
	return nil
}

// AddDenyPattern adds a pattern to the denylist
func (e *Engine) AddDenyPattern(pattern, description string) error {
	p := Pattern{
		Pattern:     pattern,
		Description: description,
	}
	
	if err := p.Compile(); err != nil {
		return err
	}
	
	e.policy.Denylist = append(e.policy.Denylist, p)
	return nil
}

// AddAllowPattern adds a pattern to the allowlist
func (e *Engine) AddAllowPattern(pattern, description string) error {
	p := Pattern{
		Pattern:     pattern,
		Description: description,
	}
	
	if err := p.Compile(); err != nil {
		return err
	}
	
	e.policy.Allowlist = append(e.policy.Allowlist, p)
	return nil
}
