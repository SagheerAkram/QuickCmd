package k8s

import (
	"fmt"
	"regexp"
	"strings"
	
	"github.com/yourusername/quickcmd/core/plugins"
)

// K8sPlugin handles Kubernetes-related command translations
type K8sPlugin struct{}

func init() {
	plugin := &K8sPlugin{}
	metadata := &plugins.PluginMetadata{
		Name:        "k8s",
		Version:     "1.0.0",
		Description: "Kubernetes operations with RBAC checks and cluster state protection",
		Author:      "QuickCMD Team",
		Scopes:      []string{"k8s:read", "k8s:write", "k8s:admin"},
		Enabled:     true,
	}
	
	plugins.Register(plugin, metadata)
}

// Name returns the plugin name
func (p *K8sPlugin) Name() string {
	return "k8s"
}

// Translate translates Kubernetes-related prompts into kubectl commands
func (p *K8sPlugin) Translate(ctx plugins.Context, prompt string) ([]*plugins.Candidate, error) {
	promptLower := strings.ToLower(prompt)
	
	var candidates []*plugins.Candidate
	
	// Pattern: scale deployment
	if matched, _ := regexp.MatchString(`(?i)scale\s+deployment`, promptLower); matched {
		deployPattern := regexp.MustCompile(`(?i)deployment\s+(\S+).*?(\d+)\s+replicas?`)
		matches := deployPattern.FindStringSubmatch(prompt)
		
		deploymentName := "deployment-name"
		replicas := "3"
		if len(matches) > 2 {
			deploymentName = matches[1]
			replicas = matches[2]
		}
		
		candidates = append(candidates, &plugins.Candidate{
			Command:     fmt.Sprintf("kubectl scale deployment %s --replicas=%s", deploymentName, replicas),
			Explanation: fmt.Sprintf("Scales deployment '%s' to %s replicas", deploymentName, replicas),
			Breakdown: []plugins.Step{
				{Description: "Scale deployment", Command: fmt.Sprintf("kubectl scale deployment %s --replicas=%s", deploymentName, replicas)},
			},
			Confidence:      90,
			RiskLevel:       plugins.RiskMedium,
			Destructive:     false,
			RequiresConfirm: true,
			DocLinks:        []string{"https://kubernetes.io/docs/reference/kubectl/cheatsheet/#scaling-resources"},
			PluginMetadata: map[string]interface{}{
				"resource_type": "deployment",
				"operation":     "scale",
				"replicas":      replicas,
			},
		})
	}
	
	// Pattern: get pods
	if matched, _ := regexp.MatchString(`(?i)(?:get|list|show)\s+pods?`, promptLower); matched {
		namespace := "default"
		nsPattern := regexp.MustCompile(`(?i)(?:in|from)\s+namespace\s+(\S+)`)
		if matches := nsPattern.FindStringSubmatch(prompt); len(matches) > 1 {
			namespace = matches[1]
		}
		
		cmd := "kubectl get pods"
		if namespace != "default" {
			cmd += fmt.Sprintf(" -n %s", namespace)
		}
		
		candidates = append(candidates, &plugins.Candidate{
			Command:     cmd,
			Explanation: fmt.Sprintf("Lists all pods in namespace '%s'", namespace),
			Breakdown: []plugins.Step{
				{Description: "Get pods", Command: cmd},
			},
			Confidence: 95,
			RiskLevel:  plugins.RiskSafe,
			DocLinks:   []string{"https://kubernetes.io/docs/reference/kubectl/cheatsheet/#viewing-finding-resources"},
			PluginMetadata: map[string]interface{}{
				"resource_type": "pod",
				"operation":     "get",
				"namespace":     namespace,
			},
		})
	}
	
	// Pattern: delete pod
	if matched, _ := regexp.MatchString(`(?i)delete\s+pod`, promptLower); matched {
		podPattern := regexp.MustCompile(`(?i)pod\s+(\S+)`)
		matches := podPattern.FindStringSubmatch(prompt)
		
		podName := "pod-name"
		if len(matches) > 1 {
			podName = matches[1]
		}
		
		candidates = append(candidates, &plugins.Candidate{
			Command:         fmt.Sprintf("kubectl delete pod %s", podName),
			Explanation:     fmt.Sprintf("Deletes pod '%s'", podName),
			Breakdown:       []plugins.Step{{Description: "Delete pod", Command: fmt.Sprintf("kubectl delete pod %s", podName)}},
			Confidence:      88,
			RiskLevel:       plugins.RiskHigh,
			Destructive:     true,
			RequiresConfirm: true,
			DocLinks:        []string{"https://kubernetes.io/docs/reference/kubectl/cheatsheet/#deleting-resources"},
			PluginMetadata: map[string]interface{}{
				"resource_type": "pod",
				"operation":     "delete",
			},
		})
	}
	
	// Pattern: apply manifest
	if matched, _ := regexp.MatchString(`(?i)apply\s+(?:manifest|config|yaml)`, promptLower); matched {
		filePattern := regexp.MustCompile(`(?i)(?:file|manifest)\s+(\S+\.ya?ml)`)
		matches := filePattern.FindStringSubmatch(prompt)
		
		fileName := "manifest.yaml"
		if len(matches) > 1 {
			fileName = matches[1]
		}
		
		candidates = append(candidates, &plugins.Candidate{
			Command:         fmt.Sprintf("kubectl apply -f %s", fileName),
			Explanation:     fmt.Sprintf("Applies Kubernetes manifest from '%s'", fileName),
			Breakdown:       []plugins.Step{{Description: "Apply manifest", Command: fmt.Sprintf("kubectl apply -f %s", fileName)}},
			Confidence:      92,
			RiskLevel:       plugins.RiskHigh,
			Destructive:     false,
			RequiresConfirm: true,
			AffectedPaths:   []string{fileName},
			DocLinks:        []string{"https://kubernetes.io/docs/reference/kubectl/cheatsheet/#apply"},
			PluginMetadata: map[string]interface{}{
				"operation": "apply",
				"file":      fileName,
			},
		})
	}
	
	// Pattern: describe resource
	if matched, _ := regexp.MatchString(`(?i)describe\s+(?:pod|deployment|service)`, promptLower); matched {
		resourcePattern := regexp.MustCompile(`(?i)describe\s+(pod|deployment|service)\s+(\S+)`)
		matches := resourcePattern.FindStringSubmatch(prompt)
		
		resourceType := "pod"
		resourceName := "resource-name"
		if len(matches) > 2 {
			resourceType = matches[1]
			resourceName = matches[2]
		}
		
		candidates = append(candidates, &plugins.Candidate{
			Command:     fmt.Sprintf("kubectl describe %s %s", resourceType, resourceName),
			Explanation: fmt.Sprintf("Shows detailed information about %s '%s'", resourceType, resourceName),
			Breakdown:   []plugins.Step{{Description: "Describe resource", Command: fmt.Sprintf("kubectl describe %s %s", resourceType, resourceName)}},
			Confidence:  94,
			RiskLevel:   plugins.RiskSafe,
			DocLinks:    []string{"https://kubernetes.io/docs/reference/kubectl/cheatsheet/#viewing-finding-resources"},
			PluginMetadata: map[string]interface{}{
				"resource_type": resourceType,
				"operation":     "describe",
			},
		})
	}
	
	return candidates, nil
}

