package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/spf13/cobra"
	"github.com/yourusername/quickcmd/core/audit"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View command execution history",
	Long: `Displays the audit log of previously executed commands.
	
Shows timestamp, prompt, command, exit code, and undo options where available.`,
	RunE: showHistory,
}

func init() {
	historyCmd.Flags().IntP("limit", "n", 20, "number of entries to show")
	historyCmd.Flags().StringP("filter", "f", "", "filter by command or prompt")
	historyCmd.Flags().Bool("stats", false, "show statistics instead of history")
}

func showHistory(cmd *cobra.Command, args []string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	filter, _ := cmd.Flags().GetString("filter")
	showStats, _ := cmd.Flags().GetBool("stats")
	
	// Open audit database
	dbPath := getAuditDBPath()
	store, err := audit.NewSQLiteStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open audit database: %w", err)
	}
	defer store.Close()
	
	if showStats {
		return displayStats(store)
	}
	
	// Get history
	records, err := store.GetHistory(limit, filter)
	if err != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}
	
	if len(records) == 0 {
		fmt.Println("No execution history found.")
		fmt.Println("\nExecute commands with --sandbox to start building history.")
		return nil
	}
	
	// Display header
	fmt.Printf("%sCommand History%s", colorBold, colorReset)
	if filter != "" {
		fmt.Printf(" (filtered by: %s)", filter)
	}
	fmt.Printf("\n\n")
	
	// Display records
	for i, record := range records {
		displayRecord(i+1, record)
		if i < len(records)-1 {
			fmt.Println(strings.Repeat("─", 80))
		}
	}
	
	fmt.Printf("\n%sShowing %d of %d total executions%s\n", 
		colorBold, len(records), len(records), colorReset)
	
	return nil
}

func displayRecord(num int, record *audit.RunRecord) {
	// Parse timestamp
	timestamp, _ := time.Parse(time.RFC3339, record.Timestamp)
	
	// Header with number and timestamp
	fmt.Printf("%s%d. %s%s", colorBold, num, timestamp.Format("2006-01-02 15:04:05"), colorReset)
	
	// Risk level indicator
	riskColor := colorGreen
	riskIcon := "✅"
	switch record.RiskLevel {
	case "medium":
		riskColor = colorYellow
		riskIcon = "⚠️"
	case "high":
		riskColor = colorRed
		riskIcon = "🔴"
	}
	fmt.Printf(" %s%s %s%s\n", riskColor, riskIcon, record.RiskLevel, colorReset)
	
	// Prompt
	if record.Prompt != "" {
		fmt.Printf("   Prompt: %s\n", record.Prompt)
	}
	
	// Command
	fmt.Printf("   %sCommand:%s %s\n", colorBold, colorReset, record.SelectedCommand)
	
	// Execution details
	if record.Executed {
		exitIcon := "✓"
		exitColor := colorGreen
		if record.ExitCode != 0 {
			exitIcon = "✗"
			exitColor = colorRed
		}
		
		fmt.Printf("   %s%s Exit Code: %d%s", exitColor, exitIcon, record.ExitCode, colorReset)
		
		if record.DurationMs > 0 {
			fmt.Printf(" | Duration: %dms", record.DurationMs)
		}
		
		if record.SandboxID != "" {
			fmt.Printf(" | Sandbox: %s", record.SandboxID)
		}
		
		fmt.Println()
	} else {
		fmt.Printf("   %s⊘ Not executed (dry-run)%s\n", colorYellow, colorReset)
	}
	
	// Output preview
	if len(record.Stdout) > 0 {
		preview := string(record.Stdout)
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		fmt.Printf("   Output: %s\n", strings.TrimSpace(preview))
	}
	
	// Snapshot info
	if record.Snapshot != "" {
		snapshot, err := audit.DecodeSnapshot(record.Snapshot)
		if err == nil && snapshot.Reversible {
			fmt.Printf("   %s↶ Undo: %s%s\n", colorYellow, snapshot.RestoreCmd, colorReset)
		}
	}
	
	fmt.Println()
}

func displayStats(store *audit.SQLiteStore) error {
	stats, err := store.GetStats()
	if err != nil {
		return fmt.Errorf("failed to get statistics: %w", err)
	}
	
	fmt.Printf("%sExecution Statistics%s\n\n", colorBold, colorReset)
	
	// Total executions
	total, _ := stats["total_executions"].(int64)
	fmt.Printf("Total Executions: %s%d%s\n", colorBold, total, colorReset)
	
	// Success rate
	if successRate, ok := stats["success_rate"].(float64); ok {
		fmt.Printf("Success Rate: %s%.1f%%%s\n", colorGreen, successRate, colorReset)
	}
	
	// By risk level
	if riskCounts, ok := stats["by_risk_level"].(map[string]int64); ok {
		fmt.Printf("\n%sBy Risk Level:%s\n", colorBold, colorReset)
		for risk, count := range riskCounts {
			riskColor := colorGreen
			switch risk {
			case "medium":
				riskColor = colorYellow
			case "high":
				riskColor = colorRed
			}
			fmt.Printf("  %s%-8s%s: %d\n", riskColor, risk, colorReset, count)
		}
	}
	
	return nil
}

func getAuditDBPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "quickcmd", "audit.db")
	}
	return filepath.Join(homeDir, ".quickcmd", "audit.db")
}
