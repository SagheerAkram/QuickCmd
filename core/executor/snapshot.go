package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SnapshotMetadata contains information about a pre-run snapshot
type SnapshotMetadata struct {
	Type        string    // "git" or "filesystem"
	Location    string    // Branch name or backup directory
	Timestamp   time.Time
	AffectedPaths []string
	Reversible  bool
	RestoreCmd  string // Command to restore the snapshot
}

// Snapshotter creates pre-run snapshots for undo capability
type Snapshotter struct {
	backupDir string
}

// NewSnapshotter creates a new snapshotter
func NewSnapshotter() *Snapshotter {
	backupDir := filepath.Join(os.TempDir(), "quickcmd", "backups")
	return &Snapshotter{
		backupDir: backupDir,
	}
}

// CreateSnapshot creates a snapshot based on the working directory
func (s *Snapshotter) CreateSnapshot(workingDir string, affectedPaths []string) (*SnapshotMetadata, error) {
	// Check if we're in a Git repository
	if isGitRepo(workingDir) {
		return s.createGitSnapshot(workingDir)
	}
	
	// Fall back to filesystem snapshot
	return s.createFilesystemSnapshot(workingDir, affectedPaths)
}

// createGitSnapshot creates a Git backup branch
func (s *Snapshotter) createGitSnapshot(workingDir string) (*SnapshotMetadata, error) {
	timestamp := time.Now().Format("20060102-150405")
	branchName := fmt.Sprintf("quickcmd/backup/%s", timestamp)
	
	// Create backup branch
	cmd := exec.Command("git", "branch", branchName)
	cmd.Dir = workingDir
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to create backup branch: %w\n%s", err, output)
	}
	
	// Get current branch for restore command
	currentBranch, err := getCurrentGitBranch(workingDir)
	if err != nil {
		currentBranch = "main"
	}
	
	return &SnapshotMetadata{
		Type:       "git",
		Location:   branchName,
		Timestamp:  time.Now(),
		Reversible: true,
		RestoreCmd: fmt.Sprintf("git checkout %s && git branch -D %s", currentBranch, branchName),
	}, nil
}

// createFilesystemSnapshot creates a filesystem backup
func (s *Snapshotter) createFilesystemSnapshot(workingDir string, affectedPaths []string) (*SnapshotMetadata, error) {
	timestamp := time.Now().Format("20060102-150405")
	snapshotID := fmt.Sprintf("snapshot-%s", timestamp)
	backupPath := filepath.Join(s.backupDir, snapshotID)
	
	// Create backup directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Copy affected files
	copiedPaths := []string{}
	for _, path := range affectedPaths {
		fullPath := filepath.Join(workingDir, path)
		
		// Check if file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			continue
		}
		
		// Create destination directory
		destPath := filepath.Join(backupPath, path)
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			continue
		}
		
		// Copy file
		if err := copyFile(fullPath, destPath); err != nil {
			continue
		}
		
		copiedPaths = append(copiedPaths, path)
	}
	
	if len(copiedPaths) == 0 {
		// No files backed up - remove directory
		os.RemoveAll(backupPath)
		return &SnapshotMetadata{
			Type:       "filesystem",
			Location:   "",
			Timestamp:  time.Now(),
			Reversible: false,
		}, nil
	}
	
	return &SnapshotMetadata{
		Type:          "filesystem",
		Location:      backupPath,
		Timestamp:     time.Now(),
		AffectedPaths: copiedPaths,
		Reversible:    true,
		RestoreCmd:    fmt.Sprintf("cp -r %s/* %s/", backupPath, workingDir),
	}, nil
}

// RestoreSnapshot restores a snapshot
func (s *Snapshotter) RestoreSnapshot(metadata *SnapshotMetadata, workingDir string) error {
	if !metadata.Reversible {
		return fmt.Errorf("snapshot is not reversible")
	}
	
	switch metadata.Type {
	case "git":
		return s.restoreGitSnapshot(metadata, workingDir)
	case "filesystem":
		return s.restoreFilesystemSnapshot(metadata, workingDir)
	default:
		return fmt.Errorf("unknown snapshot type: %s", metadata.Type)
	}
}

// restoreGitSnapshot restores a Git snapshot
func (s *Snapshotter) restoreGitSnapshot(metadata *SnapshotMetadata, workingDir string) error {
	// Checkout the backup branch
	cmd := exec.Command("git", "checkout", metadata.Location)
	cmd.Dir = workingDir
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to restore Git snapshot: %w\n%s", err, output)
	}
	
	return nil
}

// restoreFilesystemSnapshot restores a filesystem snapshot
func (s *Snapshotter) restoreFilesystemSnapshot(metadata *SnapshotMetadata, workingDir string) error {
	if metadata.Location == "" {
		return fmt.Errorf("no backup location")
	}
	
	// Copy files back
	for _, path := range metadata.AffectedPaths {
		srcPath := filepath.Join(metadata.Location, path)
		destPath := filepath.Join(workingDir, path)
		
		if err := copyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to restore %s: %w", path, err)
		}
	}
	
	return nil
}

// CleanupOldSnapshots removes snapshots older than the specified duration
func (s *Snapshotter) CleanupOldSnapshots(maxAge time.Duration) error {
	if _, err := os.Stat(s.backupDir); os.IsNotExist(err) {
		return nil // No backups directory
	}
	
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		return err
	}
	
	cutoff := time.Now().Add(-maxAge)
	
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		if info.ModTime().Before(cutoff) {
			path := filepath.Join(s.backupDir, entry.Name())
			os.RemoveAll(path)
		}
	}
	
	return nil
}

// Helper functions

func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

func getCurrentGitBranch(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	return strings.TrimSpace(string(output)), nil
}

func copyFile(src, dst string) error {
	sourceData, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	return os.WriteFile(dst, sourceData, 0644)
}
