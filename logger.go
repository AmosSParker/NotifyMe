package notifyme

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Logger struct holds different loggers for various log levels
type Logger struct {
	infoLogger     *log.Logger
	warnLogger     *log.Logger
	errorLogger    *log.Logger
	criticalLogger *log.Logger
	level          int        // Current logging level
	mu             sync.Mutex // Mutex to make logging concurrency safe
}

// Log levels constants
const (
	LevelInfo = iota
	LevelWarn
	LevelError
	LevelCritical
)

// NewLogger creates and returns a new Logger instance
// It takes an optional output destination (like a file path) for logging
func NewLogger(level int, output ...string) *Logger {
	// Default to stdout if no output file is specified
	var logOutput *os.File
	if len(output) > 0 {
		var err error
		logOutput, err = os.OpenFile(output[0], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
	} else {
		logOutput = os.Stdout
	}

	// Initialize loggers for each level
	return &Logger{
		infoLogger:     log.New(logOutput, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLogger:     log.New(logOutput, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger:    log.New(logOutput, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		criticalLogger: log.New(logOutput, "CRITICAL: ", log.Ldate|log.Ltime|log.Lshortfile),
		level:          level,
	}
}

// SetLevel changes the logging level of the logger instance
func (l *Logger) SetLevel(level int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// logMessage formats and logs a message
// This method is concurrency safe due to the mutex lock
func (l *Logger) logMessage(logger *log.Logger, level string, message string, context string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Format the message with additional arguments if provided
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	// Create a log entry with a timestamp, level, message, and context
	timestamp := time.Now().Format(time.RFC3339)
	logEntry := fmt.Sprintf("%s [%s] %s - %s", timestamp, level, message, context)
	logger.Println(logEntry)
}

// Info logs an informational message if the log level is set to INFO or lower
func (l *Logger) Info(message string, context string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.logMessage(l.infoLogger, "INFO", message, context, args...)
	}
}

// Warn logs a warning message if the log level is set to WARN or lower
func (l *Logger) Warn(message string, context string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.logMessage(l.warnLogger, "WARN", message, context, args...)
	}
}

// Error logs an error message if the log level is set to ERROR or lower
func (l *Logger) Error(message string, context string, args ...interface{}) {
	if l.level <= LevelError {
		l.logMessage(l.errorLogger, "ERROR", message, context, args...)
	}
}

// Critical logs a critical message if the log level is set to CRITICAL or lower
// Returns an error with the message
func (l *Logger) Critical(message string, context string, args ...interface{}) error {
	if l.level <= LevelCritical {
		l.logMessage(l.criticalLogger, "CRITICAL", message, context, args...)
	}
	return fmt.Errorf(message, args...)
}

// Notify handles logging based on the message type
// It uses the appropriate logger based on the type (Info, Warn, Error, Critical)
func (l *Logger) Notify(messageType string, message string, context ...interface{}) {
	var formattedMessage string

	// Format the message with the provided context if any
	if len(context) > 0 {
		formattedMessage = fmt.Sprintf(message, context...)
	} else {
		formattedMessage = message
	}

	// Switch case to handle different message types
	switch messageType {
	case "Info":
		l.Info(formattedMessage, "Notify")
	case "Warn":
		l.Warn(formattedMessage, "Notify")
	case "Error":
		l.Error(formattedMessage, "Notify")
	case "Critical":
		l.Critical(formattedMessage, "Notify")
	default:
		l.Error("Unknown message type:", "Notify")
	}
}

// InitFromEnv sets the log level based on an environment variable
// It looks for LOG_LEVEL and sets the level accordingly
func (l *Logger) InitFromEnv() {
	if logLevel, exists := os.LookupEnv("LOG_LEVEL"); exists {
		switch logLevel {
		case "INFO":
			l.SetLevel(LevelInfo)
		case "WARN":
			l.SetLevel(LevelWarn)
		case "ERROR":
			l.SetLevel(LevelError)
		case "CRITICAL":
			l.SetLevel(LevelCritical)
		default:
			l.SetLevel(LevelError) // Default level if an unknown value is found
		}
	}
}
