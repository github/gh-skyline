// Package logger provides thread-safe logging capabilities with different severity levels.
package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// LogLevel represents the severity level of a log message
type LogLevel int

// Log levels ordered by increasing severity
const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

// String returns the string representation of a LogLevel
func (l LogLevel) String() string {
	return [...]string{"DEBUG", "INFO", "WARNING", "ERROR"}[l]
}

// Logger provides thread-safe logging capabilities with different severity levels
type Logger struct {
	debug   *log.Logger
	info    *log.Logger
	warning *log.Logger
	error   *log.Logger
	level   LogLevel
	mu      sync.Mutex
}

var (
	instance *Logger
	once     sync.Once
)

// GetLogger returns a singleton instance of the Logger
// It initializes the logger on first call using sync.Once
func GetLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			debug:   log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
			info:    log.New(os.Stdout, "", 0),
			warning: log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime),
			error:   log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime),
			level:   INFO,
		}
	})
	return instance
}

// SetLevel changes the minimum severity level for logging
// Thread-safe through mutex locking
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// logf is an internal helper that handles mutex locking and level checking
func (l *Logger) logf(level LogLevel, format string, v ...interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.level <= level {
		msg := fmt.Sprintf(format, v...)
		var err error

		switch level {
		case DEBUG:
			err = l.debug.Output(3, msg)
		case INFO:
			err = l.info.Output(2, msg)
		case WARNING:
			err = l.warning.Output(2, msg)
		case ERROR:
			err = l.error.Output(2, msg)
		}
		return err
	}
	return nil
}

// Debug logs a debug-level message
func (l *Logger) Debug(format string, v ...interface{}) error {
	return l.logf(DEBUG, format, v...)
}

// Info logs an info-level message
func (l *Logger) Info(format string, v ...interface{}) error {
	return l.logf(INFO, format, v...)
}

// Warning logs a warning-level message
func (l *Logger) Warning(format string, v ...interface{}) error {
	return l.logf(WARNING, format, v...)
}

// Error logs an error-level message
func (l *Logger) Error(format string, v ...interface{}) error {
	return l.logf(ERROR, format, v...)
}
