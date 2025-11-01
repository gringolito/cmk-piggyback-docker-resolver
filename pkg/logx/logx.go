package logx

import (
	"fmt"
	"log"
)

// Logger is a tiny structured logger used by the project.
// It wraps the standard library logger and provides convenience methods
// for different log levels.
type Logger struct {
	std *log.Logger
}

// New returns a configured Logger using the default standard logger.
func New() *Logger {
	return &Logger{std: log.Default()}
}

// Info logs an informational message with optional key/value pairs.
func (l *Logger) Info(msg string, kv ...any) {
	l.kv("INFO", msg, kv...)
}

// Warn logs a non-fatal warning message and an associated error.
func (l *Logger) Warn(msg string, err error, kv ...any) {
	l.kv("WARN", msg, append([]any{"err", err}, kv...)...)
}

// Error logs an error message and the associated error value.
func (l *Logger) Error(msg string, err error, kv ...any) {
	l.kv("ERROR", msg, append([]any{"err", err}, kv...)...)
}

func (l *Logger) kv(level, msg string, kv ...any) {
	pairs := ""
	for i := 0; i+1 < len(kv); i += 2 {
		pairs += fmt.Sprintf(" %v=%v", kv[i], kv[i+1])
	}
	l.std.Printf("level=%s msg=%q%s", level, msg, pairs)
}
