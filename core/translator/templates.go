package translator

import (
	"fmt"
	"regexp"
	"strings"
)

// Template represents a command template with pattern matching
type Template struct {
	Patterns    []*regexp.Regexp // Regex patterns to match user input
	Generator   func(matches []string) *Candidate
	Keywords    []string // Keywords that boost confidence
	Category    string   // Category (file, git, docker, etc.)
	Description string   // Template description
}

// CommandTemplate is the global registry of command templates
var CommandTemplates = []*Template{
	// File operations - Find
	{
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)find.*files?\s+(?:larger|bigger)\s+than\s+(\d+)\s*(mb|gb|kb)?`),
			regexp.MustCompile(`(?i)(?:list|show).*files?\s+(?:over|above)\s+(\d+)\s*(mb|gb|kb)?`),
		},
		Keywords:    []string{"find", "large", "files", "size"},
		Category:    "file",
		Description: "Find files larger than specified size",
		Generator: func(matches []string) *Candidate {
			size := matches[1]
			unit := "M"
			if len(matches) > 2 && matches[2] != "" {
				switch strings.ToLower(matches[2]) {
				case "kb":
					unit = "k"
				case "gb":
					unit = "G"
				}
			}
			
			cmd := fmt.Sprintf("find . -type f -size +%s%s", size, unit)
			
			return &Candidate{
				Command:     cmd,
				Explanation: fmt.Sprintf("Finds all files in the current directory and subdirectories larger than %s%s", size, matches[2]),
				Breakdown: []Step{
					{Description: "Search current directory recursively", Command: "find ."},
					{Description: "Filter for regular files only", Command: "-type f"},
					{Description: fmt.Sprintf("Match files larger than %s%s", size, matches[2]), Command: fmt.Sprintf("-size +%s%s", size, unit)},
				},
				Confidence:  95,
				RiskLevel:   RiskSafe,
				Destructive: false,
				DocLinks:    []string{"https://man7.org/linux/man-pages/man1/find.1.html"},
			}
		},
	},
	
	// File operations - Find and delete
	{
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:delete|remove|rm)\s+(?:all\s+)?\.?(\w+)\s+files?`),
			regexp.MustCompile(`(?i)(?:find|search)\s+and\s+(?:delete|remove)\s+\.?(\w+)`),
		},
		Keywords:    []string{"delete", "remove", "files"},
		Category:    "file",
		Description: "Find and delete files by pattern",
		Generator: func(matches []string) *Candidate {
			pattern := matches[1]
			if !strings.HasPrefix(pattern, ".") && !strings.Contains(pattern, "*") {
				pattern = "*." + pattern
			}
			
			cmd := fmt.Sprintf("find . -name \"%s\" -type f -delete", pattern)
			
			return &Candidate{
				Command:     cmd,
				Explanation: fmt.Sprintf("Finds and deletes all files matching pattern '%s' in current directory and subdirectories", pattern),
				Breakdown: []Step{
					{Description: "Search current directory recursively", Command: "find ."},
					{Description: fmt.Sprintf("Match files named '%s'", pattern), Command: fmt.Sprintf("-name \"%s\"", pattern)},
					{Description: "Filter for regular files only", Command: "-type f"},
					{Description: "Delete matched files", Command: "-delete"},
				},
				Confidence:      90,
				RiskLevel:       RiskHigh,
				Destructive:     true,
				RequiresConfirm: true,
				DocLinks:        []string{"https://man7.org/linux/man-pages/man1/find.1.html"},
			}
		},
	},
	
	// File operations - Archive/compress
	{
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:archive|compress|tar|zip)\s+(?:logs?|files?)\s+older\s+than\s+(\d+)\s+days?`),
			regexp.MustCompile(`(?i)(?:create|make)\s+(?:archive|backup)\s+of\s+(?:logs?|files?)\s+(?:from\s+)?(\d+)\s+days?`),
		},
		Keywords:    []string{"archive", "compress", "tar", "old", "logs"},
		Category:    "file",
		Description: "Archive old log files",
		Generator: func(matches []string) *Candidate {
			days := matches[1]
			
			cmd := fmt.Sprintf("find . -name \"*.log\" -type f -mtime +%s -print0 | tar -czvf logs_archive_$(date +%%Y%%m%%d).tar.gz --null -T -", days)
			
			return &Candidate{
				Command:     cmd,
				Explanation: fmt.Sprintf("Finds log files older than %s days and creates a compressed archive", days),
				Breakdown: []Step{
					{Description: "Find .log files", Command: "find . -name \"*.log\" -type f"},
					{Description: fmt.Sprintf("Filter files modified more than %s days ago", days), Command: fmt.Sprintf("-mtime +%s", days)},
					{Description: "Output null-separated paths", Command: "-print0"},
					{Description: "Create compressed tar archive", Command: "tar -czvf logs_archive_$(date +%Y%m%d).tar.gz"},
				},
				Confidence:  88,
				RiskLevel:   RiskMedium,
				Destructive: false,
				DocLinks:    []string{"https://man7.org/linux/man-pages/man1/tar.1.html", "https://man7.org/linux/man-pages/man1/find.1.html"},
			}
		},
	},
	
	// Git operations - Status/diff
	{
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:show|display|list)\s+(?:git\s+)?(?:changes|modifications|diffs?)`),
			regexp.MustCompile(`(?i)(?:what|which)\s+files?\s+(?:changed|modified)`),
		},
		Keywords:    []string{"git", "changes", "status", "diff"},
		Category:    "git",
		Description: "Show git changes",
		Generator: func(matches []string) *Candidate {
			return &Candidate{
				Command:     "git status --short",
				Explanation: "Shows a concise summary of changed files in the git repository",
				Breakdown: []Step{
					{Description: "Display git working tree status", Command: "git status"},
					{Description: "Use short format for cleaner output", Command: "--short"},
				},
				Confidence:  98,
				RiskLevel:   RiskSafe,
				Destructive: false,
				DocLinks:    []string{"https://git-scm.com/docs/git-status"},
			}
		},
	},
	
	// Git operations - Commit
	{
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:commit|save)\s+(?:all\s+)?changes?\s+(?:with\s+message\s+)?["\']?(.+?)["\']?$`),
			regexp.MustCompile(`(?i)git\s+commit.*["\'](.+?)["\']`),
		},
		Keywords:    []string{"commit", "git", "save", "changes"},
		Category:    "git",
		Description: "Commit changes to git",
		Generator: func(matches []string) *Candidate {
			message := "Update files"
			if len(matches) > 1 && matches[1] != "" {
				message = matches[1]
			}
			
			return &Candidate{
				Command:     fmt.Sprintf("git add -A && git commit -m \"%s\"", message),
				Explanation: fmt.Sprintf("Stages all changes and commits them with message: '%s'", message),
				Breakdown: []Step{
					{Description: "Stage all changes (new, modified, deleted)", Command: "git add -A"},
					{Description: fmt.Sprintf("Commit with message '%s'", message), Command: fmt.Sprintf("git commit -m \"%s\"", message)},
				},
				Confidence:  92,
				RiskLevel:   RiskMedium,
				Destructive: false,
				DocLinks:    []string{"https://git-scm.com/docs/git-commit"},
			}
		},
	},
	
	// Docker operations - Cleanup
	{
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:clean|cleanup|remove|delete)\s+(?:unused\s+)?docker\s+(?:containers?|images?)`),
			regexp.MustCompile(`(?i)docker\s+(?:prune|cleanup|clean)`),
		},
		Keywords:    []string{"docker", "cleanup", "prune", "remove"},
		Category:    "docker",
		Description: "Clean up Docker resources",
		Generator: func(matches []string) *Candidate {
			return &Candidate{
				Command:     "docker system prune -a --volumes",
				Explanation: "Removes all unused Docker containers, images, networks, and volumes to free up disk space",
				Breakdown: []Step{
					{Description: "Run Docker system cleanup", Command: "docker system prune"},
					{Description: "Remove all unused images, not just dangling ones", Command: "-a"},
					{Description: "Also remove unused volumes", Command: "--volumes"},
				},
				Confidence:      85,
				RiskLevel:       RiskHigh,
				Destructive:     true,
				RequiresConfirm: true,
				DocLinks:        []string{"https://docs.docker.com/engine/reference/commandline/system_prune/"},
			}
		},
	},
	
	// System operations - Disk usage
	{
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:show|display|check)\s+(?:disk\s+)?(?:usage|space)`),
			regexp.MustCompile(`(?i)(?:how\s+much|what)\s+(?:disk\s+)?space`),
		},
		Keywords:    []string{"disk", "usage", "space", "du"},
		Category:    "system",
		Description: "Show disk usage",
		Generator: func(matches []string) *Candidate {
			return &Candidate{
				Command:     "du -sh * | sort -hr | head -20",
				Explanation: "Shows disk usage of directories and files, sorted by size (largest first), limited to top 20",
				Breakdown: []Step{
					{Description: "Calculate disk usage for each item", Command: "du -sh *"},
					{Description: "Sort by human-readable size (largest first)", Command: "sort -hr"},
					{Description: "Show only top 20 results", Command: "head -20"},
				},
				Confidence:  95,
				RiskLevel:   RiskSafe,
				Destructive: false,
				DocLinks:    []string{"https://man7.org/linux/man-pages/man1/du.1.html"},
			}
		},
	},
	
	// Search operations - Find TODO comments
	{
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:find|search|list)\s+(?:all\s+)?todo\s+(?:comments?|items?)`),
			regexp.MustCompile(`(?i)(?:show|display)\s+todos?`),
		},
		Keywords:    []string{"todo", "find", "search", "comments"},
		Category:    "search",
		Description: "Find TODO comments in code",
		Generator: func(matches []string) *Candidate {
			return &Candidate{
				Command:     "grep -rn \"TODO\" --include=\"*.go\" --include=\"*.py\" --include=\"*.js\" --include=\"*.java\" .",
				Explanation: "Searches recursively for TODO comments in common source code files",
				Breakdown: []Step{
					{Description: "Search recursively", Command: "grep -r"},
					{Description: "Show line numbers", Command: "-n"},
					{Description: "Search for 'TODO'", Command: "\"TODO\""},
					{Description: "Include common source file types", Command: "--include=\"*.go\" --include=\"*.py\" ..."},
				},
				Confidence:  93,
				RiskLevel:   RiskSafe,
				Destructive: false,
				DocLinks:    []string{"https://man7.org/linux/man-pages/man1/grep.1.html"},
			}
		},
	},
}

// MatchTemplate attempts to match a prompt against a template
func (t *Template) Match(prompt string) ([]string, bool) {
	for _, pattern := range t.Patterns {
		if matches := pattern.FindStringSubmatch(prompt); matches != nil {
			return matches, true
		}
	}
	return nil, false
}

// CalculateKeywordBonus calculates confidence bonus based on keyword matches
func (t *Template) CalculateKeywordBonus(prompt string) int {
	promptLower := strings.ToLower(prompt)
	matchCount := 0
	
	for _, keyword := range t.Keywords {
		if strings.Contains(promptLower, keyword) {
			matchCount++
		}
	}
	
	// Each keyword match adds 1-2% confidence, max 10%
	bonus := matchCount * 2
	if bonus > 10 {
		bonus = 10
	}
	
	return bonus
}
