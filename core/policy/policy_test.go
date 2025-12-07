package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEngine_Validate(t *testing.T) {
	engine := NewEngine()
	
	tests := []struct {
		name        string
		command     string
		riskLevel   string
		destructive bool
		wantAllowed bool
		wantConfirm bool
	}{
		{
			name:        "Safe command",
			command:     "ls -la",
			riskLevel:   "safe",
			destructive: false,
			wantAllowed: true,
			wantConfirm: false,
		},
		{
			name:        "Blocked - rm -rf /",
			command:     "rm -rf /",
			riskLevel:   "high",
			destructive: true,
			wantAllowed: false,
			wantConfirm: false,
		},
		{
			name:        "Blocked - fork bomb",
			command:     ":() { :|:& };:",
			riskLevel:   "high",
			destructive: true,
			wantAllowed: false,
			wantConfirm: false,
		},
		{
			name:        "High risk requires confirmation",
			command:     "docker system prune -a",
			riskLevel:   "high",
			destructive: true,
			wantAllowed: true,
			wantConfirm: true,
		},
		{
			name:        "Destructive requires confirmation",
			command:     "find . -name '*.log' -delete",
			riskLevel:   "medium",
			destructive: true,
			wantAllowed: true,
			wantConfirm: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Validate(tt.command, tt.riskLevel, tt.destructive)
			
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Validate() allowed = %v, want %v. Reason: %s", 
					result.Allowed, tt.wantAllowed, result.Reason)
			}
			
			if result.RequiresConfirm != tt.wantConfirm {
				t.Errorf("Validate() requiresConfirm = %v, want %v", 
					result.RequiresConfirm, tt.wantConfirm)
			}
		})
	}
}

func TestEngine_LoadFromFile(t *testing.T) {
	// Create a temporary policy file
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.yaml")
	
	policyYAML := `
allowlist:
  - pattern: "^ls "
    description: "Allow ls commands"
  - pattern: "^git status"
    description: "Allow git status"

denylist:
  - pattern: "rm -rf /"
    description: "Block root deletion"

approval:
  high_risk: true
  require_confirmation: true
`
	
	if err := os.WriteFile(policyFile, []byte(policyYAML), 0644); err != nil {
		t.Fatalf("Failed to create test policy file: %v", err)
	}
	
	engine, err := NewEngineFromFile(policyFile)
	if err != nil {
		t.Fatalf("NewEngineFromFile() error = %v", err)
	}
	
	// Test allowlist enforcement
	result := engine.Validate("ls -la", "safe", false)
	if !result.Allowed {
		t.Errorf("Expected 'ls -la' to be allowed, got: %s", result.Reason)
	}
	
	// Test command not in allowlist
	result = engine.Validate("echo hello", "safe", false)
	if result.Allowed {
		t.Error("Expected 'echo hello' to be blocked (not in allowlist)")
	}
	
	// Test denylist
	result = engine.Validate("rm -rf /", "high", true)
	if result.Allowed {
		t.Error("Expected 'rm -rf /' to be blocked by denylist")
	}
}

func TestEngine_AddPatterns(t *testing.T) {
	engine := NewEngine()
	
	// Add deny pattern
	err := engine.AddDenyPattern("^dangerous", "Block dangerous commands")
	if err != nil {
		t.Errorf("AddDenyPattern() error = %v", err)
	}
	
	result := engine.Validate("dangerous command", "high", false)
	if result.Allowed {
		t.Error("Expected newly added deny pattern to block command")
	}
	
	// Add allow pattern
	err = engine.AddAllowPattern("^safe", "Allow safe commands")
	if err != nil {
		t.Errorf("AddAllowPattern() error = %v", err)
	}
	
	// With allowlist, only matching commands are allowed
	result = engine.Validate("safe command", "safe", false)
	if !result.Allowed {
		t.Errorf("Expected command matching allowlist to be allowed, got: %s", result.Reason)
	}
}

func TestSecretRedactor_Redact(t *testing.T) {
	redactor := NewSecretRedactor()
	
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Redact password",
			input: "password=secret123",
			want:  "password=***REDACTED***",
		},
		{
			name:  "Redact API key",
			input: "api_key=abc123xyz",
			want:  "api_key=***REDACTED***",
		},
		{
			name:  "Redact Bearer token",
			input: "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			want:  "Authorization: ***REDACTED***",
		},
		{
			name:  "Safe command unchanged",
			input: "ls -la /home/user",
			want:  "ls -la /home/user",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactor.Redact(tt.input)
			if got != tt.want {
				t.Errorf("Redact() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSecretRedactor_RedactEnvVars(t *testing.T) {
	redactor := NewSecretRedactor()
	
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Redact PASSWORD env var",
			input: "PASSWORD=secret123 ./script.sh",
			want:  "PASSWORD=***REDACTED*** ./script.sh",
		},
		{
			name:  "Redact API_KEY env var",
			input: "API_KEY=abc123 curl example.com",
			want:  "API_KEY=***REDACTED*** curl example.com",
		},
		{
			name:  "Keep non-sensitive env vars",
			input: "DEBUG=true ./app",
			want:  "DEBUG=true ./app",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactor.RedactEnvVars(tt.input)
			if got != tt.want {
				t.Errorf("RedactEnvVars() = %q, want %q", got, tt.want)
			}
		})
	}
}
