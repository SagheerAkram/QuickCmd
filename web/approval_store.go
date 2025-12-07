package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	
	_ "github.com/mattn/go-sqlite3"
)

// ApprovalStatus represents the status of an approval
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

// Approval represents a pending approval for a job
type Approval struct {
	ID               int            `json:"id"`
	RunID            int            `json:"run_id"`
	Prompt           string         `json:"prompt"`
	Command          string         `json:"command"`
	RiskLevel        string         `json:"risk_level"`
	RequiredScopes   []string       `json:"required_scopes"`
	PluginMetadata   map[string]interface{} `json:"plugin_metadata"`
	RequestedBy      string         `json:"requested_by"`
	RequestedAt      time.Time      `json:"requested_at"`
	Status           ApprovalStatus `json:"status"`
	ApprovedBy       string         `json:"approved_by,omitempty"`
	ApprovedAt       *time.Time     `json:"approved_at,omitempty"`
	RejectedBy       string         `json:"rejected_by,omitempty"`
	RejectedAt       *time.Time     `json:"rejected_at,omitempty"`
	RejectionReason  string         `json:"rejection_reason,omitempty"`
	Confirmation     string         `json:"confirmation,omitempty"`
	ApprovalNote     string         `json:"approval_note,omitempty"`
}

// ApprovalStore manages approval records
type ApprovalStore struct {
	db *sql.DB
}

// NewApprovalStore creates a new approval store
func NewApprovalStore(dbPath string) (*ApprovalStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	
	store := &ApprovalStore{db: db}
	
	// Create table if not exists
	if err := store.createTable(); err != nil {
		return nil, err
	}
	
	return store, nil
}

// createTable creates the approvals table
func (s *ApprovalStore) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS approvals (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		prompt TEXT NOT NULL,
		command TEXT NOT NULL,
		risk_level TEXT NOT NULL,
		required_scopes TEXT,
		plugin_metadata TEXT,
		requested_by TEXT NOT NULL,
		requested_at TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		approved_by TEXT,
		approved_at TEXT,
		rejected_by TEXT,
		rejected_at TEXT,
		rejection_reason TEXT,
		confirmation TEXT,
		approval_note TEXT,
		FOREIGN KEY (run_id) REFERENCES runs(id)
	);
	
	CREATE INDEX IF NOT EXISTS idx_approvals_status ON approvals(status);
	CREATE INDEX IF NOT EXISTS idx_approvals_run_id ON approvals(run_id);
	`
	
	_, err := s.db.Exec(query)
	return err
}

// CreateApproval creates a new pending approval
func (s *ApprovalStore) CreateApproval(approval *Approval) (int, error) {
	scopes, _ := json.Marshal(approval.RequiredScopes)
	metadata, _ := json.Marshal(approval.PluginMetadata)
	
	result, err := s.db.Exec(`
		INSERT INTO approvals (
			run_id, prompt, command, risk_level, required_scopes, plugin_metadata,
			requested_by, requested_at, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, approval.RunID, approval.Prompt, approval.Command, approval.RiskLevel,
		string(scopes), string(metadata), approval.RequestedBy,
		approval.RequestedAt.Format(time.RFC3339), ApprovalStatusPending)
	
	if err != nil {
		return 0, err
	}
	
	id, err := result.LastInsertId()
	return int(id), err
}

// GetPendingApprovals retrieves all pending approvals
func (s *ApprovalStore) GetPendingApprovals() ([]*Approval, error) {
	rows, err := s.db.Query(`
		SELECT id, run_id, prompt, command, risk_level, required_scopes, plugin_metadata,
		       requested_by, requested_at, status
		FROM approvals
		WHERE status = ?
		ORDER BY requested_at DESC
	`, ApprovalStatusPending)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var approvals []*Approval
	for rows.Next() {
		approval, err := s.scanApproval(rows)
		if err != nil {
			return nil, err
		}
		approvals = append(approvals, approval)
	}
	
	return approvals, nil
}

