package k8s

import (
	"testing"
	"time"
	
	"github.com/yourusername/quickcmd/core/plugins"
)

func TestK8sPlugin_Translate(t *testing.T) {
	plugin := &K8sPlugin{}
	ctx := plugins.Context{
		WorkingDir: "/test",
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
			name:           "Scale deployment",
			prompt:         "scale deployment api to 5 replicas",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.Command == "kubectl scale deployment api --replicas=5" && c.RequiresConfirm
			},
		},
		{
			name:           "Get pods",
			prompt:         "get pods",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.Command == "kubectl get pods" && c.RiskLevel == plugins.RiskSafe
			},
		},
		{
			name:           "Get pods in namespace",
			prompt:         "get pods in namespace production",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.Command == "kubectl get pods -n production"
			},
		},
		{
			name:           "Delete pod",
			prompt:         "delete pod nginx-123",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.Command == "kubectl delete pod nginx-123" && c.Destructive
			},
		},
		{
			name:           "Apply manifest",
			prompt:         "apply manifest deployment.yaml",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.Command == "kubectl apply -f deployment.yaml" && c.RequiresConfirm
			},
		},
		{
			name:           "Describe deployment",
			prompt:         "describe deployment api",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.Command == "kubectl describe deployment api" && c.RiskLevel == plugins.RiskSafe
			},
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

func TestK8sPlugin_PreRunCheck(t *testing.T) {
	plugin := &K8sPlugin{}
	
	tests := []struct {
		name         string
		candidate    *plugins.Candidate
		wantAllowed  bool
		wantApproval bool
	}{
		{
			name: "Read operation",
			candidate: &plugins.Candidate{
				Command: "kubectl get pods",
				PluginMetadata: map[string]interface{}{
					"operation": "get",
				},
			},
			wantAllowed:  true,
			wantApproval: false,
		},
		{
			name: "Apply operation",
			candidate: &plugins.Candidate{
				Command: "kubectl apply -f manifest.yaml",
				PluginMetadata: map[string]interface{}{
					"operation": "apply",
				},
			},
			wantAllowed:  true,
			wantApproval: true,
		},
		{
			name: "Delete operation",
			candidate: &plugins.Candidate{
				Command: "kubectl delete pod nginx",
				PluginMetadata: map[string]interface{}{
					"operation": "delete",
				},
			},
			wantAllowed:  true,
			wantApproval: true,
		},
		{
			name: "Scale operation",
			candidate: &plugins.Candidate{
				Command: "kubectl scale deployment api --replicas=5",
				PluginMetadata: map[string]interface{}{
					"operation": "scale",
				},
			},
			wantAllowed:  true,
			wantApproval: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := plugins.Context{
				WorkingDir: "/test",
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
			
			if result.RequiresApproval != tt.wantApproval {
				t.Errorf("PreRunCheck() requires approval = %v, want %v", result.RequiresApproval, tt.wantApproval)
			}
		})
	}
}

func TestK8sPlugin_RequiresApproval(t *testing.T) {
	plugin := &K8sPlugin{}
	
	tests := []struct {
		name      string
		candidate *plugins.Candidate
		want      bool
	}{
		{
			name: "Destructive operation",
			candidate: &plugins.Candidate{
				Command:     "kubectl delete pod nginx",
				Destructive: true,
			},
			want: true,
		},
		{
			name: "Apply operation",
			candidate: &plugins.Candidate{
				Command: "kubectl apply -f manifest.yaml",
				PluginMetadata: map[string]interface{}{
					"operation": "apply",
				},
			},
			want: true,
		},
		{
			name: "Get operation",
			candidate: &plugins.Candidate{
				Command: "kubectl get pods",
				PluginMetadata: map[string]interface{}{
					"operation": "get",
				},
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
