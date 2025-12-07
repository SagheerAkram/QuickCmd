package agent

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// JobPayload represents a signed job request from the controller
type JobPayload struct {
	JobID          string                 `json:"job_id"`
	Prompt         string                 `json:"prompt"`
	Command        string                 `json:"command"`
	CandidateMetadata map[string]interface{} `json:"candidate_metadata"`
	PluginMetadata map[string]interface{} `json:"plugin_metadata"`
	RequiredScopes []string               `json:"required_scopes"`
	SnapshotMetadata string               `json:"snapshot_metadata"` // JSON-encoded
	TTL            int64                  `json:"ttl"` // Unix timestamp
	Timestamp      int64                  `json:"timestamp"` // Unix timestamp
	ControllerID   string                 `json:"controller_id"`
}

// JobSignature contains the HMAC signature for a job payload
type JobSignature struct {
	Signature string `json:"signature"`
	Algorithm string `json:"algorithm"` // "HMAC-SHA256"
}

// SignedJob combines payload and signature
type SignedJob struct {
	Payload   JobPayload   `json:"payload"`
	Signature JobSignature `json:"signature"`
}

// JobStatus represents the current status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusRejected  JobStatus = "rejected"
)

// JobResult contains the execution result
type JobResult struct {
	JobID      string    `json:"job_id"`
	Status     JobStatus `json:"status"`
	SandboxID  string    `json:"sandbox_id"`
	ExitCode   int       `json:"exit_code"`
	Stdout     string    `json:"stdout"`
	Stderr     string    `json:"stderr"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	DurationMs int64     `json:"duration_ms"`
	Error      string    `json:"error,omitempty"`
	Snapshot   string    `json:"snapshot,omitempty"` // JSON-encoded snapshot metadata
}

// LogFrame represents a single log message during streaming
type LogFrame struct {
	JobID     string    `json:"job_id"`
	Timestamp time.Time `json:"timestamp"`
	Stream    string    `json:"stream"` // "stdout" or "stderr"
	Data      string    `json:"data"`
	Final     bool      `json:"final"` // True for the last frame
}

// SignPayload creates an HMAC signature for a job payload
func SignPayload(payload *JobPayload, secret string) (string, error) {
	// Serialize payload to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	
	// Create HMAC
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	signature := hex.EncodeToString(h.Sum(nil))
	
	return signature, nil
}

// ValidateSignature verifies the HMAC signature of a job
func ValidateSignature(job *SignedJob, secret string) error {
	// Generate expected signature
	expectedSig, err := SignPayload(&job.Payload, secret)
	if err != nil {
		return err
	}
	
	// Compare signatures (constant-time comparison)
	if !hmac.Equal([]byte(expectedSig), []byte(job.Signature.Signature)) {
		return ErrInvalidSignature
	}
	
	return nil
}

// ValidateTTL checks if the job has expired
func ValidateTTL(payload *JobPayload) error {
	now := time.Now().Unix()
	
	if payload.TTL < now {
		return ErrJobExpired
	}
	
	// Also check if timestamp is too old (prevent replay attacks)
	maxAge := int64(300) // 5 minutes
	if now-payload.Timestamp > maxAge {
		return ErrJobTooOld
	}
	
	return nil
}

// Errors
var (
	ErrInvalidSignature = &AgentError{Code: "INVALID_SIGNATURE", Message: "Invalid job signature"}
	ErrJobExpired       = &AgentError{Code: "JOB_EXPIRED", Message: "Job has expired (TTL exceeded)"}
	ErrJobTooOld        = &AgentError{Code: "JOB_TOO_OLD", Message: "Job timestamp too old (possible replay attack)"}
	ErrInsufficientScopes = &AgentError{Code: "INSUFFICIENT_SCOPES", Message: "Insufficient scopes for job execution"}
)

// AgentError represents an agent-specific error
type AgentError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AgentError) Error() string {
	return e.Code + ": " + e.Message
}
