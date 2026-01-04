package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// DEBUG level for detailed debugging information
	DEBUG LogLevel = iota
	// INFO level for general informational messages
	INFO
	// WARN level for warning messages
	WARN
	// ERROR level for error messages
	ERROR
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Color returns the ANSI color code for the log level
func (l LogLevel) Color() string {
	switch l {
	case DEBUG:
		return "\033[36m" // Cyan
	case INFO:
		return "\033[32m" // Green
	case WARN:
		return "\033[33m" // Yellow
	case ERROR:
		return "\033[31m" // Red
	default:
		return "\033[0m" // Reset
	}
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Component string
	Message   string
	Fields    map[string]any
}

// Logger provides structured logging with context
type Logger struct {
	mu         sync.RWMutex
	level      LogLevel
	output     *os.File
	entries    []LogEntry
	maxEntries int
	component  string
	fileLog    bool
	logFile    *os.File
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level         LogLevel
	Output        *os.File
	MaxEntries    int
	Component     string
	EnableFileLog bool
	LogDir        string
}

// DefaultLoggerConfig returns default logger configuration
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:         INFO,
		Output:        os.Stdout,
		MaxEntries:    1000,
		Component:     "BettaFish",
		EnableFileLog: false,
		LogDir:        "logs",
	}
}

// NewLogger creates a new logger with the given configuration
func NewLogger(config LoggerConfig) (*Logger, error) {
	logger := &Logger{
		level:      config.Level,
		output:     config.Output,
		maxEntries: config.MaxEntries,
		component:  config.Component,
		entries:    make([]LogEntry, 0),
	}

	if config.EnableFileLog {
		if err := os.MkdirAll(config.LogDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		logFileName := fmt.Sprintf("%s_%s.log",
			config.Component,
			time.Now().Format("20060102_150405"))
		logPath := filepath.Join(config.LogDir, logFileName)

		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		logger.fileLog = true
		logger.logFile = file

		// Write to both console and file
		log.SetOutput(io.MultiWriter(os.Stdout, file))
	}

	return logger, nil
}

// NewSimpleLogger creates a simple logger with default settings
func NewSimpleLogger(component string) *Logger {
	logger, _ := NewLogger(LoggerConfig{
		Level:      INFO,
		Output:     os.Stdout,
		MaxEntries: 1000,
		Component:  component,
	})
	return logger
}

// Close closes the log file if open
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// log internal logging method
func (l *Logger) log(level LogLevel, component string, message string, fields map[string]any) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Component: component,
		Message:   message,
		Fields:    fields,
	}

	l.mu.Lock()
	l.entries = append(l.entries, entry)
	if len(l.entries) > l.maxEntries {
		l.entries = l.entries[len(l.entries)-l.maxEntries:]
	}
	l.mu.Unlock()

	// Format the log message
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")
	color := level.Color()
	reset := "\033[0m"

	var fieldStrs []string
	for k, v := range fields {
		fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", k, v))
	}
	fieldsStr := ""
	if len(fieldStrs) > 0 {
		fieldsStr = " [" + strings.Join(fieldStrs, ", ") + "]"
	}

	logLine := fmt.Sprintf("%s [%s] %s%-5s%s %s: %s%s",
		timestamp,
		component,
		color,
		level.String(),
		reset,
		component,
		message,
		fieldsStr,
	)

	fmt.Fprintln(l.output, logLine)
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields map[string]any) {
	l.log(DEBUG, l.component, message, fields)
}

// Info logs an info message
func (l *Logger) Info(message string, fields map[string]any) {
	l.log(INFO, l.component, message, fields)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields map[string]any) {
	l.log(WARN, l.component, message, fields)
}

// Error logs an error message
func (l *Logger) Error(message string, fields map[string]any) {
	l.log(ERROR, l.component, message, fields)
}

// WithComponent returns a new logger with the specified component name
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		level:      l.level,
		output:     l.output,
		maxEntries: l.maxEntries,
		component:  component,
		entries:    l.entries,
		fileLog:    l.fileLog,
		logFile:    l.logFile,
	}
}

// GetEntries returns all log entries
func (l *Logger) GetEntries() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	entries := make([]LogEntry, len(l.entries))
	copy(entries, l.entries)
	return entries
}

// Clear clears all log entries
func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = make([]LogEntry, 0)
}

// ExecutionTracker tracks execution progress
type ExecutionTracker struct {
	logger    *Logger
	startTime time.Time
	steps     map[string]time.Time
	mu        sync.RWMutex
}

// NewExecutionTracker creates a new execution tracker
func NewExecutionTracker(logger *Logger) *ExecutionTracker {
	return &ExecutionTracker{
		logger:    logger,
		startTime: time.Now(),
		steps:     make(map[string]time.Time),
	}
}

// StartStep marks the start of an execution step
func (t *ExecutionTracker) StartStep(stepName string) {
	t.mu.Lock()
	t.steps[stepName] = time.Now()
	t.mu.Unlock()

	t.logger.Info(fmt.Sprintf("Starting step: %s", stepName), nil)
}

// EndStep marks the end of an execution step
func (t *ExecutionTracker) EndStep(stepName string, fields map[string]any) {
	t.mu.Lock()
	startTime, exists := t.steps[stepName]
	if exists {
		delete(t.steps, stepName)
	}
	t.mu.Unlock()

	if !exists {
		t.logger.Warn(fmt.Sprintf("Step %s was not started", stepName), nil)
		return
	}

	duration := time.Since(startTime)
	if fields == nil {
		fields = make(map[string]any)
	}
	fields["duration_ms"] = duration.Milliseconds()

	t.logger.Info(fmt.Sprintf("Completed step: %s", stepName), fields)
}

// LogProgress logs progress information
func (t *ExecutionTracker) LogProgress(stepName, message string, progress int) {
	t.logger.Info(message, map[string]any{
		"step":     stepName,
		"progress": fmt.Sprintf("%d%%", progress),
	})
}

// GetElapsedTime returns the total elapsed time since tracking started
func (t *ExecutionTracker) GetElapsedTime() time.Duration {
	return time.Since(t.startTime)
}

// ContextKey is used for context values
type ContextKey string

const (
	// LoggerKey is the context key for the logger
	LoggerKey ContextKey = "logger"
	// TrackerKey is the context key for the execution tracker
	TrackerKey ContextKey = "tracker"
)

// WithLogger returns a context with the logger attached
func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

// LoggerFromContext extracts the logger from the context
func LoggerFromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(LoggerKey).(*Logger); ok {
		return logger
	}
	return NewSimpleLogger("default")
}

// WithTracker returns a context with the execution tracker attached
func WithTracker(ctx context.Context, tracker *ExecutionTracker) context.Context {
	return context.WithValue(ctx, TrackerKey, tracker)
}

// TrackerFromContext extracts the execution tracker from the context
func TrackerFromContext(ctx context.Context) *ExecutionTracker {
	if tracker, ok := ctx.Value(TrackerKey).(*ExecutionTracker); ok {
		return tracker
	}
	return NewExecutionTracker(NewSimpleLogger("default"))
}
