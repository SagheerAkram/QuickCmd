package executor

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UndoEngine manages undo/rollback functionality
type UndoEngine struct {
	backupDir string
	strategies map[string]UndoStrategy
}

// UndoStrategy defines how to undo a specific type of command
type UndoStrategy interface {
	CanUndo(command string) bool
	CreateBackup(command string, affectedPaths []string) (*UndoRecord, error)
	Undo(record *UndoRecord) error
}

// UndoRecord represents an undoable operation
type UndoRecord struct {
	ID              string
	Command         string
	UndoCommand     string
	BackupLocation  string
	BackupSize      int64
	CanUndo         bool
	Strategy        string
	AffectedPaths   []string
	CreatedAt       time.Time
	ExpiresAt       time.Time
}

// NewUndoEngine creates a new undo engine
func NewUndoEngine(backupDir string) *UndoEngine {
	ue := &UndoEngine{
		backupDir:  backupDir,
		strategies: make(map[string]UndoStrategy),
	}
	
	// Register strategies
	ue.strategies["file"] = &FileUndoStrategy{backupDir: backupDir}
	ue.strategies["git"] = &GitUndoStrategy{}
	ue.strategies["kubectl"] = &KubectlUndoStrategy{}
	
	return ue
}

// CreateUndo creates an undo record for a command
func (ue *UndoEngine) CreateUndo(command string, affectedPaths []string) (*UndoRecord, error) {
	// Determine strategy
	var strategy UndoStrategy
	var strategyName string
	
	if strings.HasPrefix(command, "rm") || strings.HasPrefix(command, "mv") {
		strategy = ue.strategies["file"]
		strategyName = "file"
	} else if strings.HasPrefix(command, "git") {
		strategy = ue.strategies["git"]
		strategyName = "git"
	} else if strings.HasPrefix(command, "kubectl") {
		strategy = ue.strategies["kubectl"]
		strategyName = "kubectl"
	}
	
	if strategy == nil || !strategy.CanUndo(command) {
		return &UndoRecord{
			Command: command,
			CanUndo: false,
		}, nil
	}
	
	// Create backup
	record, err := strategy.CreateBackup(command, affectedPaths)
	if err != nil {
		return nil, err
	}
	
	record.Command = command
	record.Strategy = strategyName
	record.CreatedAt = time.Now()
	record.ExpiresAt = time.Now().Add(7 * 24 * time.Hour) // 7 days
	
	return record, nil
}

// Undo performs an undo operation
func (ue *UndoEngine) Undo(record *UndoRecord) error {
	if !record.CanUndo {
		return fmt.Errorf("command cannot be undone")
	}
	
	strategy := ue.strategies[record.Strategy]
	if strategy == nil {
		return fmt.Errorf("unknown undo strategy: %s", record.Strategy)
	}
	
	return strategy.Undo(record)
}

// FileUndoStrategy handles file operation undos
type FileUndoStrategy struct {
	backupDir string
}

func (s *FileUndoStrategy) CanUndo(command string) bool {
	return strings.HasPrefix(command, "rm") || strings.HasPrefix(command, "mv")
}

func (s *FileUndoStrategy) CreateBackup(command string, affectedPaths []string) (*UndoRecord, error) {
	if len(affectedPaths) == 0 {
		return &UndoRecord{CanUndo: false}, nil
	}
	
	// Create backup tar.gz
	timestamp := time.Now().Format("20060102-150405")
	backupFile := filepath.Join(s.backupDir, fmt.Sprintf("backup-%s.tar.gz", timestamp))
	
	// Ensure backup directory exists
	os.MkdirAll(s.backupDir, 0755)
	
	// Create tar.gz
	size, err := createTarGz(backupFile, affectedPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}
	
	return &UndoRecord{
		CanUndo:        true,
		BackupLocation: backupFile,
		BackupSize:     size,
		AffectedPaths:  affectedPaths,
		UndoCommand:    fmt.Sprintf("tar -xzf %s -C /", backupFile),
	}, nil
}

func (s *FileUndoStrategy) Undo(record *UndoRecord) error {
	// Extract backup
	return extractTarGz(record.BackupLocation, "/")
}

// GitUndoStrategy handles git operation undos
type GitUndoStrategy struct{}

func (s *GitUndoStrategy) CanUndo(command string) bool {
	return strings.Contains(command, "git reset") || 
	       strings.Contains(command, "git push --force")
}

func (s *GitUndoStrategy) CreateBackup(command string, affectedPaths []string) (*UndoRecord, error) {
	// For git operations, save reflog state
	timestamp := time.Now().Format("20060102-150405")
	backupBranch := fmt.Sprintf("backup-%s", timestamp)
	
	undoCmd := ""
	if strings.Contains(command, "git reset") {
		undoCmd = "git reset --hard ORIG_HEAD"
	} else if strings.Contains(command, "git push --force") {
		undoCmd = fmt.Sprintf("git push --force origin %s:main", backupBranch)
	}
	
	return &UndoRecord{
		CanUndo:     true,
		UndoCommand: undoCmd,
	}, nil
}

func (s *GitUndoStrategy) Undo(record *UndoRecord) error {
	// Execute undo command
	// TODO: Implement command execution
	return nil
}

// KubectlUndoStrategy handles kubectl operation undos
type KubectlUndoStrategy struct{}

func (s *KubectlUndoStrategy) CanUndo(command string) bool {
	return strings.Contains(command, "kubectl delete")
}

func (s *KubectlUndoStrategy) CreateBackup(command string, affectedPaths []string) (*UndoRecord, error) {
	// Save YAML manifest before deletion
	// TODO: Implement kubectl get -o yaml
	
	return &UndoRecord{
		CanUndo:     true,
		UndoCommand: "kubectl apply -f backup.yaml",
	}, nil
}

func (s *KubectlUndoStrategy) Undo(record *UndoRecord) error {
	// Apply saved manifest
	// TODO: Implement kubectl apply
	return nil
}

// Helper functions

func createTarGz(filename string, paths []string) (int64, error) {
	file, err := os.Create(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	
	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()
	
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()
	
	for _, path := range paths {
		if err := addToTar(tarWriter, path); err != nil {
			return 0, err
		}
	}
	
	info, _ := file.Stat()
	return info.Size(), nil
}

func addToTar(tw *tar.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	info, err := file.Stat()
	if err != nil {
		return err
	}
	
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = path
	
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	
	if !info.IsDir() {
		if _, err := io.Copy(tw, file); err != nil {
			return err
		}
	}
	
	return nil
}

func extractTarGz(filename, dest string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()
	
	tarReader := tar.NewReader(gzReader)
	
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		
		target := filepath.Join(dest, header.Name)
		
		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, 0755)
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0755)
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	
	return nil
}
