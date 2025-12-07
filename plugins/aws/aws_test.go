package aws

import (
	"testing"
	"time"
	
	"github.com/yourusername/quickcmd/core/plugins"
)

func TestAWSPlugin_Translate(t *testing.T) {
	plugin := &AWSPlugin{costThreshold: 10.0}
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
			name:           "List EC2 instances",
			prompt:         "list ec2 instances",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.RiskLevel == plugins.RiskSafe && c.PluginMetadata["service"] == "ec2"
			},
		},
		{
			name:           "Increase ASG",
			prompt:         "increase asg my-asg to 5",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.RequiresConfirm && c.PluginMetadata["service"] == "autoscaling"
			},
		},
		{
			name:           "List S3 buckets",
			prompt:         "list s3 buckets",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.Command == "aws s3 ls" && c.RiskLevel == plugins.RiskSafe
			},
		},
		{
			name:           "Create S3 bucket",
			prompt:         "create s3 bucket my-bucket",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.Command == "aws s3 mb s3://my-bucket" && c.RequiresConfirm
			},
		},
		{
			name:           "Describe CloudFormation stack",
			prompt:         "describe cloudformation stack my-stack",
			wantCandidates: 1,
			checkCommand: func(c *plugins.Candidate) bool {
				return c.RiskLevel == plugins.RiskSafe && c.PluginMetadata["service"] == "cloudformation"
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

func TestAWSPlugin_PreRunCheck(t *testing.T) {
	plugin := &AWSPlugin{costThreshold: 10.0}
	
	tests := []struct {
		name         string
		candidate    *plugins.Candidate
		wantAllowed  bool
		wantApproval bool
		wantReason   string
	}{
		{
			name: "Read operation",
			candidate: &plugins.Candidate{
				Command: "aws ec2 describe-instances",
				PluginMetadata: map[string]interface{}{
					"operation":      "describe-instances",
					"estimated_cost": 0.0,
				},
			},
			wantAllowed:  true,
			wantApproval: false,
		},
		{
			name: "Cost below threshold",
			candidate: &plugins.Candidate{
				Command: "aws s3 mb s3://my-bucket",
				PluginMetadata: map[string]interface{}{
					"operation":      "create-bucket",
					"estimated_cost": 5.0,
				},
			},
			wantAllowed:  true,
			wantApproval: false,
		},
		{
			name: "Cost above threshold",
			candidate: &plugins.Candidate{
				Command: "aws autoscaling set-desired-capacity --auto-scaling-group-name my-asg --desired-capacity 100",
				PluginMetadata: map[string]interface{}{
					"operation":      "set-desired-capacity",
					"estimated_cost": 50.0,
				},
			},
			wantAllowed:  true,
			wantApproval: true,
		},
		{
			name: "Sensitive credentials in command",
			candidate: &plugins.Candidate{
				Command: "aws configure set aws_access_key_id AKIAIOSFODNN7EXAMPLE",
			},
			wantAllowed: false,
			wantReason:  "Command contains sensitive credential parameters. Use AWS credential configuration instead.",
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
			
			if tt.wantReason != "" && result.Reason != tt.wantReason {
				t.Errorf("PreRunCheck() reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestAWSPlugin_RequiresApproval(t *testing.T) {
	plugin := &AWSPlugin{costThreshold: 10.0}
	
	tests := []struct {
		name      string
		candidate *plugins.Candidate
		want      bool
	}{
		{
			name: "Create operation",
			candidate: &plugins.Candidate{
				Command: "aws s3 mb s3://my-bucket",
				PluginMetadata: map[string]interface{}{
					"operation":      "create-bucket",
					"estimated_cost": 5.0,
				},
			},
			want: true,
		},
		{
			name: "High cost operation",
			candidate: &plugins.Candidate{
				Command: "aws autoscaling set-desired-capacity",
				PluginMetadata: map[string]interface{}{
					"operation":      "set-desired-capacity",
					"estimated_cost": 50.0,
				},
			},
			want: true,
		},
		{
			name: "Read operation",
			candidate: &plugins.Candidate{
				Command: "aws ec2 describe-instances",
				PluginMetadata: map[string]interface{}{
					"operation":      "describe-instances",
					"estimated_cost": 0.0,
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
