package learning

import (
	"fmt"
	"strings"
)

// Explainer provides detailed explanations of commands
type Explainer struct {
	knowledgeBase map[string]*CommandKnowledge
}

// CommandKnowledge represents knowledge about a command
type CommandKnowledge struct {
	Command     string
	Description string
	Flags       map[string]*FlagExplanation
	Examples    []string
	Tips        []string
	Related     []string
}

// FlagExplanation explains a command flag
type FlagExplanation struct {
	Flag        string
	Description string
	Example     string
	Alternatives []string
}

// Explanation represents a detailed command explanation
type Explanation struct {
	Command     string
	Summary     string
	Breakdown   []*BreakdownStep
	Tips        []string
	Optimizations []string
	Related     []string
	Difficulty  string
}

// BreakdownStep represents one step in command breakdown
type BreakdownStep struct {
	Part        string
	Explanation string
	Example     string
}

// NewExplainer creates a new explainer
func NewExplainer() *Explainer {
	e := &Explainer{
		knowledgeBase: make(map[string]*CommandKnowledge),
	}
	e.loadKnowledgeBase()
	return e
}

// loadKnowledgeBase loads command knowledge
func (e *Explainer) loadKnowledgeBase() {
	// Find command
	e.knowledgeBase["find"] = &CommandKnowledge{
		Command:     "find",
		Description: "Search for files and directories based on criteria",
		Flags: map[string]*FlagExplanation{
			"-type": {
				Flag:        "-type",
				Description: "Filter by file type",
				Example:     "-type f (files), -type d (directories)",
			},
			"-name": {
				Flag:        "-name",
				Description: "Match files by name pattern",
				Example:     "-name '*.log' (all .log files)",
			},
			"-size": {
				Flag:        "-size",
				Description: "Match files by size",
				Example:     "+100M (larger than 100MB), -1G (smaller than 1GB)",
			},
			"-mtime": {
				Flag:        "-mtime",
				Description: "Match files by modification time",
				Example:     "-mtime -7 (modified in last 7 days)",
			},
		},
		Examples: []string{
			"find . -name '*.log'",
			"find /var -type f -size +100M",
			"find . -mtime -7",
		},
		Tips: []string{
			"Use -exec ... + instead of \\; for better performance",
			"Combine with -maxdepth to limit search depth",
			"Use -iname for case-insensitive search",
		},
		Related: []string{"grep", "locate", "fd"},
	}

	// Git command
	e.knowledgeBase["git"] = &CommandKnowledge{
		Command:     "git",
		Description: "Version control system",
		Flags: map[string]*FlagExplanation{
			"--force": {
				Flag:        "--force",
				Description: "Force the operation (dangerous!)",
				Example:     "git push --force (overwrites remote)",
			},
		},
		Tips: []string{
			"Always pull before push",
			"Use --force-with-lease instead of --force",
			"Create backup branches before dangerous operations",
		},
	}

	// Kubectl command
	e.knowledgeBase["kubectl"] = &CommandKnowledge{
		Command:     "kubectl",
		Description: "Kubernetes command-line tool",
		Tips: []string{
			"Use -n to specify namespace",
			"Add --dry-run=client to preview changes",
			"Use -o yaml to see full resource definition",
		},
	}
}

// Explain generates a detailed explanation of a command
func (e *Explainer) Explain(command string) (*Explanation, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	baseCmd := parts[0]
	knowledge := e.knowledgeBase[baseCmd]

	explanation := &Explanation{
		Command:    command,
		Breakdown:  []*BreakdownStep{},
		Tips:       []string{},
		Difficulty: "beginner",
	}

	if knowledge != nil {
		explanation.Summary = knowledge.Description
		explanation.Tips = knowledge.Tips
		explanation.Related = knowledge.Related

		// Break down each part
		for i, part := range parts {
			step := &BreakdownStep{
				Part: part,
			}

			if i == 0 {
				step.Explanation = knowledge.Description
			} else if flagExp, ok := knowledge.Flags[part]; ok {
				step.Explanation = flagExp.Description
				step.Example = flagExp.Example
			} else {
				step.Explanation = "Argument or parameter"
			}

			explanation.Breakdown = append(explanation.Breakdown, step)
		}
	} else {
		// Generic explanation
		explanation.Summary = fmt.Sprintf("Executes the '%s' command", baseCmd)
		for _, part := range parts {
			explanation.Breakdown = append(explanation.Breakdown, &BreakdownStep{
				Part:        part,
				Explanation: "Command component",
			})
		}
	}

	// Add optimizations
	explanation.Optimizations = e.suggestOptimizations(command)

	return explanation, nil
}

// suggestOptimizations suggests command optimizations
func (e *Explainer) suggestOptimizations(command string) []string {
	optimizations := []string{}

	// Find + grep → grep -r
	if strings.Contains(command, "find") && strings.Contains(command, "grep") {
		optimizations = append(optimizations,
			"Use 'grep -r' instead of 'find | grep' for faster recursive search")
	}

	// find -exec \; → find -exec +
	if strings.Contains(command, "-exec") && strings.Contains(command, "\\;") {
		optimizations = append(optimizations,
			"Use '-exec ... +' instead of '\\;' to run command once with all files")
	}

	return optimizations
}

// Format formats an explanation for display
func (exp *Explanation) Format() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("📚 Command: %s\n\n", exp.Command))
	sb.WriteString(fmt.Sprintf("%s\n\n", exp.Summary))

	if len(exp.Breakdown) > 0 {
		sb.WriteString("Breakdown:\n")
		for _, step := range exp.Breakdown {
			sb.WriteString(fmt.Sprintf("  %-20s %s\n", step.Part, step.Explanation))
			if step.Example != "" {
				sb.WriteString(fmt.Sprintf("                       Example: %s\n", step.Example))
			}
		}
		sb.WriteString("\n")
	}

	if len(exp.Tips) > 0 {
		sb.WriteString("💡 Tips:\n")
		for _, tip := range exp.Tips {
			sb.WriteString(fmt.Sprintf("  • %s\n", tip))
		}
		sb.WriteString("\n")
	}

	if len(exp.Optimizations) > 0 {
		sb.WriteString("⚡ Optimizations:\n")
		for _, opt := range exp.Optimizations {
			sb.WriteString(fmt.Sprintf("  • %s\n", opt))
		}
		sb.WriteString("\n")
	}

	if len(exp.Related) > 0 {
		sb.WriteString(fmt.Sprintf("🔗 Related: %s\n", strings.Join(exp.Related, ", ")))
	}

	return sb.String()
}
