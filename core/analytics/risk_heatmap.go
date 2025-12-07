package analytics

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// RiskHeatmap provides visual risk analysis
type RiskHeatmap struct {
	data map[string]map[string]int // [category][risk_level] = count
}

// NewRiskHeatmap creates a new risk heatmap
func NewRiskHeatmap() *RiskHeatmap {
	return &RiskHeatmap{
		data: make(map[string]map[string]int),
	}
}

// AddCommand adds a command to the heatmap
func (rh *RiskHeatmap) AddCommand(category, riskLevel string) {
	if rh.data[category] == nil {
		rh.data[category] = make(map[string]int)
	}
	rh.data[category][riskLevel]++
}

// Visualize creates a visual heatmap
func (rh *RiskHeatmap) Visualize() string {
	var sb strings.Builder
	
	sb.WriteString("\n📊 Risk Assessment Heatmap\n\n")
	sb.WriteString("Category          │ Safe │ Medium │ High │ Total\n")
	sb.WriteString("──────────────────┼──────┼────────┼──────┼──────\n")
	
	// Sort categories
	categories := []string{}
	for cat := range rh.data {
		categories = append(categories, cat)
	}
	sort.Strings(categories)
	
	for _, cat := range categories {
		risks := rh.data[cat]
		safe := risks["safe"]
		medium := risks["medium"]
		high := risks["high"]
		total := safe + medium + high
		
		// Color coding
		safeBar := colorBar(safe, total, "green")
		mediumBar := colorBar(medium, total, "yellow")
		highBar := colorBar(high, total, "red")
		
		sb.WriteString(fmt.Sprintf("%-18s│ %4d │ %6d │ %4d │ %5d\n",
			cat, safe, medium, high, total))
		sb.WriteString(fmt.Sprintf("                  │ %s │ %s │ %s │\n",
			safeBar, mediumBar, highBar))
	}
	
	return sb.String()
}

// GetRiskScore calculates overall risk score (0-100)
func (rh *RiskHeatmap) GetRiskScore() int {
	totalCommands := 0
	riskPoints := 0
	
	for _, risks := range rh.data {
		safe := risks["safe"]
		medium := risks["medium"]
		high := risks["high"]
		
		totalCommands += safe + medium + high
		riskPoints += medium*50 + high*100 // Medium=50pts, High=100pts
	}
	
	if totalCommands == 0 {
		return 0
	}
	
	return riskPoints / totalCommands
}

// GetTrends analyzes risk trends over time
func (rh *RiskHeatmap) GetTrends(historical []*RiskHeatmap) *RiskTrend {
	if len(historical) < 2 {
		return &RiskTrend{Direction: "stable"}
	}
	
	current := rh.GetRiskScore()
	previous := historical[len(historical)-2].GetRiskScore()
	
	change := current - previous
	direction := "stable"
	if change > 5 {
		direction = "increasing"
	} else if change < -5 {
		direction = "decreasing"
	}
	
	return &RiskTrend{
		Current:   current,
		Previous:  previous,
		Change:    change,
		Direction: direction,
	}
}

// RiskTrend represents risk trend analysis
type RiskTrend struct {
	Current   int
	Previous  int
	Change    int
	Direction string
}

// TimelineHeatmap shows risk over time
type TimelineHeatmap struct {
	hourly  map[int]*RiskHeatmap
	daily   map[string]*RiskHeatmap
	weekly  map[string]*RiskHeatmap
}

// NewTimelineHeatmap creates a timeline heatmap
func NewTimelineHeatmap() *TimelineHeatmap {
	return &TimelineHeatmap{
		hourly: make(map[int]*RiskHeatmap),
		daily:  make(map[string]*RiskHeatmap),
		weekly: make(map[string]*RiskHeatmap),
	}
}

// AddCommand adds a command with timestamp
func (th *TimelineHeatmap) AddCommand(timestamp time.Time, category, riskLevel string) {
	hour := timestamp.Hour()
	day := timestamp.Format("2006-01-02")
	week := timestamp.Format("2006-W01")
	
	// Hourly
	if th.hourly[hour] == nil {
		th.hourly[hour] = NewRiskHeatmap()
	}
	th.hourly[hour].AddCommand(category, riskLevel)
	
	// Daily
	if th.daily[day] == nil {
		th.daily[day] = NewRiskHeatmap()
	}
	th.daily[day].AddCommand(category, riskLevel)
	
	// Weekly
	if th.weekly[week] == nil {
		th.weekly[week] = NewRiskHeatmap()
	}
	th.weekly[week].AddCommand(category, riskLevel)
}

// VisualizeHourly shows hourly risk pattern
func (th *TimelineHeatmap) VisualizeHourly() string {
	var sb strings.Builder
	
	sb.WriteString("\n⏰ Hourly Risk Pattern\n\n")
	
	for hour := 0; hour < 24; hour++ {
		heatmap := th.hourly[hour]
		if heatmap == nil {
			continue
		}
		
		score := heatmap.GetRiskScore()
		bar := strings.Repeat("█", score/5)
		
		sb.WriteString(fmt.Sprintf("%02d:00 %s %d\n", hour, bar, score))
	}
	
	return sb.String()
}

// Helper functions

func colorBar(value, total int, color string) string {
	if total == 0 {
		return "    "
	}
	
	percentage := (value * 100) / total
	bars := percentage / 10
	
	return strings.Repeat("█", bars)
}

// CategoryAnalysis provides detailed category analysis
type CategoryAnalysis struct {
	Category      string
	TotalCommands int
	SafeCount     int
	MediumCount   int
	HighCount     int
	RiskScore     int
	Trend         string
	TopCommands   []string
}

// AnalyzeCategory analyzes a specific category
func (rh *RiskHeatmap) AnalyzeCategory(category string) *CategoryAnalysis {
	risks := rh.data[category]
	if risks == nil {
		return nil
	}
	
	safe := risks["safe"]
	medium := risks["medium"]
	high := risks["high"]
	total := safe + medium + high
	
	riskScore := 0
	if total > 0 {
		riskScore = (medium*50 + high*100) / total
	}
	
	return &CategoryAnalysis{
		Category:      category,
		TotalCommands: total,
		SafeCount:     safe,
		MediumCount:   medium,
		HighCount:     high,
		RiskScore:     riskScore,
	}
}

// GetHighRiskCategories returns categories with high risk
func (rh *RiskHeatmap) GetHighRiskCategories() []string {
	highRisk := []string{}
	
	for category := range rh.data {
		analysis := rh.AnalyzeCategory(category)
		if analysis.RiskScore > 50 {
			highRisk = append(highRisk, category)
		}
	}
	
	return highRisk
}

// Recommendations provides security recommendations
func (rh *RiskHeatmap) Recommendations() []string {
	recommendations := []string{}
	
	highRiskCats := rh.GetHighRiskCategories()
	if len(highRiskCats) > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("🔴 High risk detected in: %s", strings.Join(highRiskCats, ", ")))
		recommendations = append(recommendations,
			"Consider adding more restrictive policies")
	}
	
	overallScore := rh.GetRiskScore()
	if overallScore > 60 {
		recommendations = append(recommendations,
			"⚠️  Overall risk score is high - review recent commands")
	} else if overallScore < 20 {
		recommendations = append(recommendations,
			"✓ Good security posture - risk score is low")
	}
	
	return recommendations
}
