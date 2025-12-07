package analytics

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// ApprovalAnalytics analyzes approval workflow patterns
type ApprovalAnalytics struct {
	approvals []*ApprovalRecord
}

// ApprovalRecord represents an approval record
type ApprovalRecord struct {
	ID            string
	Command       string
	Requester     string
	Approver      string
	Status        string // "pending", "approved", "rejected"
	RequestedAt   time.Time
	RespondedAt   time.Time
	ResponseTime  time.Duration
	RiskLevel     string
}

// NewApprovalAnalytics creates new approval analytics
func NewApprovalAnalytics() *ApprovalAnalytics {
	return &ApprovalAnalytics{
		approvals: []*ApprovalRecord{},
	}
}

// AddApproval adds an approval record
func (aa *ApprovalAnalytics) AddApproval(record *ApprovalRecord) {
	aa.approvals = append(aa.approvals, record)
}

// GetMetrics calculates approval metrics
func (aa *ApprovalAnalytics) GetMetrics() *ApprovalMetrics {
	if len(aa.approvals) == 0 {
		return &ApprovalMetrics{}
	}
	
	metrics := &ApprovalMetrics{
		TotalApprovals: len(aa.approvals),
	}
	
	responseTimes := []time.Duration{}
	
	for _, approval := range aa.approvals {
		switch approval.Status {
		case "approved":
			metrics.Approved++
		case "rejected":
			metrics.Rejected++
		case "pending":
			metrics.Pending++
		}
		
		if !approval.RespondedAt.IsZero() {
			responseTimes = append(responseTimes, approval.ResponseTime)
		}
	}
	
	// Calculate average response time
	if len(responseTimes) > 0 {
		total := time.Duration(0)
		for _, rt := range responseTimes {
			total += rt
		}
		metrics.AvgResponseTime = total / time.Duration(len(responseTimes))
		
		// Find min/max
		metrics.MinResponseTime = responseTimes[0]
		metrics.MaxResponseTime = responseTimes[0]
		for _, rt := range responseTimes {
			if rt < metrics.MinResponseTime {
				metrics.MinResponseTime = rt
			}
			if rt > metrics.MaxResponseTime {
				metrics.MaxResponseTime = rt
			}
		}
	}
	
	// Calculate approval rate
	if metrics.TotalApprovals > 0 {
		metrics.ApprovalRate = float64(metrics.Approved) / float64(metrics.TotalApprovals) * 100
	}
	
	return metrics
}

// ApprovalMetrics represents approval workflow metrics
type ApprovalMetrics struct {
	TotalApprovals  int
	Approved        int
	Rejected        int
	Pending         int
	ApprovalRate    float64
	AvgResponseTime time.Duration
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
}

// Format formats metrics for display
func (am *ApprovalMetrics) Format() string {
	var sb strings.Builder
	
	sb.WriteString("\n📊 Approval Workflow Analytics\n\n")
	sb.WriteString(fmt.Sprintf("Total Approvals:   %d\n", am.TotalApprovals))
	sb.WriteString(fmt.Sprintf("  ✓ Approved:      %d (%.1f%%)\n", am.Approved, am.ApprovalRate))
	sb.WriteString(fmt.Sprintf("  ✗ Rejected:      %d (%.1f%%)\n", am.Rejected, 
		float64(am.Rejected)/float64(am.TotalApprovals)*100))
	sb.WriteString(fmt.Sprintf("  ⏳ Pending:       %d\n\n", am.Pending))
	
	if am.AvgResponseTime > 0 {
		sb.WriteString("Response Times:\n")
		sb.WriteString(fmt.Sprintf("  Average:         %v\n", am.AvgResponseTime.Round(time.Second)))
		sb.WriteString(fmt.Sprintf("  Fastest:         %v\n", am.MinResponseTime.Round(time.Second)))
		sb.WriteString(fmt.Sprintf("  Slowest:         %v\n", am.MaxResponseTime.Round(time.Second)))
	}
	
	return sb.String()
}

