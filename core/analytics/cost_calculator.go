package analytics

import (
	"fmt"
	"strings"
)

// CostCalculator estimates cloud operation costs
type CostCalculator struct {
	pricing map[string]*ResourcePricing
}

// ResourcePricing defines pricing for a resource type
type ResourcePricing struct {
	ResourceType string
	PricePerHour float64
	PricePerGB   float64
	PricePerOp   float64
	Currency     string
}

// CostEstimate represents a cost estimation
type CostEstimate struct {
	Command      string
	Resources    []*ResourceCost
	TotalCost    float64
	MonthlyCost  float64
	Currency     string
	Breakdown    []string
	Savings      []string
}

// ResourceCost represents cost for a specific resource
type ResourceCost struct {
	Type     string
	Quantity int
	Unit     string
	Cost     float64
}

// NewCostCalculator creates a new cost calculator
func NewCostCalculator() *CostCalculator {
	cc := &CostCalculator{
		pricing: make(map[string]*ResourcePricing),
	}
	cc.loadPricing()
	return cc
}

// loadPricing loads cloud pricing data
func (cc *CostCalculator) loadPricing() {
	// AWS EC2 pricing (simplified)
	cc.pricing["ec2-t3.micro"] = &ResourcePricing{
		ResourceType: "EC2 t3.micro",
		PricePerHour: 0.0104,
		Currency:     "USD",
	}
	cc.pricing["ec2-t3.small"] = &ResourcePricing{
		ResourceType: "EC2 t3.small",
		PricePerHour: 0.0208,
		Currency:     "USD",
	}
	cc.pricing["ec2-t3.medium"] = &ResourcePricing{
		ResourceType: "EC2 t3.medium",
		PricePerHour: 0.0416,
		Currency:     "USD",
	}
	
	// AWS S3 pricing
	cc.pricing["s3-storage"] = &ResourcePricing{
		ResourceType: "S3 Storage",
		PricePerGB:   0.023,
		Currency:     "USD",
	}
	
	// AWS RDS pricing
	cc.pricing["rds-db.t3.micro"] = &ResourcePricing{
		ResourceType: "RDS db.t3.micro",
		PricePerHour: 0.017,
		Currency:     "USD",
	}
	
	// Kubernetes node pricing (approximate)
	cc.pricing["k8s-node"] = &ResourcePricing{
		ResourceType: "Kubernetes Node",
		PricePerHour: 0.05,
		Currency:     "USD",
	}
}

// EstimateCost estimates cost for a command
func (cc *CostCalculator) EstimateCost(command string) *CostEstimate {
	estimate := &CostEstimate{
		Command:   command,
		Resources: []*ResourceCost{},
		Currency:  "USD",
		Breakdown: []string{},
		Savings:   []string{},
	}
	
	// AWS EC2 operations
	if contains(command, "aws ec2 run-instances") {
		instanceType := extractInstanceType(command)
		count := extractCount(command)
		
		pricing := cc.pricing["ec2-"+instanceType]
		if pricing != nil {
			hourlyCost := pricing.PricePerHour * float64(count)
			monthlyCost := hourlyCost * 730 // Average hours per month
			
			estimate.Resources = append(estimate.Resources, &ResourceCost{
				Type:     instanceType,
				Quantity: count,
				Unit:     "instances",
				Cost:     monthlyCost,
			})
			
			estimate.TotalCost += hourlyCost
			estimate.MonthlyCost += monthlyCost
			
			estimate.Breakdown = append(estimate.Breakdown,
				fmt.Sprintf("%d x %s: $%.2f/hour ($%.2f/month)",
					count, instanceType, hourlyCost, monthlyCost))
		}
	}
	
	// Kubernetes scaling
	if contains(command, "kubectl scale") && contains(command, "--replicas") {
		replicas := extractReplicas(command)
		nodeCost := cc.pricing["k8s-node"].PricePerHour * float64(replicas)
		monthlyCost := nodeCost * 730
		
		estimate.Resources = append(estimate.Resources, &ResourceCost{
			Type:     "Kubernetes Pods",
			Quantity: replicas,
			Unit:     "replicas",
			Cost:     monthlyCost,
		})
		
		estimate.TotalCost += nodeCost
		estimate.MonthlyCost += monthlyCost
		
		estimate.Breakdown = append(estimate.Breakdown,
			fmt.Sprintf("%d replicas: $%.2f/hour ($%.2f/month)",
				replicas, nodeCost, monthlyCost))
	}
	
	// Add savings suggestions
	estimate.Savings = cc.suggestSavings(command, estimate)
	
	return estimate
}

