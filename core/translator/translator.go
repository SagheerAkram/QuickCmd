package translator

import (
	"errors"
	"sort"
	"strings"
)

var (
	// ErrNoMatch is returned when no templates match the prompt
	ErrNoMatch = errors.New("no matching command templates found")
)

// Translator handles natural language to command translation
type Translator struct {
	templates []*Template
}

// New creates a new Translator with default templates
func New() *Translator {
	return &Translator{
		templates: CommandTemplates,
	}
}

// NewWithTemplates creates a Translator with custom templates
func NewWithTemplates(templates []*Template) *Translator {
	return &Translator{
		templates: templates,
	}
}

// Translate converts a natural language prompt into command candidates
func (t *Translator) Translate(prompt string) ([]*Candidate, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, errors.New("empty prompt")
	}
	
	var candidates []*Candidate
	
	// Try to match against all templates
	for _, template := range t.templates {
		if matches, ok := template.Match(prompt); ok {
			candidate := template.Generator(matches)
			
			// Apply keyword bonus
			keywordBonus := template.CalculateKeywordBonus(prompt)
			candidate.Confidence += keywordBonus
			
			// Cap confidence at 100
			if candidate.Confidence > 100 {
				candidate.Confidence = 100
			}
			
			candidates = append(candidates, candidate)
		}
	}
	
	if len(candidates) == 0 {
		return nil, ErrNoMatch
	}
	
	// Sort candidates by confidence (highest first)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Confidence > candidates[j].Confidence
	})
	
	// Return top 3 candidates
	if len(candidates) > 3 {
		candidates = candidates[:3]
	}
	
	return candidates, nil
}

// AddTemplate adds a custom template to the translator
func (t *Translator) AddTemplate(template *Template) {
	t.templates = append(t.templates, template)
}

// GetTemplatesByCategory returns all templates in a category
func (t *Translator) GetTemplatesByCategory(category string) []*Template {
	var result []*Template
	for _, template := range t.templates {
		if template.Category == category {
			result = append(result, template)
		}
	}
	return result
}

// ListCategories returns all available template categories
func (t *Translator) ListCategories() []string {
	categoryMap := make(map[string]bool)
	for _, template := range t.templates {
		categoryMap[template.Category] = true
	}
	
	categories := make([]string, 0, len(categoryMap))
	for category := range categoryMap {
		categories = append(categories, category)
	}
	
	sort.Strings(categories)
	return categories
}