// GetApproverStats returns statistics per approver
func (aa *ApprovalAnalytics) GetApproverStats() map[string]*ApproverStats {
	stats := make(map[string]*ApproverStats)
	
	for _, approval := range aa.approvals {
		if approval.Approver == "" {
			continue
		}
		
		if stats[approval.Approver] == nil {
			stats[approval.Approver] = &ApproverStats{
				Approver: approval.Approver,
			}
		}
		
		s := stats[approval.Approver]
		s.TotalReviewed++
		
		if approval.Status == "approved" {
			s.Approved++
		} else if approval.Status == "rejected" {
			s.Rejected++
		}
		
		if !approval.RespondedAt.IsZero() {
			s.ResponseTimes = append(s.ResponseTimes, approval.ResponseTime)
		}
	}
	
	// Calculate averages
	for _, s := range stats {
		if len(s.ResponseTimes) > 0 {
			total := time.Duration(0)
			for _, rt := range s.ResponseTimes {
				total += rt
			}
			s.AvgResponseTime = total / time.Duration(len(s.ResponseTimes))
		}
		
		if s.TotalReviewed > 0 {
			s.ApprovalRate = float64(s.Approved) / float64(s.TotalReviewed) * 100
		}
	}
	
	return stats
}

// ApproverStats represents statistics for an approver
type ApproverStats struct {
	Approver        string
	TotalReviewed   int
	Approved        int
	Rejected        int
	ApprovalRate    float64
	AvgResponseTime time.Duration
	ResponseTimes   []time.Duration
}

// GetBottlenecks identifies approval bottlenecks
func (aa *ApprovalAnalytics) GetBottlenecks() []*Bottleneck {
	bottlenecks := []*Bottleneck{}
	
	// Find slow approvers
	approverStats := aa.GetApproverStats()
	for approver, stats := range approverStats {
		if stats.AvgResponseTime > 1*time.Hour {
			bottlenecks = append(bottlenecks, &Bottleneck{
				Type:        "slow_approver",
				Description: fmt.Sprintf("%s takes avg %v to respond", approver, stats.AvgResponseTime.Round(time.Minute)),
				Severity:    "medium",
				Suggestion:  "Consider adding backup approvers",
			})
		}
	}
	
	// Find high rejection rates
	for approver, stats := range approverStats {
		if stats.ApprovalRate < 50 && stats.TotalReviewed > 10 {
			bottlenecks = append(bottlenecks, &Bottleneck{
				Type:        "high_rejection",
				Description: fmt.Sprintf("%s rejects %.1f%% of requests", approver, 100-stats.ApprovalRate),
				Severity:    "high",
				Suggestion:  "Review policy rules - may be too restrictive",
			})
		}
	}
	
	// Find pending backlog
	pending := 0
	for _, approval := range aa.approvals {
		if approval.Status == "pending" {
			pending++
		}
	}
	
	if pending > 10 {
		bottlenecks = append(bottlenecks, &Bottleneck{
			Type:        "pending_backlog",
			Description: fmt.Sprintf("%d approvals pending", pending),
			Severity:    "high",
			Suggestion:  "Add more approvers or automate low-risk approvals",
		})
	}
	
	return bottlenecks
}

// Bottleneck represents an approval workflow bottleneck
type Bottleneck struct {
	Type        string
	Description string
	Severity    string
	Suggestion  string
}