// suggestSavings suggests cost optimization opportunities
func (cc *CostCalculator) suggestSavings(command string, estimate *CostEstimate) []string {
	savings := []string{}
	
	// Suggest spot instances for EC2
	if contains(command, "aws ec2 run-instances") && !contains(command, "spot") {
		potentialSavings := estimate.MonthlyCost * 0.7 // 70% savings
		savings = append(savings,
			fmt.Sprintf("💰 Use spot instances to save ~$%.2f/month (70%% discount)", potentialSavings))
	}
	
	// Suggest reserved instances for long-running
	if estimate.MonthlyCost > 100 {
		potentialSavings := estimate.MonthlyCost * 0.4 // 40% savings
		savings = append(savings,
			fmt.Sprintf("💰 Use reserved instances to save ~$%.2f/month (40%% discount)", potentialSavings))
	}
	
	// Suggest autoscaling
	if contains(command, "kubectl scale") && !contains(command, "autoscale") {
		savings = append(savings,
			"💰 Use horizontal pod autoscaling to optimize costs based on load")
	}
	
	return savings
}

// Format formats the cost estimate for display
func (ce *CostEstimate) Format() string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("\n💰 Cost Estimate: %s\n\n", ce.Command))
	
	if ce.TotalCost == 0 {
		sb.WriteString("No cost information available for this command\n")
		return sb.String()
	}
	
	sb.WriteString("Breakdown:\n")
	for _, line := range ce.Breakdown {
		sb.WriteString(fmt.Sprintf("  • %s\n", line))
	}
	
	sb.WriteString(fmt.Sprintf("\nTotal: $%.2f/hour ($%.2f/month)\n", ce.TotalCost, ce.MonthlyCost))
	
	if len(ce.Savings) > 0 {
		sb.WriteString("\nCost Optimization Opportunities:\n")
		for _, saving := range ce.Savings {
			sb.WriteString(fmt.Sprintf("  %s\n", saving))
		}
	}
	
	return sb.String()
}

// CompareCosts compares costs of different approaches
func (cc *CostCalculator) CompareCosts(commands []string) *CostComparison {
	estimates := []*CostEstimate{}
	
	for _, cmd := range commands {
		estimate := cc.EstimateCost(cmd)
		estimates = append(estimates, estimate)
	}
	
	// Find cheapest
	cheapest := estimates[0]
	for _, est := range estimates {
		if est.MonthlyCost < cheapest.MonthlyCost {
			cheapest = est
		}
	}
	
	return &CostComparison{
		Estimates: estimates,
		Cheapest:  cheapest,
	}
}

// CostComparison compares multiple cost estimates
type CostComparison struct {
	Estimates []*CostEstimate
	Cheapest  *CostEstimate
}

// Helper functions

func extractInstanceType(command string) string {
	// Extract instance type from command
	// Simplified - would need proper parsing
	if contains(command, "t3.micro") {
		return "t3.micro"
	}
	if contains(command, "t3.small") {
		return "t3.small"
	}
	if contains(command, "t3.medium") {
		return "t3.medium"
	}
	return "t3.micro" // default
}

func extractCount(command string) int {
	// Extract count from --count flag
	// Simplified - would need proper parsing
	return 1
}

func extractReplicas(command string) int {
	// Extract replicas from --replicas flag
	// Simplified - would need proper parsing
	return 3
}

// BudgetAlert checks if operation exceeds budget
type BudgetAlert struct {
	Budget      float64
	EstimatedCost float64
	Exceeded    bool
	Message     string
}

// CheckBudget checks if cost exceeds budget
func (cc *CostCalculator) CheckBudget(command string, monthlyBudget float64) *BudgetAlert {
	estimate := cc.EstimateCost(command)
	
	exceeded := estimate.MonthlyCost > monthlyBudget
	
	alert := &BudgetAlert{
		Budget:        monthlyBudget,
		EstimatedCost: estimate.MonthlyCost,
		Exceeded:      exceeded,
	}
	
	if exceeded {
		overage := estimate.MonthlyCost - monthlyBudget
		alert.Message = fmt.Sprintf(
			"⚠️  This operation will exceed monthly budget by $%.2f (%.1f%%)",
			overage, (overage/monthlyBudget)*100)
	} else {
		remaining := monthlyBudget - estimate.MonthlyCost
		alert.Message = fmt.Sprintf(
			"✓ Within budget - $%.2f remaining (%.1f%% of budget)",
			remaining, (estimate.MonthlyCost/monthlyBudget)*100)
	}
	
	return alert
}
