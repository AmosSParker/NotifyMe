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

// LogWithLevel logs a message with the given numeric log level
func (l *Logger) LogWithLevel(level int, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	switch level {
	case LevelInfo:
		l.Info(message)
	case LevelWarn:
		l.Warn(message)
	case LevelError:
		l.Error(message)
	case LevelCritical:
		l.Critical(message)
	default:
		l.Error(fmt.Sprintf("Unknown log level: %d", level))
	}
}

// Info logs an informational message if the log level is set to INFO or lower
func (l *Logger) Info(message string) {
	if l.level <= LevelInfo {
		l.infoLogger.Println(message)
	}
}

// Warn logs a warning message if the log level is set to WARN or lower
func (l *Logger) Warn(message string) {
	if l.level <= LevelWarn {
		l.warnLogger.Println(message)
	}
}

// Error logs an error message if the log level is set to ERROR or lower
func (l *Logger) Error(message string) {
	if l.level <= LevelError {
		l.errorLogger.Println(message)
	}
}

// Critical logs a critical message if the log level is set to CRITICAL or lower
func (l *Logger) Critical(message string) {
	if l.level <= LevelCritical {
		l.criticalLogger.Println(message)
	}
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
		globalLogger.Info(formattedMessage)
	case "Warn":
		globalLogger.Warn(formattedMessage)
	case "Error":
		globalLogger.Error(formattedMessage)
	case "Critical":
		globalLogger.Critical(formattedMessage)
	default:
		globalLogger.Error("Unknown message type: " + messageType)
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
