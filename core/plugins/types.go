package plugins

import (
	"context"
	"time"
)

// Plugin defines the interface that all QuickCMD plugins must implement
type Plugin interface {
	// Name returns the unique name of the plugin
	Name() string
	
	// Translate attempts to translate a natural language prompt into command candidates
	// Returns empty slice if the plugin doesn't handle this prompt
	Translate(ctx Context, prompt string) ([]*Candidate, error)
	
	// PreRunCheck performs safety checks before command execution
	// Can add additional approval requirements or deny execution
	PreRunCheck(ctx Context, candidate *Candidate) (*CheckResult, error)
	
	// RequiresApproval returns true if this candidate requires approval
	RequiresApproval(candidate *Candidate) bool
	
	// Scopes returns the required scopes/permissions for this plugin
	Scopes() []string
}

// Context provides execution context to plugins
type Context struct {
	WorkingDir string
	User       string
	Timestamp  time.Time
	Metadata   map[string]interface{}
}

// Candidate represents a command candidate with plugin metadata
type Candidate struct {
	Command        string
	Explanation    string
	Breakdown      []Step
	Confidence     int
	RiskLevel      Risk
	AffectedPaths  []string
	NetworkTargets []string
	Destructive    bool
	RequiresConfirm bool
	DocLinks       []string
	
	// Plugin-specific metadata
	PluginName     string
	PluginMetadata map[string]interface{}
	UndoStrategy   *UndoStrategy
}

// Step represents a single step in command breakdown
type Step struct {
	Description string
	Command     string
}

// Risk represents the risk level of a command
type Risk string

const (
	RiskSafe   Risk = "safe"
	RiskMedium Risk = "medium"
	RiskHigh   Risk = "high"
)

// CheckResult contains the result of a pre-run safety check
type CheckResult struct {
	Allowed         bool
	Reason          string
	RequiresApproval bool
	ApprovalMessage string
	AdditionalChecks []string
	Metadata        map[string]interface{}
}

// UndoStrategy describes how to undo a command
type UndoStrategy struct {
	Type        string // "git", "filesystem", "none"
	Description string
	Command     string
	Metadata    map[string]interface{}
}

// PluginError represents a plugin-specific error
type PluginError struct {
	PluginName string
	Operation  string
	Err        error
}

func (e *PluginError) Error() string {
	return "plugin " + e.PluginName + " " + e.Operation + ": " + e.Err.Error()
}

func (e *PluginError) Unwrap() error {
	return e.Err
}

// PluginMetadata contains plugin registration metadata
type PluginMetadata struct {
	Name        string
	Version     string
	Description string
	Author      string
	Scopes      []string
	Enabled     bool
}

// Hook types for plugin lifecycle
type HookType string

const (
	HookPreTranslate  HookType = "pre_translate"
	HookPostTranslate HookType = "post_translate"
	HookPreExecution  HookType = "pre_execution"
	HookPostExecution HookType = "post_execution"
	HookAuditMetadata HookType = "audit_metadata"
)

// HookFunc is a function that can be registered as a hook
type HookFunc func(ctx Context, data interface{}) error

// TranslateHookData contains data for translation hooks
type TranslateHookData struct {
	Prompt     string
	Candidates []*Candidate
}

// ExecutionHookData contains data for execution hooks
type ExecutionHookData struct {
	Candidate  *Candidate
	SandboxID  string
	ExitCode   int
	Stdout     []byte
	Stderr     []byte
}

// AuditHookData contains data for audit metadata hooks
type AuditHookData struct {
	Candidate *Candidate
	Metadata  map[string]interface{}
}
