package aliases

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// AliasManager manages command aliases and macros
type AliasManager struct {
	aliases map[string]*Alias
	macros  map[string]*Macro
}

// Alias represents a command alias
type Alias struct {
	ID          string
	Name        string
	Template    string
	Description string
	Variables   []string
	UserID      string
	TeamID      string
	IsPublic    bool
	UsageCount  int
	CreatedAt   time.Time
}

// Macro represents a multi-step command macro
type Macro struct {
	ID          string
	Name        string
	Description string
	Steps       []*MacroStep
	Variables   []string
	UserID      string
	TeamID      string
	CreatedAt   time.Time
}

// MacroStep represents one step in a macro
type MacroStep struct {
	Order       int
	Description string
	Command     string
	ContinueOnError bool
}

// NewAliasManager creates a new alias manager
func NewAliasManager() *AliasManager {
	return &AliasManager{
		aliases: make(map[string]*Alias),
		macros:  make(map[string]*Macro),
	}
}

// CreateAlias creates a new alias
func (am *AliasManager) CreateAlias(name, template, description, userID string) (*Alias, error) {
	if name == "" || template == "" {
		return nil, fmt.Errorf("name and template are required")
	}
	
	// Extract variables from template
	variables := extractVariables(template)
	
	alias := &Alias{
		ID:          generateID(),
		Name:        name,
		Template:    template,
		Description: description,
		Variables:   variables,
		UserID:      userID,
		CreatedAt:   time.Now(),
	}
	
	am.aliases[name] = alias
	return alias, nil
}

// ExecuteAlias executes an alias with given variables
func (am *AliasManager) ExecuteAlias(name string, vars map[string]string) (string, error) {
	alias := am.aliases[name]
	if alias == nil {
		return "", fmt.Errorf("alias not found: %s", name)
	}
	
	// Expand template
	command := alias.Template
	for _, varName := range alias.Variables {
		value, ok := vars[varName]
		if !ok {
			return "", fmt.Errorf("missing variable: %s", varName)
		}
		command = strings.ReplaceAll(command, "{"+varName+"}", value)
	}
	
	// Update usage count
	alias.UsageCount++
	
	return command, nil
}

// CreateMacro creates a new macro
func (am *AliasManager) CreateMacro(name, description string, steps []*MacroStep, userID string) (*Macro, error) {
	if name == "" || len(steps) == 0 {
		return nil, fmt.Errorf("name and steps are required")
	}
	
	// Extract variables from all steps
	variables := []string{}
	for _, step := range steps {
		stepVars := extractVariables(step.Command)
		variables = append(variables, stepVars...)
	}
	variables = unique(variables)
	
	macro := &Macro{
		ID:          generateID(),
		Name:        name,
		Description: description,
		Steps:       steps,
		Variables:   variables,
		UserID:      userID,
		CreatedAt:   time.Now(),
	}
	
	am.macros[name] = macro
	return macro, nil
}

// ExecuteMacro executes a macro with given variables
func (am *AliasManager) ExecuteMacro(name string, vars map[string]string) ([]string, error) {
	macro := am.macros[name]
	if macro == nil {
		return nil, fmt.Errorf("macro not found: %s", name)
	}
	
	commands := []string{}
	for _, step := range macro.Steps {
		// Expand template
		command := step.Command
		for _, varName := range macro.Variables {
			value, ok := vars[varName]
			if !ok {
				return nil, fmt.Errorf("missing variable: %s", varName)
			}
			command = strings.ReplaceAll(command, "{"+varName+"}", value)
		}
		commands = append(commands, command)
	}
	
	return commands, nil
}

// ListAliases returns all aliases for a user
func (am *AliasManager) ListAliases(userID string) []*Alias {
	aliases := []*Alias{}
	for _, alias := range am.aliases {
		if alias.UserID == userID || alias.IsPublic {
			aliases = append(aliases, alias)
		}
	}
	return aliases
}

// ShareAlias shares an alias with a team
func (am *AliasManager) ShareAlias(name, teamID string) error {
	alias := am.aliases[name]
	if alias == nil {
		return fmt.Errorf("alias not found: %s", name)
	}
	
	alias.TeamID = teamID
	return nil
}

// DeleteAlias deletes an alias
func (am *AliasManager) DeleteAlias(name string) error {
	if am.aliases[name] == nil {
		return fmt.Errorf("alias not found: %s", name)
	}
	
	delete(am.aliases, name)
	return nil
}

// Helper functions

func extractVariables(template string) []string {
	// Extract {variable} patterns
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(template, -1)
	
	variables := []string{}
	for _, match := range matches {
		if len(match) > 1 {
			// Handle {name:default} syntax
			varName := strings.Split(match[1], ":")[0]
			variables = append(variables, varName)
		}
	}
	
	return unique(variables)
}

func unique(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// CommunityAliases returns popular community aliases
func CommunityAliases() []*Alias {
	return []*Alias{
		{
			Name:        "k-pods",
			Template:    "kubectl get pods -n {namespace}",
			Description: "List pods in namespace",
			Variables:   []string{"namespace"},
			IsPublic:    true,
		},
		{
			Name:        "k-logs",
			Template:    "kubectl logs -f deployment/{name} -n {namespace}",
			Description: "Follow logs for deployment",
			Variables:   []string{"name", "namespace"},
			IsPublic:    true,
		},
		{
			Name:        "git-undo",
			Template:    "git reset --soft HEAD~{count}",
			Description: "Undo last N commits (keep changes)",
			Variables:   []string{"count"},
			IsPublic:    true,
		},
		{
			Name:        "find-large",
			Template:    "find {path} -type f -size +{size}M -exec ls -lh {} \\;",
			Description: "Find large files",
			Variables:   []string{"path", "size"},
			IsPublic:    true,
		},
	}
}
