package optimizer

import (
	"fmt"
	"strings"
)

// PerformanceOptimizer suggests command optimizations
type PerformanceOptimizer struct {
	rules []*OptimizationRule
}

// OptimizationRule represents an optimization rule
type OptimizationRule struct {
	Name        string
	Pattern     string
	Replacement string
	Improvement string
	Speedup     float64
}

// NewPerformanceOptimizer creates a new optimizer
func NewPerformanceOptimizer() *PerformanceOptimizer {
	po := &PerformanceOptimizer{
		rules: []*OptimizationRule{},
	}
	po.loadRules()
	return po
}

// loadRules loads optimization rules
func (po *PerformanceOptimizer) loadRules() {
	po.rules = append(po.rules,
		&OptimizationRule{
			Name:        "find-exec-optimization",
			Pattern:     "-exec.*\\;",
			Replacement: "-exec ... +",
			Improvement: "Runs command once with all files instead of once per file",
			Speedup:     10.0,
		},
		&OptimizationRule{
			Name:        "grep-recursive",
			Pattern:     "find.*|.*grep",
			Replacement: "grep -r",
			Improvement: "Built-in recursive search is faster",
			Speedup:     3.0,
		},
	)
}

// Optimize suggests optimizations for a command
func (po *PerformanceOptimizer) Optimize(command string) []*Optimization {
	optimizations := []*Optimization{}
	
	for _, rule := range po.rules {
		if strings.Contains(command, rule.Pattern) {
			optimizations = append(optimizations, &Optimization{
				Rule:        rule,
				Original:    command,
				Optimized:   applyRule(command, rule),
				Improvement: rule.Improvement,
				Speedup:     rule.Speedup,
			})
		}
	}
	
	return optimizations
}

// Optimization represents an optimization suggestion
type Optimization struct {
	Rule        *OptimizationRule
	Original    string
	Optimized   string
	Improvement string
	Speedup     float64
}

// Format formats the optimization for display
func (o *Optimization) Format() string {
	return fmt.Sprintf(
		"⚡ Optimization: %s\n"+
			"   Original:  %s\n"+
			"   Optimized: %s\n"+
			"   Speedup:   %.1fx faster\n"+
			"   Why:       %s",
		o.Rule.Name,
		o.Original,
		o.Optimized,
		o.Speedup,
		o.Improvement,
	)
}

func applyRule(command string, rule *OptimizationRule) string {
	return strings.ReplaceAll(command, rule.Pattern, rule.Replacement)
}