// GetTrends analyzes approval trends over time
func (aa *ApprovalAnalytics) GetTrends() *ApprovalTrends {
	if len(aa.approvals) < 2 {
		return &ApprovalTrends{}
	}
	
	// Sort by time
	sorted := make([]*ApprovalRecord, len(aa.approvals))
	copy(sorted, aa.approvals)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].RequestedAt.Before(sorted[j].RequestedAt)
	})
	
	// Split into two halves
	mid := len(sorted) / 2
	firstHalf := sorted[:mid]
	secondHalf := sorted[mid:]
	
	// Calculate metrics for each half
	firstMetrics := calculateMetricsForSlice(firstHalf)
	secondMetrics := calculateMetricsForSlice(secondHalf)
	
	trends := &ApprovalTrends{
		ApprovalRateChange: secondMetrics.ApprovalRate - firstMetrics.ApprovalRate,
		ResponseTimeChange: secondMetrics.AvgResponseTime - firstMetrics.AvgResponseTime,
	}
	
	// Determine trend direction
	if trends.ApprovalRateChange > 5 {
		trends.ApprovalTrend = "improving"
	} else if trends.ApprovalRateChange < -5 {
		trends.ApprovalTrend = "declining"
	} else {
		trends.ApprovalTrend = "stable"
	}
	
	if trends.ResponseTimeChange > 10*time.Minute {
		trends.ResponseTimeTrend = "slowing"
	} else if trends.ResponseTimeChange < -10*time.Minute {
		trends.ResponseTimeTrend = "improving"
	} else {
		trends.ResponseTimeTrend = "stable"
	}
	
	return trends
}

// ApprovalTrends represents approval workflow trends
type ApprovalTrends struct {
	ApprovalRateChange  float64
	ResponseTimeChange  time.Duration
	ApprovalTrend       string
	ResponseTimeTrend   string
}

// Helper function
func calculateMetricsForSlice(approvals []*ApprovalRecord) *ApprovalMetrics {
	if len(approvals) == 0 {
		return &ApprovalMetrics{}
	}
	
	metrics := &ApprovalMetrics{
		TotalApprovals: len(approvals),
	}
	
	responseTimes := []time.Duration{}
	
	for _, approval := range approvals {
		if approval.Status == "approved" {
			metrics.Approved++
		}
		if !approval.RespondedAt.IsZero() {
			responseTimes = append(responseTimes, approval.ResponseTime)
		}
	}
	
	if metrics.TotalApprovals > 0 {
		metrics.ApprovalRate = float64(metrics.Approved) / float64(metrics.TotalApprovals) * 100
	}
	
	if len(responseTimes) > 0 {
		total := time.Duration(0)
		for _, rt := range responseTimes {
			total += rt
		}
		metrics.AvgResponseTime = total / time.Duration(len(responseTimes))
	}
	
	return metrics
}

// GenerateReport generates a comprehensive approval analytics report
func (aa *ApprovalAnalytics) GenerateReport() string {
	var sb strings.Builder
	
	// Overall metrics
	metrics := aa.GetMetrics()
	sb.WriteString(metrics.Format())
	sb.WriteString("\n")
	
	// Approver stats
	sb.WriteString("👥 Approver Performance:\n\n")
	approverStats := aa.GetApproverStats()
	for approver, stats := range approverStats {
		sb.WriteString(fmt.Sprintf("  %s:\n", approver))
		sb.WriteString(fmt.Sprintf("    Reviewed: %d\n", stats.TotalReviewed))
		sb.WriteString(fmt.Sprintf("    Approval Rate: %.1f%%\n", stats.ApprovalRate))
		sb.WriteString(fmt.Sprintf("    Avg Response: %v\n\n", stats.AvgResponseTime.Round(time.Minute)))
	}
	
	// Bottlenecks
	bottlenecks := aa.GetBottlenecks()
	if len(bottlenecks) > 0 {
		sb.WriteString("⚠️  Bottlenecks Detected:\n\n")
		for _, b := range bottlenecks {
			sb.WriteString(fmt.Sprintf("  • %s\n", b.Description))
			sb.WriteString(fmt.Sprintf("    💡 %s\n\n", b.Suggestion))
		}
	}
	
	// Trends
	trends := aa.GetTrends()
	sb.WriteString("📈 Trends:\n\n")
	sb.WriteString(fmt.Sprintf("  Approval Rate: %s\n", trends.ApprovalTrend))
	sb.WriteString(fmt.Sprintf("  Response Time: %s\n", trends.ResponseTimeTrend))
	
	return sb.String()
}