// PreRunCheck performs safety checks before Kubernetes command execution
func (p *K8sPlugin) PreRunCheck(ctx plugins.Context, candidate *plugins.Candidate) (*plugins.CheckResult, error) {
	result := &plugins.CheckResult{
		Allowed:  true,
		Metadata: make(map[string]interface{}),
	}
	
	// Check if kubectl is available
	// (In production, you'd actually check this)
	
	// Flag operations that alter cluster state as high-risk
	if candidate.PluginMetadata != nil {
		if operation, ok := candidate.PluginMetadata["operation"].(string); ok {
			alteringOps := []string{"apply", "delete", "scale", "patch", "edit"}
			for _, op := range alteringOps {
				if operation == op {
					result.RequiresApproval = true
					result.ApprovalMessage = fmt.Sprintf("Kubernetes %s operation requires approval. Type 'K8S %s' to confirm", strings.ToUpper(operation), strings.ToUpper(operation))
					result.AdditionalChecks = append(result.AdditionalChecks, "cluster_state_change")
					result.Metadata["alters_cluster_state"] = true
					break
				}
			}
		}
	}
	
	// Add RBAC hint
	result.Metadata["rbac_required"] = true
	result.Metadata["kube_context"] = "current" // In production, get actual context
	
	return result, nil
}

// RequiresApproval checks if the candidate requires approval
func (p *K8sPlugin) RequiresApproval(candidate *plugins.Candidate) bool {
	// All destructive operations require approval
	if candidate.Destructive {
		return true
	}
	
	// Operations that alter cluster state require approval
	if candidate.PluginMetadata != nil {
		if operation, ok := candidate.PluginMetadata["operation"].(string); ok {
			alteringOps := []string{"apply", "delete", "scale", "patch", "edit", "create"}
			for _, op := range alteringOps {
				if operation == op {
					return true
				}
			}
		}
	}
	
	return false
}

// Scopes returns required scopes
func (p *K8sPlugin) Scopes() []string {
	return []string{"k8s:read", "k8s:write", "k8s:admin"}
}