// GetApproval retrieves an approval by ID
func (s *ApprovalStore) GetApproval(id int) (*Approval, error) {
	row := s.db.QueryRow(`
		SELECT id, run_id, prompt, command, risk_level, required_scopes, plugin_metadata,
		       requested_by, requested_at, status, approved_by, approved_at,
		       rejected_by, rejected_at, rejection_reason, confirmation, approval_note
		FROM approvals
		WHERE id = ?
	`, id)
	
	return s.scanFullApproval(row)
}

// ApproveApproval approves a pending approval
func (s *ApprovalStore) ApproveApproval(id int, approvedBy, confirmation, note string) error {
	now := time.Now()
	
	result, err := s.db.Exec(`
		UPDATE approvals
		SET status = ?, approved_by = ?, approved_at = ?, confirmation = ?, approval_note = ?
		WHERE id = ? AND status = ?
	`, ApprovalStatusApproved, approvedBy, now.Format(time.RFC3339), confirmation, note, id, ApprovalStatusPending)
	
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return fmt.Errorf("approval not found or already processed")
	}
	
	return nil
}

// RejectApproval rejects a pending approval
func (s *ApprovalStore) RejectApproval(id int, rejectedBy, reason string) error {
	now := time.Now()
	
	result, err := s.db.Exec(`
		UPDATE approvals
		SET status = ?, rejected_by = ?, rejected_at = ?, rejection_reason = ?
		WHERE id = ? AND status = ?
	`, ApprovalStatusRejected, rejectedBy, now.Format(time.RFC3339), reason, id, ApprovalStatusPending)
	
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return fmt.Errorf("approval not found or already processed")
	}
	
	return nil
}

// Helper functions

func (s *ApprovalStore) scanApproval(rows *sql.Rows) (*Approval, error) {
	var approval Approval
	var requestedAt string
	var scopesJSON, metadataJSON string
	
	err := rows.Scan(
		&approval.ID, &approval.RunID, &approval.Prompt, &approval.Command,
		&approval.RiskLevel, &scopesJSON, &metadataJSON,
		&approval.RequestedBy, &requestedAt, &approval.Status,
	)
	
	if err != nil {
		return nil, err
	}
	
	approval.RequestedAt, _ = time.Parse(time.RFC3339, requestedAt)
	json.Unmarshal([]byte(scopesJSON), &approval.RequiredScopes)
	json.Unmarshal([]byte(metadataJSON), &approval.PluginMetadata)
	
	return &approval, nil
}

func (s *ApprovalStore) scanFullApproval(row *sql.Row) (*Approval, error) {
	var approval Approval
	var requestedAt, approvedAt, rejectedAt sql.NullString
	var scopesJSON, metadataJSON string
	var approvedBy, rejectedBy, rejectionReason, confirmation, note sql.NullString
	
	err := row.Scan(
		&approval.ID, &approval.RunID, &approval.Prompt, &approval.Command,
		&approval.RiskLevel, &scopesJSON, &metadataJSON,
		&approval.RequestedBy, &requestedAt, &approval.Status,
		&approvedBy, &approvedAt, &rejectedBy, &rejectedAt,
		&rejectionReason, &confirmation, &note,
	)
	
	if err != nil {
		return nil, err
	}
	
	approval.RequestedAt, _ = time.Parse(time.RFC3339, requestedAt.String)
	json.Unmarshal([]byte(scopesJSON), &approval.RequiredScopes)
	json.Unmarshal([]byte(metadataJSON), &approval.PluginMetadata)
	
	if approvedBy.Valid {
		approval.ApprovedBy = approvedBy.String
		if approvedAt.Valid {
			t, _ := time.Parse(time.RFC3339, approvedAt.String)
			approval.ApprovedAt = &t
		}
		approval.Confirmation = confirmation.String
		approval.ApprovalNote = note.String
	}
	
	if rejectedBy.Valid {
		approval.RejectedBy = rejectedBy.String
		if rejectedAt.Valid {
			t, _ := time.Parse(time.RFC3339, rejectedAt.String)
			approval.RejectedAt = &t
		}
		approval.RejectionReason = rejectionReason.String
	}
	
	return &approval, nil
}

// Close closes the database connection
func (s *ApprovalStore) Close() error {
	return s.db.Close()
}
