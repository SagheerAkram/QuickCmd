package agent

import (
	"context"
	"fmt"
	"time"
	
	"github.com/SagheerAkram/QuickCmd/core/audit"
	"github.com/SagheerAkram/QuickCmd/core/executor"
	"github.com/SagheerAkram/QuickCmd/core/plugins"
	"github.com/SagheerAkram/QuickCmd/core/policy"
)

// JobExecutor executes jobs using the sandbox runner
type JobExecutor struct {
	config       *Config
	dockerRunner *executor.DockerRunner
	policyEngine *policy.Engine
	auditStore   *audit.SQLiteStore
	snapshotter  *executor.Snapshotter
}

// NewJobExecutor creates a new job executor
func NewJobExecutor(config *Config) (*JobExecutor, error) {
	// Create Docker runner
	dockerRunner, err := executor.NewDockerRunner()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker runner: %w", err)
	}
	
	// Load policy engine
	policyEngine, err := policy.NewEngine(policy.DefaultPolicy())
	if err != nil {
		return nil, fmt.Errorf("failed to create policy engine: %w", err)
	}
	
	// Create audit store
	auditStore, err := audit.NewSQLiteStore(config.AuditDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit store: %w", err)
	}
	
	// Create snapshotter
	snapshotter := executor.NewSnapshotter()
	
	return &JobExecutor{
		config:       config,
		dockerRunner: dockerRunner,
		policyEngine: policyEngine,
		auditStore:   auditStore,
		snapshotter:  snapshotter,
	}, nil
}

