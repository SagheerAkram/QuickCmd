package audit

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SearchQuery represents a search query for command history
type SearchQuery struct {
	Text         string
	RiskLevel    string
	StartDate    time.Time
	EndDate      time.Time
	ExitCode     *int
	UserID       string
	PluginType   string
	MinDuration  time.Duration
	MaxDuration  time.Duration
	Limit        int
	Offset       int
}

// SearchEngine provides advanced search capabilities
type SearchEngine struct {
	db *sql.DB
}

// NewSearchEngine creates a new search engine
func NewSearchEngine(db *sql.DB) *SearchEngine {
	return &SearchEngine{db: db}
}

// Search performs an advanced search on command history
func (se *SearchEngine) Search(query SearchQuery) ([]*RunRecord, error) {
	// Build SQL query
	sql := `SELECT * FROM runs WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	// Full-text search
	if query.Text != "" {
		sql += fmt.Sprintf(` AND (
			prompt LIKE $%d OR 
			selected_command LIKE $%d OR 
			stdout LIKE $%d OR 
			stderr LIKE $%d
		)`, argIdx, argIdx, argIdx, argIdx)
		args = append(args, "%"+query.Text+"%")
		argIdx++
	}

	// Risk level filter
	if query.RiskLevel != "" {
		sql += fmt.Sprintf(` AND risk_level = $%d`, argIdx)
		args = append(args, query.RiskLevel)
		argIdx++
	}

	// Date range filter
	if !query.StartDate.IsZero() {
		sql += fmt.Sprintf(` AND timestamp >= $%d`, argIdx)
		args = append(args, query.StartDate.Format(time.RFC3339))
		argIdx++
	}
	if !query.EndDate.IsZero() {
		sql += fmt.Sprintf(` AND timestamp <= $%d`, argIdx)
		args = append(args, query.EndDate.Format(time.RFC3339))
		argIdx++
	}

	// Exit code filter
	if query.ExitCode != nil {
		sql += fmt.Sprintf(` AND exit_code = $%d`, argIdx)
		args = append(args, *query.ExitCode)
		argIdx++
	}

	// User filter
	if query.UserID != "" {
		sql += fmt.Sprintf(` AND user_id = $%d`, argIdx)
		args = append(args, query.UserID)
		argIdx++
	}

	// Duration filter
	if query.MinDuration > 0 {
		sql += fmt.Sprintf(` AND duration_ms >= $%d`, argIdx)
		args = append(args, query.MinDuration.Milliseconds())
		argIdx++
	}
	if query.MaxDuration > 0 {
		sql += fmt.Sprintf(` AND duration_ms <= $%d`, argIdx)
		args = append(args, query.MaxDuration.Milliseconds())
		argIdx++
	}

	// Order by timestamp descending
	sql += ` ORDER BY timestamp DESC`

	// Limit and offset
	if query.Limit > 0 {
		sql += fmt.Sprintf(` LIMIT $%d`, argIdx)
		args = append(args, query.Limit)
		argIdx++
	}
	if query.Offset > 0 {
		sql += fmt.Sprintf(` OFFSET $%d`, argIdx)
		args = append(args, query.Offset)
	}

	// Execute query
	rows, err := se.db.Query(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer rows.Close()

	// Parse results
	results := []*RunRecord{}
	for rows.Next() {
		record := &RunRecord{}
		err := rows.Scan(
			&record.ID,
			&record.Timestamp,
			&record.Prompt,
			&record.SelectedCommand,
			&record.SandboxID,
			&record.ExitCode,
			&record.Stdout,
			&record.Stderr,
			&record.RiskLevel,
			&record.Snapshot,
			&record.Executed,
			&record.DurationMs,
		)
		if err != nil {
			continue
		}
		results = append(results, record)
	}

	return results, nil
}

// SavedSearch represents a saved search query
type SavedSearch struct {
	ID        int
	Name      string
	Query     SearchQuery
	UserID    string
	CreatedAt time.Time
}

// SaveSearch saves a search query for later use
func (se *SearchEngine) SaveSearch(name, userID string, query SearchQuery) error {
	// TODO: Implement saved searches
	return nil
}

// GetSavedSearches retrieves saved searches for a user
func (se *SearchEngine) GetSavedSearches(userID string) ([]*SavedSearch, error) {
	// TODO: Implement
	return nil, nil
}

// QuickFilters provides common filter presets
func QuickFilters() map[string]SearchQuery {
	return map[string]SearchQuery{
		"today": {
			StartDate: time.Now().Truncate(24 * time.Hour),
		},
		"this-week": {
			StartDate: time.Now().AddDate(0, 0, -7),
		},
		"high-risk": {
			RiskLevel: "high",
		},
		"failed": {
			ExitCode: intPtr(1),
		},
		"slow": {
			MinDuration: 5 * time.Second,
		},
	}
}

func intPtr(i int) *int {
	return &i
}
