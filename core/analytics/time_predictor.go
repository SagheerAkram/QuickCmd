package analytics

import (
	"fmt"
	"math"
	"time"
)

// TimePredictor predicts command execution time
type TimePredictor struct {
	history map[string][]int64 // command pattern -> durations (ms)
}

// NewTimePredictor creates a new time predictor
func NewTimePredictor() *TimePredictor {
	return &TimePredictor{
		history: make(map[string][]int64),
	}
}

// RecordExecution records a command execution time
func (tp *TimePredictor) RecordExecution(command string, duration time.Duration) {
	pattern := extractPattern(command)
	tp.history[pattern] = append(tp.history[pattern], duration.Milliseconds())
	
	// Keep only last 100 executions
	if len(tp.history[pattern]) > 100 {
		tp.history[pattern] = tp.history[pattern][1:]
	}
}

// Predict predicts execution time for a command
func (tp *TimePredictor) Predict(command string) *TimePrediction {
	pattern := extractPattern(command)
	durations := tp.history[pattern]
	
	if len(durations) == 0 {
		return &TimePrediction{
			Predicted:  0,
			Confidence: 0,
			Message:    "No historical data available",
		}
	}
	
	// Calculate statistics
	avg := average(durations)
	stdDev := standardDeviation(durations, avg)
	min := minimum(durations)
	max := maximum(durations)
	
	// Confidence based on sample size and variance
	confidence := calculateConfidence(len(durations), stdDev, avg)
	
	return &TimePrediction{
		Predicted:  time.Duration(avg) * time.Millisecond,
		Min:        time.Duration(min) * time.Millisecond,
		Max:        time.Duration(max) * time.Millisecond,
		StdDev:     time.Duration(stdDev) * time.Millisecond,
		Confidence: confidence,
		SampleSize: len(durations),
		Message:    formatPrediction(avg, confidence),
	}
}

// TimePrediction represents a time prediction
type TimePrediction struct {
	Predicted  time.Duration
	Min        time.Duration
	Max        time.Duration
	StdDev     time.Duration
	Confidence int
	SampleSize int
	Message    string
}

// Format formats the prediction for display
func (tp *TimePrediction) Format() string {
	if tp.Confidence == 0 {
		return tp.Message
	}
	
	return fmt.Sprintf(
		"⏱️  Estimated: %v (±%v)\n"+
			"   Range: %v - %v\n"+
			"   Confidence: %d%% (based on %d runs)",
		tp.Predicted.Round(time.Millisecond),
		tp.StdDev.Round(time.Millisecond),
		tp.Min.Round(time.Millisecond),
		tp.Max.Round(time.Millisecond),
		tp.Confidence,
		tp.SampleSize,
	)
}

// CompareActual compares prediction with actual execution
func (tp *TimePrediction) CompareActual(actual time.Duration) string {
	if tp.Confidence == 0 {
		return "No prediction available"
	}
	
	diff := actual - tp.Predicted
	percentage := float64(diff) / float64(tp.Predicted) * 100
	
	if math.Abs(percentage) < 10 {
		return fmt.Sprintf("✓ Within expected range (%.1f%% difference)", percentage)
	} else if diff > 0 {
		return fmt.Sprintf("⚠️  Slower than expected (+%.1f%%)", percentage)
	} else {
		return fmt.Sprintf("✓ Faster than expected (%.1f%%)", percentage)
	}
}

// WarnIfSlow checks if command is unusually slow
func (tp *TimePredictor) WarnIfSlow(command string, actual time.Duration) *SlowWarning {
	prediction := tp.Predict(command)
	
	if prediction.Confidence < 50 {
		return nil // Not enough data
	}
	
	// Warn if >2x slower than average
	threshold := prediction.Predicted * 2
	if actual > threshold {
		return &SlowWarning{
			Expected: prediction.Predicted,
			Actual:   actual,
			Slowdown: float64(actual) / float64(prediction.Predicted),
			Message:  fmt.Sprintf("Command took %.1fx longer than usual", float64(actual)/float64(prediction.Predicted)),
		}
	}
	
	return nil
}

// SlowWarning represents a slow execution warning
type SlowWarning struct {
	Expected time.Duration
	Actual   time.Duration
	Slowdown float64
	Message  string
}

// OptimizationSuggestion suggests optimizations for slow commands
type OptimizationSuggestion struct {
	Command     string
	CurrentTime time.Duration
	Suggestion  string
	Improvement string
}

// SuggestOptimizations suggests ways to speed up commands
func (tp *TimePredictor) SuggestOptimizations(command string, duration time.Duration) []*OptimizationSuggestion {
	suggestions := []*OptimizationSuggestion{}
	
	// Find commands with -exec \;
	if contains(command, "-exec") && contains(command, "\\;") {
		suggestions = append(suggestions, &OptimizationSuggestion{
			Command:     command,
			CurrentTime: duration,
			Suggestion:  "Use '-exec ... +' instead of '\\;'",
			Improvement: "Up to 10x faster for many files",
		})
	}
	
	// Find | grep patterns
	if contains(command, "|") && contains(command, "grep") {
		suggestions = append(suggestions, &OptimizationSuggestion{
			Command:     command,
			CurrentTime: duration,
			Suggestion:  "Use 'grep -r' instead of piping",
			Improvement: "2-3x faster for recursive search",
		})
	}
	
	// Slow find commands
	if contains(command, "find") && duration > 5*time.Second {
		suggestions = append(suggestions, &OptimizationSuggestion{
			Command:     command,
			CurrentTime: duration,
			Suggestion:  "Add '-maxdepth N' to limit search depth",
			Improvement: "Significantly faster for shallow searches",
		})
	}
	
	return suggestions
}

// Helper functions

func extractPattern(command string) string {
	// Extract base command pattern
	// e.g., "find . -name *.log" -> "find"
	parts := splitCommand(command)
	if len(parts) > 0 {
		return parts[0]
	}
	return command
}

func splitCommand(command string) []string {
	// Simple split by space
	result := []string{}
	current := ""
	inQuote := false
	
	for _, char := range command {
		if char == '"' || char == '\'' {
			inQuote = !inQuote
		} else if char == ' ' && !inQuote {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		result = append(result, current)
	}
	
	return result
}

func average(values []int64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := int64(0)
	for _, v := range values {
		sum += v
	}
	
	return float64(sum) / float64(len(values))
}

func standardDeviation(values []int64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	variance := 0.0
	for _, v := range values {
		diff := float64(v) - mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	
	return math.Sqrt(variance)
}

func minimum(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func maximum(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func calculateConfidence(sampleSize int, stdDev, mean float64) int {
	// Confidence based on sample size and coefficient of variation
	if sampleSize == 0 || mean == 0 {
		return 0
	}
	
	// More samples = higher confidence
	sizeConfidence := math.Min(float64(sampleSize)*10, 100)
	
	// Lower variance = higher confidence
	cv := stdDev / mean // Coefficient of variation
	varianceConfidence := math.Max(0, 100-cv*100)
	
	// Combined confidence
	return int((sizeConfidence + varianceConfidence) / 2)
}

func formatPrediction(avgMs float64, confidence int) string {
	duration := time.Duration(avgMs) * time.Millisecond
	
	if confidence >= 80 {
		return fmt.Sprintf("Expected: ~%v (high confidence)", duration.Round(time.Millisecond))
	} else if confidence >= 50 {
		return fmt.Sprintf("Expected: ~%v (medium confidence)", duration.Round(time.Millisecond))
	} else {
		return fmt.Sprintf("Expected: ~%v (low confidence)", duration.Round(time.Millisecond))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
