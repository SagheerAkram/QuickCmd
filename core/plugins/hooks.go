package plugins

import (
	"context"
	"fmt"
)

// TranslateWithPlugins translates a prompt using both core templates and plugins
func TranslateWithPlugins(ctx Context, prompt string, coreCandidates []*Candidate) ([]*Candidate, error) {
	// Execute pre-translate hooks
	hookData := &TranslateHookData{
		Prompt:     prompt,
		Candidates: coreCandidates,
	}
	
	if err := DefaultRegistry().ExecuteHooks(HookPreTranslate, ctx, hookData); err != nil {
		return nil, fmt.Errorf("pre-translate hook failed: %w", err)
	}
	
	// Get all enabled plugins
	plugins := ListEnabled()
	
	// Collect candidates from all plugins
	allCandidates := make([]*Candidate, 0, len(coreCandidates)+len(plugins)*2)
	allCandidates = append(allCandidates, coreCandidates...)
	
	for _, plugin := range plugins {
		candidates, err := plugin.Translate(ctx, prompt)
		if err != nil {
			// Log error but continue with other plugins
			continue
		}
		
		// Add plugin name to each candidate
		for _, candidate := range candidates {
			candidate.PluginName = plugin.Name()
		}
		
		allCandidates = append(allCandidates, candidates...)
	}
	
	// Execute post-translate hooks
	hookData.Candidates = allCandidates
	if err := DefaultRegistry().ExecuteHooks(HookPostTranslate, ctx, hookData); err != nil {
		return nil, fmt.Errorf("post-translate hook failed: %w", err)
	}
	
	return allCandidates, nil
}

// PreRunCheckWithPlugins performs pre-run checks using plugins
func PreRunCheckWithPlugins(ctx Context, candidate *Candidate) (*CheckResult, error) {
	// If candidate has a plugin, use that plugin's PreRunCheck
	if candidate.PluginName != "" {
		plugin, err := Get(candidate.PluginName)
		if err != nil {
			return nil, err
		}
		
		return plugin.PreRunCheck(ctx, candidate)
	}
	
	// Otherwise, run checks from all enabled plugins
	result := &CheckResult{
		Allowed:  true,
		Metadata: make(map[string]interface{}),
	}
	
	plugins := ListEnabled()
	for _, plugin := range plugins {
		checkResult, err := plugin.PreRunCheck(ctx, candidate)
		if err != nil {
			continue
		}
		
		// If any plugin denies, deny overall
		if !checkResult.Allowed {
			result.Allowed = false
			result.Reason = checkResult.Reason
			return result, nil
		}
		
		// Accumulate approval requirements
		if checkResult.RequiresApproval {
			result.RequiresApproval = true
			if result.ApprovalMessage == "" {
				result.ApprovalMessage = checkResult.ApprovalMessage
			}
		}
		
		// Merge metadata
		for k, v := range checkResult.Metadata {
			result.Metadata[k] = v
		}
	}
	
	return result, nil
}

// ExecutePreExecutionHooks runs pre-execution hooks
func ExecutePreExecutionHooks(ctx Context, candidate *Candidate) error {
	hookData := &ExecutionHookData{
		Candidate: candidate,
	}
	
	return DefaultRegistry().ExecuteHooks(HookPreExecution, ctx, hookData)
}

// ExecutePostExecutionHooks runs post-execution hooks
func ExecutePostExecutionHooks(ctx Context, candidate *Candidate, sandboxID string, exitCode int, stdout, stderr []byte) error {
	hookData := &ExecutionHookData{
		Candidate: candidate,
		SandboxID: sandboxID,
		ExitCode:  exitCode,
		Stdout:    stdout,
		Stderr:    stderr,
	}
	
	return DefaultRegistry().ExecuteHooks(HookPostExecution, ctx, hookData)
}

// AugmentAuditMetadata adds plugin metadata to audit entries
func AugmentAuditMetadata(ctx Context, candidate *Candidate, metadata map[string]interface{}) error {
	hookData := &AuditHookData{
		Candidate: candidate,
		Metadata:  metadata,
	}
	
	return DefaultRegistry().ExecuteHooks(HookAuditMetadata, ctx, hookData)
}

// CheckApprovalRequired checks if any plugin requires approval for a candidate
func CheckApprovalRequired(candidate *Candidate) bool {
	if candidate.PluginName != "" {
		plugin, err := Get(candidate.PluginName)
		if err != nil {
			return false
		}
		
		return plugin.RequiresApproval(candidate)
	}
	
	// Check all enabled plugins
	plugins := ListEnabled()
	for _, plugin := range plugins {
		if plugin.RequiresApproval(candidate) {
			return true
		}
	}
	
	return false
}
