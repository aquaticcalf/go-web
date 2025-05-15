package goweb

import "log"

// Logger interface for custom logging
type Logger interface {
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
	Debug(format string, args ...interface{})
}

// defaultLogger implements the Logger interface
type defaultLogger struct{}

func (l *defaultLogger) Info(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

func (l *defaultLogger) Error(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

func (l *defaultLogger) Debug(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}
