package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	loggerDomain "eric-cw-hsu.github.io/scalable-auction-system/internal/domain/logger"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
)

// FileStorage implements loggerDomain.LogStorage for file-based storage
type FileStorage struct {
	mu          sync.RWMutex
	baseDir     string
	currentFile *os.File
	writer      *bufio.Writer
	maxFileSize int64
}

// FileStorageConfig holds configuration for file storage
type FileStorageConfig struct {
	BaseDir     string // Base directory for log files
	MaxFileSize int64  // Maximum file size in bytes before rotation
}

// NewFileStorage creates a new file storage instance
func NewFileStorage(config FileStorageConfig) (*FileStorage, error) {
	if config.BaseDir == "" {
		config.BaseDir = "/app/logs"
	}
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 100 * 1024 * 1024 // 100MB default
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(config.BaseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	fs := &FileStorage{
		baseDir:     config.BaseDir,
		maxFileSize: config.MaxFileSize,
	}

	// Initialize current log file
	if err := fs.rotateFile(); err != nil {
		return nil, fmt.Errorf("failed to initialize log file: %w", err)
	}

	return fs, nil
}

// Store stores a log entry to the file
func (fs *FileStorage) Store(ctx context.Context, entry *loggerDomain.LogEntry) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if file rotation is needed
	if fs.needsRotation() {
		if err := fs.rotateFile(); err != nil {
			logger.Error("Failed to rotate log file", map[string]interface{}{
				"error": err.Error(),
			})
			return fmt.Errorf("failed to rotate log file: %w", err)
		}
	}

	// Convert logrus.Level to string for JSON serialization
	entryForJSON := struct {
		ID        string                 `json:"id"`
		Timestamp time.Time              `json:"timestamp"`
		Level     string                 `json:"level"`
		Service   string                 `json:"service"`
		Operation string                 `json:"operation"`
		Message   string                 `json:"message"`
		UserID    string                 `json:"user_id,omitempty"`
		TraceID   string                 `json:"trace_id,omitempty"`
		Metadata  map[string]interface{} `json:"metadata,omitempty"`
	}{
		ID:        entry.Id,
		Timestamp: entry.Timestamp,
		Level:     entry.Level.String(),
		Service:   entry.Service,
		Operation: entry.Operation,
		Message:   entry.Message,
		UserID:    entry.UserID,
		TraceID:   entry.TraceID,
		Metadata:  map[string]interface{}(entry.Metadata),
	}

	// Write JSON line
	data, err := json.Marshal(entryForJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	if _, err := fs.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	if _, err := fs.writer.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	// Flush immediately for real-time logging
	if err := fs.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush log entry: %w", err)
	}

	return nil
}

// Close closes the file storage
func (fs *FileStorage) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.writer != nil {
		if err := fs.writer.Flush(); err != nil {
			logger.Error("Failed to flush writer on close", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	if fs.currentFile != nil {
		if err := fs.currentFile.Close(); err != nil {
			return fmt.Errorf("failed to close current log file: %w", err)
		}
	}

	return nil
}

// needsRotation checks if the current log file needs rotation
func (fs *FileStorage) needsRotation() bool {
	if fs.currentFile == nil {
		return true
	}

	stat, err := fs.currentFile.Stat()
	if err != nil {
		return true
	}

	return stat.Size() >= fs.maxFileSize
}

// rotateFile creates a new log file
func (fs *FileStorage) rotateFile() error {
	// Close current file if exists
	if fs.writer != nil {
		fs.writer.Flush()
	}
	if fs.currentFile != nil {
		fs.currentFile.Close()
	}

	// Create new file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("audit-%s.jsonl", timestamp)
	filepath := filepath.Join(fs.baseDir, filename)

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create log file %s: %w", filepath, err)
	}

	fs.currentFile = file
	fs.writer = bufio.NewWriter(file)

	logger.Info("Log file rotated", map[string]interface{}{
		"file": filepath,
	})
	return nil
}
