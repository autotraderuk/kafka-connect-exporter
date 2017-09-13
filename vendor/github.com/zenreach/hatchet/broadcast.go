package hatchet

import (
	"fmt"
	"strings"
)

// Broadcast creates a logger which logs messages to all of the given loggers.
// Log messages are copied so as to avoid cross-talk and race conditions.
func Broadcast(loggers ...Logger) Logger {
	return &broadcastLogger{Loggers: loggers}
}

// broadcastLogger logs to all of the provided loggers.
type broadcastLogger struct {
	Loggers []Logger
}

// Log broadcasts a message to all of the loggers. Nil logs are discarded.
func (l *broadcastLogger) Log(log map[string]interface{}) {
	if log == nil {
		return
	}
	for _, logger := range l.Loggers {
		logger.Log(L(log).Copy())
	}
}

// Close closes each of the underlying loggers.
func (l *broadcastLogger) Close() error {
	var errs []error
	for _, logger := range l.Loggers {
		if err := logger.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		msgs := make([]string, len(errs))
		for n, err := range errs {
			msgs[n] = err.Error()
		}
		return fmt.Errorf("failed to close %d loggers: %s", len(errs), strings.Join(msgs, "; "))
	}
	return nil
}
