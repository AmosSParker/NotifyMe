package notifyme

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

// Logger struct holds different loggers for various log levels
type Logger struct {
	infoLogger     *log.Logger
	warnLogger     *log.Logger
	errorLogger    *log.Logger
	criticalLogger *log.Logger
	level          int
	mu             sync.Mutex // Added mutex for thread safety
}

// Global logger instance
var globalLogger *Logger
var once sync.Once // Ensure singleton pattern for global logger

// Log levels constants
const (
	LevelInfo = iota
	LevelWarn
	LevelError
	LevelCritical
)

// newLoggerInstance initializes and returns a new Logger instance
func newLoggerInstance(level int, output ...string) *Logger {
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

// InitializeGlobalLogger creates and initializes the global logger instance
func InitializeGlobalLogger(level int, output ...string) {
	once.Do(func() {
		globalLogger = newLoggerInstance(level, output...)
	})
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	return globalLogger
}

// NewLogger creates and returns a new Logger instance
func NewLogger(level int, output ...string) *Logger {
	return newLoggerInstance(level, output...)
}

// SetLevel sets the global log level
func SetLevel(level int) {
	if globalLogger != nil {
		globalLogger.mu.Lock()
		defer globalLogger.mu.Unlock()
		globalLogger.level = level
	}
}

// Log logs a message with the given log level
func (l *Logger) Log(level int, message string, optionalParams ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fullMessage := message
	for _, param := range optionalParams {
		fullMessage += fmt.Sprintf(" %v", param)
	}
	switch level {
	case LevelInfo:
		if l.level <= LevelInfo {
			logMessage(l.infoLogger, "INFO", fullMessage)
		}
	case LevelWarn:
		if l.level <= LevelWarn {
			logMessage(l.warnLogger, "WARN", fullMessage)
		}
	case LevelError:
		if l.level <= LevelError {
			logMessage(l.errorLogger, "ERROR", fullMessage)
		}
	case LevelCritical:
		if l.level <= LevelCritical {
			logMessage(l.criticalLogger, "CRITICAL", fullMessage)
		}
	default:
		logMessage(l.errorLogger, "ERROR", fmt.Sprintf("Unknown log level: %d", level))
	}
}

// logMessage is a helper function to log the message
func logMessage(logger *log.Logger, level string, message string) {
	logger.Printf("[%s] %s", level, message)
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
		globalLogger.Log(LevelInfo, formattedMessage)
	case "Warn":
		globalLogger.Log(LevelWarn, formattedMessage)
	case "Error":
		globalLogger.Log(LevelError, formattedMessage)
	case "Critical":
		globalLogger.Log(LevelCritical, formattedMessage)
	default:
		globalLogger.Log(LevelError, "Unknown message type: "+messageType)
	}
}

// InitFromEnv sets the log level based on an environment variable
func InitFromEnv() {
	if logLevel, exists := os.LookupEnv("LOG_LEVEL"); exists {
		switch logLevel {
		case "INFO":
			SetLevel(LevelInfo)
		case "WARN":
			SetLevel(LevelWarn)
		case "ERROR":
			SetLevel(LevelError)
		case "CRITICAL":
			SetLevel(LevelCritical)
		default:
			SetLevel(LevelError) // Default level if an unknown value is found
		}
	}
}

func (l *Logger) MarshalJSON() ([]byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return json.Marshal(&struct {
		Level int `json:"level"`
	}{
		Level: l.level,
	})
}

func (l *Logger) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Level int `json:"level"`
	}{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = aux.Level
	l.infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	l.warnLogger = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	l.errorLogger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	l.criticalLogger = log.New(os.Stdout, "CRITICAL: ", log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}
