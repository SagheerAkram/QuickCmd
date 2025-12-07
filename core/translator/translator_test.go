package translator

import (
	"testing"
)

func TestTranslator_Translate(t *testing.T) {
	translator := New()
	
	tests := []struct {
		name           string
		prompt         string
		wantErr        bool
		minCandidates  int
		maxCandidates  int
		expectedRisk   Risk
		expectedCmd    string
	}{
		{
			name:          "Find large files",
			prompt:        "find files larger than 100MB",
			wantErr:       false,
			minCandidates: 1,
			maxCandidates: 3,
			expectedRisk:  RiskSafe,
			expectedCmd:   "find . -type f -size +100M",
		},
		{
			name:          "Delete DS_Store files",
			prompt:        "delete all .DS_Store files",
			wantErr:       false,
			minCandidates: 1,
			maxCandidates: 3,
			expectedRisk:  RiskHigh,
		},
		{
			name:          "Archive old logs",
			prompt:        "archive logs older than 7 days",
			wantErr:       false,
			minCandidates: 1,
			maxCandidates: 3,
			expectedRisk:  RiskMedium,
		},
		{
			name:          "Git status",
			prompt:        "show git changes",
			wantErr:       false,
			minCandidates: 1,
			maxCandidates: 3,
			expectedRisk:  RiskSafe,
			expectedCmd:   "git status --short",
		},
		{
			name:          "Docker cleanup",
			prompt:        "cleanup docker containers",
			wantErr:       false,
			minCandidates: 1,
			maxCandidates: 3,
			expectedRisk:  RiskHigh,
		},
		{
			name:          "Disk usage",
			prompt:        "show disk usage",
			wantErr:       false,
			minCandidates: 1,
			maxCandidates: 3,
			expectedRisk:  RiskSafe,
		},
		{
			name:          "Find TODOs",
			prompt:        "find all TODO comments",
			wantErr:       false,
			minCandidates: 1,
			maxCandidates: 3,
			expectedRisk:  RiskSafe,
		},
		{
			name:    "Empty prompt",
			prompt:  "",
			wantErr: true,
		},
		{
			name:    "No match",
			prompt:  "xyzabc123 nonsense prompt that should not match anything",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates, err := translator.Translate(tt.prompt)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Translate() expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Translate() unexpected error: %v", err)
				return
			}
			
			if len(candidates) < tt.minCandidates {
				t.Errorf("Translate() got %d candidates, want at least %d", len(candidates), tt.minCandidates)
			}
			
			if len(candidates) > tt.maxCandidates {
				t.Errorf("Translate() got %d candidates, want at most %d", len(candidates), tt.maxCandidates)
			}
			
			// Check first candidate
			if len(candidates) > 0 {
				c := candidates[0]
				
				// Validate confidence is in range
				if c.Confidence < 0 || c.Confidence > 100 {
					t.Errorf("Candidate confidence %d out of range [0, 100]", c.Confidence)
				}
				
				// Check expected risk level if specified
				if tt.expectedRisk != "" && c.RiskLevel != tt.expectedRisk {
					t.Errorf("Expected risk %s, got %s", tt.expectedRisk, c.RiskLevel)
				}
				
				// Check expected command if specified
				if tt.expectedCmd != "" && c.Command != tt.expectedCmd {
					t.Errorf("Expected command %q, got %q", tt.expectedCmd, c.Command)
				}
				
				// Validate required fields
				if c.Command == "" {
					t.Error("Candidate command is empty")
				}
				if c.Explanation == "" {
					t.Error("Candidate explanation is empty")
				}
				if c.RiskLevel == "" {
					t.Error("Candidate risk level is empty")
				}
			}
		})
	}
}

func TestTranslator_AddTemplate(t *testing.T) {
	translator := New()
	initialCount := len(translator.templates)
	
	// Add a custom template
	translator.AddTemplate(&Template{
		Category:    "test",
		Description: "Test template",
	})
	
	if len(translator.templates) != initialCount+1 {
		t.Errorf("Expected %d templates, got %d", initialCount+1, len(translator.templates))
	}
}

func TestTranslator_ListCategories(t *testing.T) {
	translator := New()
	categories := translator.ListCategories()
	
	if len(categories) == 0 {
		t.Error("Expected at least one category")
	}
	
	// Check that categories are sorted
	for i := 1; i < len(categories); i++ {
		if categories[i-1] > categories[i] {
			t.Errorf("Categories not sorted: %v", categories)
			break
		}
	}
}

func TestCandidate_String(t *testing.T) {
	c := &Candidate{
		Command:     "test command",
		Explanation: "test explanation",
		Confidence:  95,
		RiskLevel:   RiskSafe,
		Breakdown: []Step{
			{Description: "Step 1", Command: "cmd1"},
			{Description: "Step 2", Command: "cmd2"},
		},
	}
	
	str := c.String()
	if str == "" {
		t.Error("Candidate String() returned empty string")
	}
	
	// Check that key information is included
	if !contains(str, "test command") {
		t.Error("String() missing command")
	}
	if !contains(str, "test explanation") {
		t.Error("String() missing explanation")
	}
	if !contains(str, "95") {
		t.Error("String() missing confidence")
	}
}

func TestCandidate_RiskIcon(t *testing.T) {
	tests := []struct {
		risk Risk
		want string
	}{
		{RiskSafe, "✅"},
		{RiskMedium, "⚠️"},
		{RiskHigh, "🔴"},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.risk), func(t *testing.T) {
			c := &Candidate{RiskLevel: tt.risk}
			if got := c.RiskIcon(); got != tt.want {
				t.Errorf("RiskIcon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != substr && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
