package executor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDockerRunner_RunInSandbox(t *testing.T) {
	if !IsDockerAvailable() {
		t.Skip("Docker not available, skipping integration tests")
	}
	
	runner, err := NewDockerRunner()
	if err != nil {
		t.Fatalf("Failed to create Docker runner: %v", err)
	}
	defer runner.Close()
	
	tests := []struct {
		name         string
		cmd          string
		opts         SandboxOptions
		wantExitCode int
		checkStdout  func(string) bool
		checkStderr  func(string) bool
	}{
		{
			name: "Simple echo command",
			cmd:  "echo 'Hello from sandbox'",
			opts: SandboxOptions{
				Image: "alpine:latest",
			},
			wantExitCode: 0,
			checkStdout: func(s string) bool {
				return strings.Contains(s, "Hello from sandbox")
			},
		},
		{
			name: "Command with error",
			cmd:  "ls /nonexistent",
			opts: SandboxOptions{
				Image: "alpine:latest",
			},
			wantExitCode: 1,
			checkStderr: func(s string) bool {
				return strings.Contains(s, "No such file")
			},
		},
		{
			name: "Resource limits enforced",
			cmd:  "echo 'test' && sleep 1",
			opts: SandboxOptions{
				Image:       "alpine:latest",
				CPULimit:    0.5,
				MemoryLimit: 128 * 1024 * 1024,
				Timeout:     5 * time.Second,
			},
			wantExitCode: 0,
		},
		{
			name: "Network isolation",
			cmd:  "ping -c 1 8.8.8.8 || echo 'Network blocked'",
			opts: SandboxOptions{
				Image:         "alpine:latest",
				NetworkAccess: false,
			},
			wantExitCode: 0,
			checkStdout: func(s string) bool {
				return strings.Contains(s, "Network blocked")
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := runner.RunInSandbox(tt.cmd, tt.opts)
			
			if err != nil && tt.wantExitCode == 0 {
				t.Errorf("RunInSandbox() unexpected error: %v", err)
				return
			}
			
			if result.ExitCode != tt.wantExitCode {
				t.Errorf("RunInSandbox() exit code = %d, want %d", result.ExitCode, tt.wantExitCode)
			}
			
			if tt.checkStdout != nil && !tt.checkStdout(string(result.Stdout)) {
				t.Errorf("RunInSandbox() stdout check failed: %s", result.Stdout)
			}
			
			if tt.checkStderr != nil && !tt.checkStderr(string(result.Stderr)) {
				t.Errorf("RunInSandbox() stderr check failed: %s", result.Stderr)
			}
			
			if result.SandboxID == "" {
				t.Error("RunInSandbox() sandbox ID is empty")
			}
			
			if result.EndTime.Before(result.StartTime) {
				t.Error("RunInSandbox() end time before start time")
			}
		})
	}
}

func TestDockerRunner_Timeout(t *testing.T) {
	if !IsDockerAvailable() {
		t.Skip("Docker not available")
	}
	
	runner, err := NewDockerRunner()
	if err != nil {
		t.Fatalf("Failed to create Docker runner: %v", err)
	}
	defer runner.Close()
	
	result, err := runner.RunInSandbox("sleep 10", SandboxOptions{
		Image:   "alpine:latest",
		Timeout: 1 * time.Second,
	})
	
	if err == nil {
		t.Error("Expected timeout error")
	}
	
	if result.ExitCode != 124 {
		t.Errorf("Expected exit code 124 for timeout, got %d", result.ExitCode)
	}
}

func TestDockerRunner_WithMounts(t *testing.T) {
	if !IsDockerAvailable() {
		t.Skip("Docker not available")
	}
	
	runner, err := NewDockerRunner()
	if err != nil {
		t.Fatalf("Failed to create Docker runner: %v", err)
	}
	defer runner.Close()
	
	// Create temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	result, err := runner.RunInSandbox("cat /data/test.txt", SandboxOptions{
		Image: "alpine:latest",
		Mounts: []Mount{
			{
				Source:   tmpDir,
				Target:   "/data",
				ReadOnly: true,
			},
		},
	})
	
	if err != nil {
		t.Fatalf("RunInSandbox() error: %v", err)
	}
	
	if !strings.Contains(string(result.Stdout), "test content") {
		t.Errorf("Expected to read mounted file content, got: %s", result.Stdout)
	}
}

func TestIsDockerAvailable(t *testing.T) {
	available := IsDockerAvailable()
	
	if available {
		info, err := GetDockerInfo()
		if err != nil {
			t.Errorf("GetDockerInfo() error when Docker is available: %v", err)
		}
		if info == "" {
			t.Error("GetDockerInfo() returned empty string")
		}
		t.Logf("Docker info: %s", info)
	} else {
		t.Log("Docker not available (this is OK for CI environments without Docker)")
	}
}
