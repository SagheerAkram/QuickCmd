package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
	
	"github.com/yourusername/quickcmd/core/executor"
)

func TestSQLiteStore_LogExecution(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_audit.db")
	
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()
	
	// Create test record
	record := &RunRecord{
		Timestamp:       time.Now().Format(time.RFC3339),
		User:            "testuser",
		Prompt:          "find large files",
		SelectedCommand: "find . -type f -size +100M",
		SandboxID:       "abc123",
		ExitCode:        0,
		Stdout:          []byte("./large_file.bin"),
		Stderr:          []byte(""),
		RiskLevel:       "safe",
		Executed:        true,
		DurationMs:      1234,
	}
	
	// Log execution
	if err := store.LogExecution(record); err != nil {
		t.Fatalf("LogExecution() error: %v", err)
	}
	
	if record.ID == 0 {
		t.Error("LogExecution() did not set record ID")
	}
	
	// Retrieve record
	retrieved, err := store.GetRecordByID(record.ID)
	if err != nil {
		t.Fatalf("GetRecordByID() error: %v", err)
	}
	
	if retrieved.Prompt != record.Prompt {
		t.Errorf("Retrieved prompt = %q, want %q", retrieved.Prompt, record.Prompt)
	}
	
	if retrieved.ExitCode != record.ExitCode {
		t.Errorf("Retrieved exit code = %d, want %d", retrieved.ExitCode, record.ExitCode)
	}
}

func TestSQLiteStore_SecretsRedaction(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_audit.db")
	
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()
	
	// Record with secrets
	record := &RunRecord{
		Timestamp:       time.Now().Format(time.RFC3339),
		User:            "testuser",
		Prompt:          "deploy app",
		SelectedCommand: "PASSWORD=secret123 ./deploy.sh",
		Stdout:          []byte("api_key=abc123xyz"),
		Stderr:          []byte(""),
		RiskLevel:       "medium",
		Executed:        true,
	}
	
	if err := store.LogExecution(record); err != nil {
		t.Fatalf("LogExecution() error: %v", err)
	}
	
	// Retrieve and check redaction
	retrieved, err := store.GetRecordByID(record.ID)
	if err != nil {
		t.Fatalf("GetRecordByID() error: %v", err)
	}
	
	// Check command redaction
	if !containsRedacted(retrieved.SelectedCommand) {
		t.Errorf("Command not redacted: %s", retrieved.SelectedCommand)
	}
	
	// Check stdout redaction
	if !containsRedacted(string(retrieved.Stdout)) {
		t.Errorf("Stdout not redacted: %s", retrieved.Stdout)
	}
}

func TestSQLiteStore_GetHistory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_audit.db")
	
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()
	
	// Insert multiple records
	for i := 0; i < 5; i++ {
		record := &RunRecord{
			Timestamp:       time.Now().Format(time.RFC3339),
			User:            "testuser",
			Prompt:          "test prompt",
			SelectedCommand: "echo test",
			RiskLevel:       "safe",
			Executed:        true,
		}
		if err := store.LogExecution(record); err != nil {
			t.Fatalf("LogExecution() error: %v", err)
		}
	}
	
	// Get history
	history, err := store.GetHistory(10, "")
	if err != nil {
		t.Fatalf("GetHistory() error: %v", err)
	}
	
	if len(history) != 5 {
		t.Errorf("GetHistory() returned %d records, want 5", len(history))
	}
	
	// Test with limit
	history, err = store.GetHistory(3, "")
	if err != nil {
		t.Fatalf("GetHistory() error: %v", err)
	}
	
	if len(history) != 3 {
		t.Errorf("GetHistory() with limit returned %d records, want 3", len(history))
	}
}

func TestSQLiteStore_GetHistoryWithFilter(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_audit.db")
	
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()
	
	// Insert records with different prompts
	prompts := []string{"find files", "delete logs", "git commit"}
	for _, prompt := range prompts {
		record := &RunRecord{
			Timestamp:       time.Now().Format(time.RFC3339),
			User:            "testuser",
			Prompt:          prompt,
			SelectedCommand: "echo " + prompt,
			RiskLevel:       "safe",
			Executed:        true,
		}
		if err := store.LogExecution(record); err != nil {
			t.Fatalf("LogExecution() error: %v", err)
		}
	}
	
	// Filter for "find"
	history, err := store.GetHistory(10, "find")
	if err != nil {
		t.Fatalf("GetHistory() error: %v", err)
	}
	
	if len(history) != 1 {
		t.Errorf("GetHistory() with filter returned %d records, want 1", len(history))
	}
	
	if history[0].Prompt != "find files" {
		t.Errorf("GetHistory() returned wrong record: %s", history[0].Prompt)
	}
}

func TestSQLiteStore_GetStats(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_audit.db")
	
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()
	
	// Insert records with different risk levels
	riskLevels := []string{"safe", "safe", "medium", "high"}
	exitCodes := []int{0, 0, 1, 0}
	
	for i, risk := range riskLevels {
		record := &RunRecord{
			Timestamp:       time.Now().Format(time.RFC3339),
			User:            "testuser",
			Prompt:          "test",
			SelectedCommand: "echo test",
			RiskLevel:       risk,
			ExitCode:        exitCodes[i],
			Executed:        true,
		}
		if err := store.LogExecution(record); err != nil {
			t.Fatalf("LogExecution() error: %v", err)
		}
	}
	
	// Get stats
	stats, err := store.GetStats()
	if err != nil {
		t.Fatalf("GetStats() error: %v", err)
	}
	
	total, ok := stats["total_executions"].(int64)
	if !ok || total != 4 {
		t.Errorf("GetStats() total_executions = %v, want 4", stats["total_executions"])
	}
	
	successRate, ok := stats["success_rate"].(float64)
	if !ok || successRate != 75.0 {
		t.Errorf("GetStats() success_rate = %v, want 75.0", stats["success_rate"])
	}
}

func TestEncodeDecodeSnapshot(t *testing.T) {
	snapshot := &executor.SnapshotMetadata{
		Type:       "git",
		Location:   "quickcmd/backup/20250107-093000",
		Timestamp:  time.Now(),
		Reversible: true,
		RestoreCmd: "git checkout main",
	}
	
	// Encode
	encoded := EncodeSnapshot(snapshot)
	if encoded == "" {
		t.Error("EncodeSnapshot() returned empty string")
	}
	
	// Decode
	decoded, err := DecodeSnapshot(encoded)
	if err != nil {
		t.Fatalf("DecodeSnapshot() error: %v", err)
	}
	
	if decoded.Type != snapshot.Type {
		t.Errorf("Decoded type = %s, want %s", decoded.Type, snapshot.Type)
	}
	
	if decoded.Location != snapshot.Location {
		t.Errorf("Decoded location = %s, want %s", decoded.Location, snapshot.Location)
	}
}

func containsRedacted(s string) bool {
	return len(s) > 0 && (s == "***REDACTED***" || 
		len(s) >= 13 && s[len(s)-13:] == "***REDACTED***" ||
		containsSubstring(s, "***REDACTED***"))
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr ||
		(len(s) > len(substr) && indexOfSubstring(s, substr) >= 0))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
