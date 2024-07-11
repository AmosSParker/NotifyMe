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

// Global logger instance
var globalLogger *Logger

// Log levels constants
const (
	LevelInfo = iota
	LevelWarn
	LevelError
	LevelCritical
)

// InitializeGlobalLogger creates and initializes the global logger instance
func InitializeGlobalLogger(level int, output ...string) {
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
	globalLogger = &Logger{
		infoLogger:     log.New(logOutput, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLogger:     log.New(logOutput, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger:    log.New(logOutput, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		criticalLogger: log.New(logOutput, "CRITICAL: ", log.Ldate|log.Ltime|log.Lshortfile),
		level:          level,
	}
}

// SetLevel changes the logging level of the global logger instance
func SetLevel(level int) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.level = level
}

// logMessage formats and logs a message
func logMessage(logger *log.Logger, level string, message string, context string, args ...interface{}) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()

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
func Info(message string, context string, args ...interface{}) {
	if globalLogger.level <= LevelInfo {
		logMessage(globalLogger.infoLogger, "INFO", message, context, args...)
	}
}

// Warn logs a warning message if the log level is set to WARN or lower
func Warn(message string, context string, args ...interface{}) {
	if globalLogger.level <= LevelWarn {
		logMessage(globalLogger.warnLogger, "WARN", message, context, args...)
	}
}

// Error logs an error message if the log level is set to ERROR or lower
func Error(message string, context string, args ...interface{}) {
	if globalLogger.level <= LevelError {
		logMessage(globalLogger.errorLogger, "ERROR", message, context, args...)
	}
}

// Critical logs a critical message if the log level is set to CRITICAL or lower
func Critical(message string, context string, args ...interface{}) error {
	if globalLogger.level <= LevelCritical {
		logMessage(globalLogger.criticalLogger, "CRITICAL", message, context, args...)
	}
	return fmt.Errorf(message, args...)
}

// Notify handles logging based on the message type
func Notify(messageType string, message string, context ...interface{}) {
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
		Info(formattedMessage, "Notify")
	case "Warn":
		Warn(formattedMessage, "Notify")
	case "Error":
		Error(formattedMessage, "Notify")
	case "Critical":
		Critical(formattedMessage, "Notify")
	default:
		Error("Unknown message type:", "Notify")
	}
}
