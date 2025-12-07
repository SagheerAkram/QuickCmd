package main

import (
	"fmt"
	"os"
	
	"github.com/spf13/cobra"
)

var (
	// Version information (set via ldflags during build)
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "quickcmd [prompt]",
	Short: "QuickCMD - AI-first CLI assistant for safe command execution",
	Long: `QuickCMD converts plain-English tasks into safe, auditable shell commands.
	
It provides AI-assisted command generation with full control, safety guardrails,
and complete traceability. Commands are validated against security policies and
can be executed in isolated sandboxes.`,
	Example: `  # Get command suggestions
  quickcmd "find files larger than 100MB"
  
  # Execute in sandbox
  quickcmd "delete all .DS_Store files" --sandbox
  
  # View history
  quickcmd history`,
	Version: Version,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("QuickCMD %s\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Built: %s\n", BuildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(historyCmd)
	
	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default: $HOME/.quickcmd/config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
