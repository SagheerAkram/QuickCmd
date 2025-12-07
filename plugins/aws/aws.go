package aws

import (
	"fmt"
	"regexp"
	"strings"
	
	"github.com/yourusername/quickcmd/core/plugins"
)

// AWSPlugin handles AWS CLI command translations
type AWSPlugin struct {
	costThreshold float64 // Cost threshold for approval (in USD)
}

func init() {
	plugin := &AWSPlugin{
		costThreshold: 10.0, // Default $10 threshold
	}
	
	metadata := &plugins.PluginMetadata{
		Name:        "aws",
		Version:     "1.0.0",
		Description: "AWS CLI operations with cost estimation and credential protection",
		Author:      "QuickCMD Team",
		Scopes:      []string{"aws:read", "aws:write", "aws:admin"},
		Enabled:     true,
	}
	
	plugins.Register(plugin, metadata)
}

// Name returns the plugin name
func (p *AWSPlugin) Name() string {
	return "aws"
}

// Translate translates AWS-related prompts into AWS CLI commands
func (p *AWSPlugin) Translate(ctx plugins.Context, prompt string) ([]*plugins.Candidate, error) {
	promptLower := strings.ToLower(prompt)
	
	var candidates []*plugins.Candidate
	
	// Pattern: list EC2 instances
	if matched, _ := regexp.MatchString(`(?i)list\s+(?:ec2\s+)?instances?`, promptLower); matched {
		candidates = append(candidates, &plugins.Candidate{
			Command:     "aws ec2 describe-instances --query 'Reservations[*].Instances[*].[InstanceId,State.Name,InstanceType]' --output table",
			Explanation: "Lists all EC2 instances with their ID, state, and type",
			Breakdown: []plugins.Step{
				{Description: "Query EC2 instances", Command: "aws ec2 describe-instances"},
				{Description: "Format output as table", Command: "--output table"},
			},
			Confidence:     92,
			RiskLevel:      plugins.RiskSafe,
			NetworkTargets: []string{"ec2.amazonaws.com"},
			DocLinks:       []string{"https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-instances.html"},
			PluginMetadata: map[string]interface{}{
				"service":   "ec2",
				"operation": "describe-instances",
				"cost":      0.0,
			},
		})
	}
	
	// Pattern: increase/modify Auto Scaling Group
	if matched, _ := regexp.MatchString(`(?i)(?:increase|set|modify)\s+(?:asg|auto\s*scaling)`, promptLower); matched {
		asgPattern := regexp.MustCompile(`(?i)(?:asg|group)\s+(\S+).*?(\d+)`)
		matches := asgPattern.FindStringSubmatch(prompt)
		
		asgName := "my-asg"
		desiredCapacity := "5"
		if len(matches) > 2 {
			asgName = matches[1]
			desiredCapacity = matches[2]
		}
		
		// Estimate cost (rough heuristic)
		estimatedCost := 0.05 * parseFloat(desiredCapacity) // $0.05 per instance per hour
		
		candidates = append(candidates, &plugins.Candidate{
			Command:         fmt.Sprintf("aws autoscaling set-desired-capacity --auto-scaling-group-name %s --desired-capacity %s", asgName, desiredCapacity),
			Explanation:     fmt.Sprintf("Sets Auto Scaling Group '%s' desired capacity to %s instances", asgName, desiredCapacity),
			Breakdown:       []plugins.Step{{Description: "Update ASG capacity", Command: fmt.Sprintf("aws autoscaling set-desired-capacity --auto-scaling-group-name %s --desired-capacity %s", asgName, desiredCapacity)}},
			Confidence:      88,
			RiskLevel:       plugins.RiskHigh,
			Destructive:     false,
			RequiresConfirm: true,
			NetworkTargets:  []string{"autoscaling.amazonaws.com"},
			DocLinks:        []string{"https://docs.aws.amazon.com/cli/latest/reference/autoscaling/set-desired-capacity.html"},
			PluginMetadata: map[string]interface{}{
				"service":          "autoscaling",
				"operation":        "set-desired-capacity",
				"estimated_cost":   estimatedCost,
				"cost_unit":        "USD/hour",
				"desired_capacity": desiredCapacity,
			},
		})
	}
	
	// Pattern: list S3 buckets
	if matched, _ := regexp.MatchString(`(?i)list\s+(?:s3\s+)?buckets?`, promptLower); matched {
		candidates = append(candidates, &plugins.Candidate{
			Command:        "aws s3 ls",
			Explanation:    "Lists all S3 buckets in the account",
			Breakdown:      []plugins.Step{{Description: "List S3 buckets", Command: "aws s3 ls"}},
			Confidence:     95,
			RiskLevel:      plugins.RiskSafe,
			NetworkTargets: []string{"s3.amazonaws.com"},
			DocLinks:       []string{"https://docs.aws.amazon.com/cli/latest/reference/s3/ls.html"},
			PluginMetadata: map[string]interface{}{
				"service":   "s3",
				"operation": "list-buckets",
				"cost":      0.0,
			},
		})
	}
	
	// Pattern: create S3 bucket
	if matched, _ := regexp.MatchString(`(?i)create\s+(?:s3\s+)?bucket`, promptLower); matched {
		bucketPattern := regexp.MustCompile(`(?i)bucket\s+(\S+)`)
		matches := bucketPattern.FindStringSubmatch(prompt)
		
		bucketName := "my-bucket"
		if len(matches) > 1 {
			bucketName = matches[1]
		}
		
		candidates = append(candidates, &plugins.Candidate{
			Command:         fmt.Sprintf("aws s3 mb s3://%s", bucketName),
			Explanation:     fmt.Sprintf("Creates a new S3 bucket named '%s'", bucketName),
			Breakdown:       []plugins.Step{{Description: "Make S3 bucket", Command: fmt.Sprintf("aws s3 mb s3://%s", bucketName)}},
			Confidence:      90,
			RiskLevel:       plugins.RiskMedium,
			Destructive:     false,
			RequiresConfirm: true,
			NetworkTargets:  []string{"s3.amazonaws.com"},
			DocLinks:        []string{"https://docs.aws.amazon.com/cli/latest/reference/s3/mb.html"},
			PluginMetadata: map[string]interface{}{
				"service":        "s3",
				"operation":      "create-bucket",
				"estimated_cost": 0.023, // $0.023 per GB per month
				"cost_unit":      "USD/GB/month",
			},
		})
	}
	
	// Pattern: describe CloudFormation stack
	if matched, _ := regexp.MatchString(`(?i)describe\s+(?:cloudformation\s+)?stack`, promptLower); matched {
		stackPattern := regexp.MustCompile(`(?i)stack\s+(\S+)`)
		matches := stackPattern.FindStringSubmatch(prompt)
		
		stackName := "my-stack"
		if len(matches) > 1 {
			stackName = matches[1]
		}
		
		candidates = append(candidates, &plugins.Candidate{
			Command:        fmt.Sprintf("aws cloudformation describe-stacks --stack-name %s", stackName),
			Explanation:    fmt.Sprintf("Describes CloudFormation stack '%s'", stackName),
			Breakdown:      []plugins.Step{{Description: "Describe stack", Command: fmt.Sprintf("aws cloudformation describe-stacks --stack-name %s", stackName)}},
			Confidence:     93,
			RiskLevel:      plugins.RiskSafe,
			NetworkTargets: []string{"cloudformation.amazonaws.com"},
			DocLinks:       []string{"https://docs.aws.amazon.com/cli/latest/reference/cloudformation/describe-stacks.html"},
			PluginMetadata: map[string]interface{}{
				"service":   "cloudformation",
				"operation": "describe-stacks",
				"cost":      0.0,
			},
		})
	}
	
	return candidates, nil
}

