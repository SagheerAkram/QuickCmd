package git

import (
	"testing"
	"time"
	
	"github.com/yourusername/quickcmd/core/plugins"
)

func TestGitPlugin_Translate(t *testing.T) {
	plugin := &GitPlugin{}
	ctx := plugins.Context{
		WorkingDir: "/test/repo",
		User:       "testuser",
		Timestamp:  time.Now(),
	}
	
	tests := []struct {
		name           string
		prompt         string
		wantCandidates int
		checkCommand   func(*plugins.Candidate) bool
	}{
		{
			name:           "Create backup branch and commit",
			prompt:         "create backup branch and commit changes",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.Command != "" && c.UndoStrategy != nil
			},
		},
		{
			name:           "Commit all changes",
			prompt:         "commit all changes",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.Command == "git add -A && git commit -m \"Update files\""
			},
		},
		{
			name:           "Commit with message",
			prompt:         "commit all changes with message \"Fix bug\"",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.Command == "git add -A && git commit -m \"Fix bug\""
			},
		},
		{
			name:           "Create branch",
			prompt:         "create branch feature-x",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.Command == "git checkout -b feature-x"
			},
		},
		{
			name:           "Revert last commit",
			prompt:         "revert last commit",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.Command == "git reset --soft HEAD~1" && c.RequiresConfirm
			},
		},
		{
			name:           "Delete branch",
			prompt:         "delete branch old-feature",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.Command == "git branch -D old-feature" && c.Destructive
			},
		},
		{
			name:           "Non-git prompt",
			prompt:         "find large files",
			wantCandidates: 0,
			checkCommand:   nil,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates, err := plugin.Translate(ctx, tt.prompt)
			if err != nil {
				t.Fatalf("Translate() error = %v", err)
			}
			
			if len(candidates) != tt.wantCandidates {
				t.Errorf("Translate() returned %d candidates, want %d", len(candidates), tt.wantCandidates)
			}
			
			if tt.checkCommand != nil && len(candidates) > 0 {
				if !tt.checkCommand(candidates[0]) {
					t.Errorf("Translate() candidate check failed for: %s", candidates[0].Command)
				}
			}
		})
	}
}

func TestGitPlugin_PreRunCheck(t *testing.T) {
	plugin := &GitPlugin{}
	
	tests := []struct {
		name              string
		candidate         *plugins.Candidate
		wantAllowed       bool
		wantApproval      bool
		wantReason        string
	}{
		{
			name: "Safe operation",
			candidate: &plugins.Candidate{
				Command:     "git status",
				Destructive: false,
			},
			wantAllowed:  true,
			wantApproval: false,
		},
		{
			name: "Destructive operation",
			candidate: &plugins.Candidate{
				Command:     "git branch -D feature",
				Destructive: true,
			},
			wantAllowed:  true,
			wantApproval: false, // May require approval if uncommitted changes
		},
		{
			name: "Force push",
			candidate: &plugins.Candidate{
				Command:     "git push -f origin main",
				Destructive: false,
			},
			wantAllowed:  true,
			wantApproval: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := plugins.Context{
				WorkingDir: "/test/repo",
				User:       "testuser",
				Timestamp:  time.Now(),
			}
			
			result, err := plugin.PreRunCheck(ctx, tt.candidate)
			if err != nil {
				t.Fatalf("PreRunCheck() error = %v", err)
			}
			
			if result.Allowed != tt.wantAllowed {
				t.Errorf("PreRunCheck() allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			
			if tt.wantApproval && !result.RequiresApproval {
				t.Errorf("PreRunCheck() should require approval")
			}
			
			if tt.wantReason != "" && result.Reason != tt.wantReason {
				t.Errorf("PreRunCheck() reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestGitPlugin_RequiresApproval(t *testing.T) {
	plugin := &GitPlugin{}
	
	tests := []struct {
		name      string
		candidate *plugins.Candidate
		want      bool
	}{
		{
			name: "Destructive operation",
			candidate: &plugins.Candidate{
				Command:     "git branch -D feature",
				Destructive: true,
			},
			want: true,
		},
		{
			name: "Force operation",
			candidate: &plugins.Candidate{
				Command:     "git push -f",
				Destructive: false,
			},
			want: true,
		},
		{
			name: "Safe operation",
			candidate: &plugins.Candidate{
				Command:     "git status",
				Destructive: false,
			},
			want: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := plugin.RequiresApproval(tt.candidate)
			if got != tt.want {
				t.Errorf("RequiresApproval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitPlugin_Scopes(t *testing.T) {
	plugin := &GitPlugin{}
	scopes := plugin.Scopes()
	
	if len(scopes) != 2 {
		t.Errorf("Scopes() returned %d scopes, want 2", len(scopes))
	}
	
	expectedScopes := map[string]bool{
		"git:read":  true,
		"git:write": true,
	}
	
	for _, scope := range scopes {
		if !expectedScopes[scope] {
			t.Errorf("Unexpected scope: %s", scope)
		}
	}
}
