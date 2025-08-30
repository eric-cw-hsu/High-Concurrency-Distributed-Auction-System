package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	loggerDomain "eric-cw-hsu.github.io/scalable-auction-system/internal/domain/logger"
	"github.com/sirupsen/logrus"
)

// FileStorage implements loggerDomain.LogStorage for file-based storage
type FileStorage struct {
	mu          sync.RWMutex
	baseDir     string
	currentFile *os.File
	writer      *bufio.Writer
	maxFileSize int64
	logger      *config.Logger
}

// FileStorageConfig holds configuration for file storage
type FileStorageConfig struct {
	BaseDir     string // Base directory for log files
	MaxFileSize int64  // Maximum file size in bytes before rotation
}

// NewFileStorage creates a new file storage instance
func NewFileStorage(config FileStorageConfig, logger *config.Logger) (*FileStorage, error) {
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
		logger:      logger,
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
			fs.logger.WithError(err).Error("Failed to rotate log file")
			return fmt.Errorf("failed to rotate log file: %w", err)
		}
	}

	// Convert logrus.Level to string for JSON serialization
	entryForJSON := struct {
		ID        string                 `json:"id"`
		Timestamp time.Time              `json:"timestamp"`
		Level     string                 `json:"level"`
		Service   string                 `json:"service"`
		EventType string                 `json:"event_type"`
		Message   string                 `json:"message"`
		UserID    string                 `json:"user_id,omitempty"`
		TraceID   string                 `json:"trace_id,omitempty"`
		Metadata  map[string]interface{} `json:"metadata,omitempty"`
	}{
		ID:        entry.ID,
		Timestamp: entry.Timestamp,
		Level:     entry.Level.String(),
		Service:   entry.Service,
		EventType: entry.EventType,
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

// Query retrieves log entries based on filter criteria
func (fs *FileStorage) Query(ctx context.Context, filter *loggerDomain.LogFilter) ([]*loggerDomain.LogEntry, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var results []*loggerDomain.LogEntry
	count := 0

	// Get all log files in directory
	files, err := fs.getLogFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get log files: %w", err)
	}

	// Process files in reverse chronological order for recent logs first
	for i := len(files) - 1; i >= 0; i-- {
		if count >= filter.Limit+filter.Offset {
			break
		}

		entries, err := fs.queryFile(files[i], filter)
		if err != nil {
			fs.logger.WithError(err).WithField("file", files[i]).Warn("Failed to query log file")
			continue
		}

		results = append(results, entries...)
		count += len(entries)
	}

	// Apply pagination
	start := filter.Offset
	if start > len(results) {
		return []*loggerDomain.LogEntry{}, nil
	}

	end := start + filter.Limit
	if end > len(results) {
		end = len(results)
	}

	return results[start:end], nil
}

// Close closes the file storage
func (fs *FileStorage) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.writer != nil {
		if err := fs.writer.Flush(); err != nil {
			fs.logger.WithError(err).Error("Failed to flush writer during close")
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

	fs.logger.WithField("file", filepath).Info("Rotated to new log file")
	return nil
}

// getLogFiles returns all log files sorted by modification time
func (fs *FileStorage) getLogFiles() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(fs.baseDir, "audit-*.jsonl"))
	if err != nil {
		return nil, err
	}
	return files, nil
}

// queryFile queries a specific log file
func (fs *FileStorage) queryFile(filename string, filter *loggerDomain.LogFilter) ([]*loggerDomain.LogEntry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []*loggerDomain.LogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var rawEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &rawEntry); err != nil {
			continue // Skip malformed lines
		}

		entry, err := fs.parseLogEntry(rawEntry)
		if err != nil {
			continue // Skip entries that can't be parsed
		}

		if fs.matchesFilter(entry, filter) {
			results = append(results, entry)
		}
	}

	return results, scanner.Err()
}

// parseLogEntry converts raw JSON to LogEntry
func (fs *FileStorage) parseLogEntry(raw map[string]interface{}) (*loggerDomain.LogEntry, error) {
	entry := &loggerDomain.LogEntry{}

	if id, ok := raw["id"].(string); ok {
		entry.ID = id
	}

	if timestampStr, ok := raw["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			entry.Timestamp = t
		}
	}

	if levelStr, ok := raw["level"].(string); ok {
		entry.Level = loggerDomain.ParseLogLevel(levelStr)
	}

	if service, ok := raw["service"].(string); ok {
		entry.Service = service
	}

	if eventType, ok := raw["event_type"].(string); ok {
		entry.EventType = eventType
	}

	if message, ok := raw["message"].(string); ok {
		entry.Message = message
	}

	if userID, ok := raw["user_id"].(string); ok {
		entry.UserID = userID
	}

	if traceID, ok := raw["trace_id"].(string); ok {
		entry.TraceID = traceID
	}

	if metadata, ok := raw["metadata"].(map[string]interface{}); ok {
		entry.Metadata = logrus.Fields(metadata)
	}

	return entry, nil
}

// matchesFilter checks if an entry matches the filter criteria
func (fs *FileStorage) matchesFilter(entry *loggerDomain.LogEntry, filter *loggerDomain.LogFilter) bool {
	// Time range filter
	if filter.StartTime != nil && entry.Timestamp.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && entry.Timestamp.After(*filter.EndTime) {
		return false
	}

	// Service filter
	if len(filter.Services) > 0 {
		found := false
		for _, service := range filter.Services {
			if entry.Service == service {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Event type filter
	if len(filter.EventTypes) > 0 {
		found := false
		for _, eventType := range filter.EventTypes {
			if entry.EventType == eventType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// User ID filter
	if len(filter.UserIDs) > 0 {
		found := false
		for _, userID := range filter.UserIDs {
			if entry.UserID == userID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Trace ID filter
	if len(filter.TraceIDs) > 0 {
		found := false
		for _, traceID := range filter.TraceIDs {
			if entry.TraceID == traceID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Level filter
	if len(filter.Levels) > 0 {
		found := false
		entryLevelStr := entry.Level.String()
		for _, level := range filter.Levels {
			if entryLevelStr == level {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