// PreRunCheck performs safety checks before AWS command execution
func (p *AWSPlugin) PreRunCheck(ctx plugins.Context, candidate *plugins.Candidate) (*plugins.CheckResult, error) {
	result := &plugins.CheckResult{
		Allowed:  true,
		Metadata: make(map[string]interface{}),
	}
	
	// Check for cost threshold
	if candidate.PluginMetadata != nil {
		if cost, ok := candidate.PluginMetadata["estimated_cost"].(float64); ok {
			result.Metadata["estimated_cost"] = cost
			
			if cost > p.costThreshold {
				result.RequiresApproval = true
				result.ApprovalMessage = fmt.Sprintf("Estimated cost $%.2f exceeds threshold $%.2f. Type 'APPROVE COST' to confirm", cost, p.costThreshold)
				result.AdditionalChecks = append(result.AdditionalChecks, "cost_threshold")
			}
		}
	}
	
	// Sanitize command to avoid exposing secrets
	if strings.Contains(candidate.Command, "access-key") || strings.Contains(candidate.Command, "secret-key") {
		result.Allowed = false
		result.Reason = "Command contains sensitive credential parameters. Use AWS credential configuration instead."
		return result, nil
	}
	
	// Mark operations that create resources as requiring scoped credentials
	if candidate.PluginMetadata != nil {
		if operation, ok := candidate.PluginMetadata["operation"].(string); ok {
			creatingOps := []string{"create", "launch", "run", "start"}
			for _, op := range creatingOps {
				if strings.Contains(operation, op) {
					result.Metadata["requires_scoped_credentials"] = true
					result.AdditionalChecks = append(result.AdditionalChecks, "scoped_credentials")
					break
				}
			}
		}
	}
	
	// Add AWS region hint
	result.Metadata["aws_region"] = "current" // In production, get from AWS config
	
	return result, nil
}

// RequiresApproval checks if the candidate requires approval
func (p *AWSPlugin) RequiresApproval(candidate *plugins.Candidate) bool {
	// Operations that create resources require approval
	if candidate.PluginMetadata != nil {
		if operation, ok := candidate.PluginMetadata["operation"].(string); ok {
			creatingOps := []string{"create", "launch", "run", "start", "set", "modify"}
			for _, op := range creatingOps {
				if strings.Contains(operation, op) {
					return true
				}
			}
		}
		
		// Check cost threshold
		if cost, ok := candidate.PluginMetadata["estimated_cost"].(float64); ok {
			if cost > p.costThreshold {
				return true
			}
		}
	}
	
	return candidate.Destructive
}

// Scopes returns required scopes
func (p *AWSPlugin) Scopes() []string {
	return []string{"aws:read", "aws:write", "aws:admin"}
}

// Helper function
func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