// Execute executes a job and streams logs
func (e *JobExecutor) Execute(ctx context.Context, payload *JobPayload, logChan chan<- *LogFrame) (*JobResult, error) {
	result := &JobResult{
		JobID:     payload.JobID,
		StartTime: time.Now(),
	}
	
	// Send initial log
	e.sendLog(logChan, payload.JobID, "stdout", "Starting job execution...")
	
	// Validate against policy engine
	e.sendLog(logChan, payload.JobID, "stdout", "Validating command against policy...")
	if err := e.policyEngine.Validate(payload.Command); err != nil {
		e.sendLog(logChan, payload.JobID, "stderr", fmt.Sprintf("Policy validation failed: %v", err))
		result.Error = err.Error()
		return result, err
	}
	
	// Plugin pre-run checks
	e.sendLog(logChan, payload.JobID, "stdout", "Running plugin safety checks...")
	pluginCtx := plugins.Context{
		WorkingDir: "/workspace",
		User:       "agent",
		Timestamp:  time.Now(),
		Metadata:   payload.PluginMetadata,
	}
	
	candidate := &plugins.Candidate{
		Command:        payload.Command,
		PluginMetadata: payload.PluginMetadata,
	}
	
	checkResult, err := plugins.PreRunCheckWithPlugins(pluginCtx, candidate)
	if err != nil {
		e.sendLog(logChan, payload.JobID, "stderr", fmt.Sprintf("Plugin check failed: %v", err))
		result.Error = err.Error()
		return result, err
	}
	
	if !checkResult.Allowed {
		e.sendLog(logChan, payload.JobID, "stderr", fmt.Sprintf("Plugin denied execution: %s", checkResult.Reason))
		result.Error = checkResult.Reason
		return result, fmt.Errorf("plugin denied: %s", checkResult.Reason)
	}
	
	// Create snapshot if needed
	var snapshot *executor.SnapshotMetadata
	if candidate.Destructive {
		e.sendLog(logChan, payload.JobID, "stdout", "Creating pre-run snapshot...")
		snap, err := e.snapshotter.CreateSnapshot("/workspace", candidate.AffectedPaths)
		if err != nil {
			e.sendLog(logChan, payload.JobID, "stderr", fmt.Sprintf("Snapshot creation failed: %v", err))
		} else {
			snapshot = snap
			e.sendLog(logChan, payload.JobID, "stdout", fmt.Sprintf("Snapshot created: %s", snap.Location))
		}
	}
	
	// Configure sandbox options
	opts := executor.SandboxOptions{
		Image:         e.config.DefaultImage,
		CPULimit:      e.config.DefaultCPULimit,
		MemoryLimit:   e.config.DefaultMemoryLimit,
		PidsLimit:     64,
		NetworkAccess: false,
		ReadOnly:      false,
		Timeout:       time.Duration(e.config.DefaultTimeout) * time.Second,
		Mounts: []executor.Mount{
			{
				Source:   "/tmp/quickcmd-workspace",
				Target:   "/workspace",
				ReadOnly: false,
			},
		},
	}
	
	// Execute in sandbox
	e.sendLog(logChan, payload.JobID, "stdout", "Executing command in sandbox...")
	sandboxResult, err := e.dockerRunner.RunInSandbox(payload.Command, opts)
	
	result.EndTime = time.Now()
	result.DurationMs = result.EndTime.Sub(result.StartTime).Milliseconds()
	
	if err != nil {
		e.sendLog(logChan, payload.JobID, "stderr", fmt.Sprintf("Execution failed: %v", err))
		result.Error = err.Error()
		result.ExitCode = sandboxResult.ExitCode
		result.Stdout = string(sandboxResult.Stdout)
		result.Stderr = string(sandboxResult.Stderr)
		result.SandboxID = sandboxResult.SandboxID
	} else {
		result.SandboxID = sandboxResult.SandboxID
		result.ExitCode = sandboxResult.ExitCode
		result.Stdout = string(sandboxResult.Stdout)
		result.Stderr = string(sandboxResult.Stderr)
		
		// Stream stdout
		if len(sandboxResult.Stdout) > 0 {
			e.sendLog(logChan, payload.JobID, "stdout", string(sandboxResult.Stdout))
		}
		
		// Stream stderr
		if len(sandboxResult.Stderr) > 0 {
			e.sendLog(logChan, payload.JobID, "stderr", string(sandboxResult.Stderr))
		}
		
		if result.ExitCode == 0 {
			e.sendLog(logChan, payload.JobID, "stdout", "Command executed successfully")
		} else {
			e.sendLog(logChan, payload.JobID, "stderr", fmt.Sprintf("Command failed with exit code %d", result.ExitCode))
		}
	}
	
	// Store snapshot metadata
	if snapshot != nil {
		result.Snapshot = audit.EncodeSnapshot(snapshot)
	}
	
	// Log to audit database
	auditRecord := &audit.RunRecord{
		Timestamp:       time.Now().Format(time.RFC3339),
		User:            "agent",
		Prompt:          payload.Prompt,
		SelectedCommand: payload.Command,
		SandboxID:       result.SandboxID,
		ExitCode:        result.ExitCode,
		Stdout:          []byte(result.Stdout),
		Stderr:          []byte(result.Stderr),
		RiskLevel:       "medium", // TODO: Get from candidate
		Snapshot:        result.Snapshot,
		Executed:        true,
		DurationMs:      result.DurationMs,
	}
	
	if err := e.auditStore.LogExecution(auditRecord); err != nil {
		e.sendLog(logChan, payload.JobID, "stderr", fmt.Sprintf("Failed to log execution: %v", err))
	}
	
	return result, nil
}

// sendLog sends a log frame to the channel
func (e *JobExecutor) sendLog(logChan chan<- *LogFrame, jobID, stream, data string) {
	select {
	case logChan <- &LogFrame{
		JobID:     jobID,
		Timestamp: time.Now(),
		Stream:    stream,
		Data:      data,
		Final:     false,
	}:
	default:
		// Channel full, skip
	}
}

// Close closes the executor resources
func (e *JobExecutor) Close() error {
	if e.dockerRunner != nil {
		e.dockerRunner.Close()
	}
	if e.auditStore != nil {
		e.auditStore.Close()
	}
	return nil
}
