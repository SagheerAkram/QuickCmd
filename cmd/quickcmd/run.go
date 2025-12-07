package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/spf13/cobra"
	"github.com/yourusername/quickcmd/core/audit"
	"github.com/yourusername/quickcmd/core/executor"
	"github.com/yourusername/quickcmd/core/policy"
	"github.com/yourusername/quickcmd/core/translator"
)

var (
	dryRun  bool
	sandbox bool
	yes     bool
)

var runCmd = &cobra.Command{
	Use:   "run [prompt]",
	Short: "Translate and optionally execute a command",
	Long: `Translates a natural language prompt into shell commands.
	
By default, commands are shown but not executed (dry-run mode).
Use --sandbox to execute in an isolated container, or --yes to execute directly.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCommand,
}

func init() {
	runCmd.Flags().BoolVar(&dryRun, "dry-run", true, "show commands without executing")
	runCmd.Flags().BoolVar(&sandbox, "sandbox", false, "execute in isolated sandbox")
	runCmd.Flags().BoolVar(&yes, "yes", false, "execute without confirmation (dangerous!)")
	
	// Make run the default command
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return runCommand(cmd, args)
	}
}

func runCommand(cmd *cobra.Command, args []string) error {
	prompt := strings.Join(args, " ")
	
	// Initialize translator and policy engine
	trans := translator.New()
	policyEngine := policy.NewEngine()
	
	// Translate prompt to candidates
	candidates, err := trans.Translate(prompt)
	if err != nil {
		if err == translator.ErrNoMatch {
			return fmt.Errorf("no matching commands found for: %q\n\nTry being more specific or use different keywords", prompt)
		}
		return fmt.Errorf("translation error: %w", err)
	}
	
	// Display candidates
	fmt.Printf("\n%s Candidates for: %s%s\n\n", colorBold, prompt, colorReset)
	
	for i, candidate := range candidates {
		displayCandidate(i+1, candidate)
		fmt.Println()
	}
	
	// Interactive selection
	if dryRun && !sandbox && !yes {
		fmt.Println(colorYellow + "ℹ️  Dry-run mode: commands will not be executed" + colorReset)
		fmt.Println("Use --sandbox to run in isolated container, or --yes to execute directly")
		return nil
	}
	
	// Select candidate
	var selectedIdx int
	if len(candidates) == 1 {
		selectedIdx = 0
	} else {
		selectedIdx = promptSelection(len(candidates))
		if selectedIdx < 0 {
			fmt.Println("Cancelled.")
			return nil
		}
	}
	
	selected := candidates[selectedIdx]
	
	// Validate against policy
	result := policyEngine.Validate(selected.Command, string(selected.RiskLevel), selected.Destructive)
	
	if !result.Allowed {
		return fmt.Errorf("❌ Command blocked by policy: %s", result.Reason)
	}
	
	// Check if confirmation required
	if result.RequiresConfirm && !yes {
		if !promptConfirmation(result.ConfirmMessage) {
			fmt.Println("Cancelled.")
			return nil
		}
	}
	
	// Execute command
	if sandbox {
		if err := executeInSandbox(selected, policyEngine); err != nil {
			return err
		}
	} else if yes {
		if err := executeDirect(selected, policyEngine); err != nil {
			return err
		}
	} else {
		// Dry-run mode - just show what would be executed
		fmt.Println(colorGreen + "✓ Command validated and ready to execute" + colorReset)
		fmt.Println("\nTo execute:")
		fmt.Println("  --sandbox : Run in isolated Docker container (recommended)")
		fmt.Println("  --yes     : Run directly on host (dangerous!)")
	}
	
	return nil
}

// executeInSandbox executes a command in a Docker sandbox
func executeInSandbox(candidate *translator.Candidate, policyEngine *policy.Engine) error {
	fmt.Println(colorCyan + "🐳 Preparing sandbox environment..." + colorReset)
	
	// Check if Docker is available
	if !executor.IsDockerAvailable() {
		fmt.Println(colorRed + "❌ Docker is not available" + colorReset)
		fmt.Println("\nDocker is required for sandbox execution.")
		fmt.Println("Please install Docker: https://docs.docker.com/get-docker/")
		fmt.Println("\nFalling back to dry-run mode.")
		return nil
	}
	
	// Get Docker info
	if info, err := executor.GetDockerInfo(); err == nil {
		fmt.Printf("Using: %s\n", info)
	}
	
	// Create snapshotter
	snapshotter := executor.NewSnapshotter()
	
	// Create snapshot if destructive
	var snapshot *executor.SnapshotMetadata
	if candidate.Destructive {
		fmt.Println(colorYellow + "📸 Creating pre-run snapshot..." + colorReset)
		
		workingDir, _ := os.Getwd()
		snap, err := snapshotter.CreateSnapshot(workingDir, candidate.AffectedPaths)
		if err != nil {
			fmt.Printf(colorYellow+"⚠️  Snapshot creation failed: %v\n"+colorReset, err)
		} else {
			snapshot = snap
			if snap.Reversible {
				fmt.Printf(colorGreen+"✓ Snapshot created: %s\n"+colorReset, snap.Location)
			}
		}
	}
	
	// Create Docker runner
	runner, err := executor.NewDockerRunner()
	if err != nil {
		return fmt.Errorf("failed to create Docker runner: %w", err)
	}
	defer runner.Close()
	
	// Get working directory
	workingDir, _ := os.Getwd()
	
	// Configure sandbox options
	opts := executor.SandboxOptions{
		Image:         "alpine:latest",
		CPULimit:      0.5,
		MemoryLimit:   256 * 1024 * 1024,
		PidsLimit:     64,
		NetworkAccess: false,
		ReadOnly:      false,
		Timeout:       5 * time.Minute,
		Mounts: []executor.Mount{
			{
				Source:   workingDir,
				Target:   "/workspace",
				ReadOnly: false,
			},
		},
	}
	
	fmt.Println(colorCyan + "🚀 Executing in sandbox..." + colorReset)
	startTime := time.Now()
	
	// Execute in sandbox
	result, err := runner.RunInSandbox(candidate.Command, opts)
	duration := time.Since(startTime)
	
	// Log to audit database
	auditStore, auditErr := audit.NewSQLiteStore(getAuditDBPath())
	if auditErr == nil {
		defer auditStore.Close()
		
		record := &audit.RunRecord{
			Timestamp:       time.Now().Format(time.RFC3339),
			Prompt:          "", // TODO: Pass prompt from caller
			SelectedCommand: candidate.Command,
			SandboxID:       result.SandboxID,
			ExitCode:        result.ExitCode,
			Stdout:          result.Stdout,
			Stderr:          result.Stderr,
			RiskLevel:       string(candidate.RiskLevel),
			Snapshot:        audit.EncodeSnapshot(snapshot),
			Executed:        true,
			DurationMs:      duration.Milliseconds(),
		}
		
		if logErr := auditStore.LogExecution(record); logErr != nil {
			fmt.Printf(colorYellow+"⚠️  Failed to log execution: %v\n"+colorReset, logErr)
		}
	}
	
	// Display results
	fmt.Printf("\n%s Execution completed in %v%s\n", colorBold, duration.Round(time.Millisecond), colorReset)
	fmt.Printf("Sandbox ID: %s\n", result.SandboxID)
	fmt.Printf("Exit Code: %d\n", result.ExitCode)
	
	if len(result.Stdout) > 0 {
		fmt.Printf("\n%sOutput:%s\n%s\n", colorBold, colorReset, string(result.Stdout))
	}
	
	if len(result.Stderr) > 0 {
		fmt.Printf("\n%sErrors:%s\n%s\n", colorRed, colorReset, string(result.Stderr))
	}
	
	if result.ExitCode == 0 {
		fmt.Println(colorGreen + "✓ Command executed successfully" + colorReset)
	} else {
		fmt.Printf(colorRed+"❌ Command failed with exit code %d\n"+colorReset, result.ExitCode)
	}
	
	// Show undo option if snapshot was created
	if snapshot != nil && snapshot.Reversible {
		fmt.Printf("\n%sUndo available:%s %s\n", colorYellow, colorReset, snapshot.RestoreCmd)
	}
	
	return nil
}

// executeDirect executes a command directly on the host (dangerous!)
func executeDirect(candidate *translator.Candidate, policyEngine *policy.Engine) error {
	fmt.Println(colorRed + "⚠️  EXECUTING DIRECTLY ON HOST" + colorReset)
	fmt.Println(colorRed + "This bypasses sandbox isolation!" + colorReset)
	
	// TODO: Implement direct execution
	fmt.Println(colorYellow + "Direct execution not yet implemented" + colorReset)
	fmt.Printf("Would execute: %s\n", candidate.Command)
	
	return nil
}

// getAuditDBPath returns the path to the audit database
func getAuditDBPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "quickcmd", "audit.db")
	}
	return filepath.Join(homeDir, ".quickcmd", "audit.db")
}

func displayCandidate(num int, c *translator.Candidate) {
	// Header with number and risk
	fmt.Printf("%s%d. %s %s%s\n", colorBold, num, c.RiskIcon(), c.RiskColor(), string(c.RiskLevel))
	fmt.Print(colorReset)
	
	// Command (copyable)
	fmt.Printf("   %s%s%s\n", colorCyan, c.Command, colorReset)
	
	// Confidence
	confidenceBar := makeProgressBar(c.Confidence, 20)
	fmt.Printf("   Confidence: %s %d%%\n", confidenceBar, c.Confidence)
	
	// Explanation
	fmt.Printf("   %s\n", c.Explanation)
	
	// Warnings
	if c.Destructive {
		fmt.Printf("   %s⚠️  DESTRUCTIVE OPERATION%s\n", colorRed, colorReset)
	}
	
	if c.RequiresConfirm {
		fmt.Printf("   %s🔒 Requires confirmation%s\n", colorYellow, colorReset)
	}
	
	// Breakdown
	if len(c.Breakdown) > 0 {
		fmt.Println("   Breakdown:")
		for i, step := range c.Breakdown {
			fmt.Printf("     %d. %s\n", i+1, step.Description)
		}
	}
	
	// Affected paths
	if len(c.AffectedPaths) > 0 {
		fmt.Printf("   Affected: %s\n", strings.Join(c.AffectedPaths, ", "))
	}
}

func promptSelection(maxNum int) int {
	reader := bufio.NewReader(os.Stdin)
	
	for {
		fmt.Printf("\nSelect candidate [1-%d] or 'q' to quit: ", maxNum)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		
		if input == "q" || input == "quit" {
			return -1
		}
		
		var num int
		if _, err := fmt.Sscanf(input, "%d", &num); err == nil {
			if num >= 1 && num <= maxNum {
				return num - 1
			}
		}
		
		fmt.Println(colorRed + "Invalid selection. Try again." + colorReset)
	}
}

func promptConfirmation(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Printf("\n%s%s%s\n", colorYellow, message, colorReset)
	fmt.Print("Confirmation: ")
	
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	// Check for exact match of expected confirmations
	expectedConfirmations := []string{
		"I UNDERSTAND",
		"CONFIRM DELETE",
		"CONFIRM",
	}
	
	for _, expected := range expectedConfirmations {
		if input == expected {
			return true
		}
	}
	
	return false
}

func makeProgressBar(value, width int) string {
	filled := (value * width) / 100
	empty := width - filled
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	
	// Color based on value
	if value >= 80 {
		return colorGreen + bar + colorReset
	} else if value >= 50 {
		return colorYellow + bar + colorReset
	}
	return colorRed + bar + colorReset
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)
