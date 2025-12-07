package audit

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"
	
	_ "github.com/mattn/go-sqlite3"
	"github.com/yourusername/quickcmd/core/executor"
	"github.com/yourusername/quickcmd/core/policy"
)

//go:embed schema.sql
var schemaSQL string

// RunRecord represents a single command execution record
type RunRecord struct {
	ID              int64
	Timestamp       string
	User            string
	Prompt          string
	SelectedCommand string
	SandboxID       string
	ExitCode        int
	Stdout          []byte
	Stderr          []byte
	RiskLevel       string
	Snapshot        string // JSON-encoded snapshot metadata
	Executed        bool
	DurationMs      int64
	CreatedAt       time.Time
}

// SQLiteStore manages audit log storage
type SQLiteStore struct {
	db       *sql.DB
	redactor *policy.SecretRedactor
}

// NewSQLiteStore creates a new SQLite audit store
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit directory: %w", err)
	}
	
	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit database: %w", err)
	}
	
	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}
	
	store := &SQLiteStore{
		db:       db,
		redactor: policy.NewSecretRedactor(),
	}
	
	// Run migrations
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	
	return store, nil
}

// migrate runs database migrations
func (s *SQLiteStore) migrate() error {
	_, err := s.db.Exec(schemaSQL)
	return err
}

// LogExecution logs a command execution
func (s *SQLiteStore) LogExecution(record *RunRecord) error {
	// Redact secrets from command and output
	record.SelectedCommand = s.redactor.Redact(record.SelectedCommand)
	record.SelectedCommand = s.redactor.RedactEnvVars(record.SelectedCommand)
	
	if record.Stdout != nil {
		redacted := s.redactor.Redact(string(record.Stdout))
		record.Stdout = []byte(redacted)
	}
	
	if record.Stderr != nil {
		redacted := s.redactor.Redact(string(record.Stderr))
		record.Stderr = []byte(redacted)
	}
	
	// Get current user if not set
	if record.User == "" {
		if currentUser, err := user.Current(); err == nil {
			record.User = currentUser.Username
		}
	}
	
	// Set timestamp if not set
	if record.Timestamp == "" {
		record.Timestamp = time.Now().Format(time.RFC3339)
	}
	
	// Insert record
	query := `
		INSERT INTO runs (
			timestamp, user, prompt, selected_command, sandbox_id,
			exit_code, stdout, stderr, risk_level, snapshot,
			executed, duration_ms
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := s.db.Exec(query,
		record.Timestamp,
		record.User,
		record.Prompt,
		record.SelectedCommand,
		record.SandboxID,
		record.ExitCode,
		record.Stdout,
		record.Stderr,
		record.RiskLevel,
		record.Snapshot,
		record.Executed,
		record.DurationMs,
	)
	
	if err != nil {
		return fmt.Errorf("failed to insert audit record: %w", err)
	}
	
	record.ID, _ = result.LastInsertId()
	return nil
}

// GetHistory retrieves execution history
func (s *SQLiteStore) GetHistory(limit int, filter string) ([]*RunRecord, error) {
	query := `
		SELECT id, timestamp, user, prompt, selected_command, sandbox_id,
		       exit_code, stdout, stderr, risk_level, snapshot, executed,
		       duration_ms, created_at
		FROM runs
		WHERE 1=1
	`
	
	args := []interface{}{}
	
	if filter != "" {
		query += " AND (prompt LIKE ? OR selected_command LIKE ?)"
		filterPattern := "%" + filter + "%"
		args = append(args, filterPattern, filterPattern)
	}
	
	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, limit)
	
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()
	
	var records []*RunRecord
	for rows.Next() {
		record := &RunRecord{}
		err := rows.Scan(
			&record.ID,
			&record.Timestamp,
			&record.User,
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
			&record.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}
		records = append(records, record)
	}
	
	return records, nil
}

// GetRecordByID retrieves a specific record
func (s *SQLiteStore) GetRecordByID(id int64) (*RunRecord, error) {
	query := `
		SELECT id, timestamp, user, prompt, selected_command, sandbox_id,
		       exit_code, stdout, stderr, risk_level, snapshot, executed,
		       duration_ms, created_at
		FROM runs
		WHERE id = ?
	`
	
	record := &RunRecord{}
	err := s.db.QueryRow(query, id).Scan(
		&record.ID,
		&record.Timestamp,
		&record.User,
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
		&record.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("record not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get record: %w", err)
	}
	
	return record, nil
}

// GetStats returns audit log statistics
func (s *SQLiteStore) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total executions
	var total int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM runs WHERE executed = 1").Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total_executions"] = total
	
	// By risk level
	rows, err := s.db.Query(`
		SELECT risk_level, COUNT(*) 
		FROM runs 
		WHERE executed = 1 
		GROUP BY risk_level
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	riskCounts := make(map[string]int64)
	for rows.Next() {
		var riskLevel string
		var count int64
		if err := rows.Scan(&riskLevel, &count); err != nil {
			continue
		}
		riskCounts[riskLevel] = count
	}
	stats["by_risk_level"] = riskCounts
	
	// Success rate
	var successful int64
	err = s.db.QueryRow("SELECT COUNT(*) FROM runs WHERE executed = 1 AND exit_code = 0").Scan(&successful)
	if err == nil && total > 0 {
		stats["success_rate"] = float64(successful) / float64(total) * 100
	}
	
	return stats, nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Helper function to encode snapshot metadata as JSON
func EncodeSnapshot(snapshot *executor.SnapshotMetadata) string {
	if snapshot == nil {
		return ""
	}
	
	data, err := json.Marshal(snapshot)
	if err != nil {
		return ""
	}
	
	return string(data)
}

// Helper function to decode snapshot metadata from JSON
func DecodeSnapshot(data string) (*executor.SnapshotMetadata, error) {
	if data == "" {
		return nil, nil
	}
	
	var snapshot executor.SnapshotMetadata
	if err := json.Unmarshal([]byte(data), &snapshot); err != nil {
		return nil, err
	}
	
	return &snapshot, nil
}
