package hatchet

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

// Async wraps a logger with asynchronous functionality. An asynchronous logger
// does not block on calls to Log. Calling Close blocks until all asynchronous
// actions complete.
func Async(logger Logger) Logger {
	return &asyncLogger{Logger: logger}
}

// asyncLogger wraps a logger with asynchronous functionality.
type asyncLogger struct {
	Logger
	wg sync.WaitGroup
}

// Log calls the underlying Log function asynchronously in a goroutine.
func (l *asyncLogger) Log(log map[string]interface{}) {
	l.wg.Add(1)
	go func(wg *sync.WaitGroup, log L) {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = errors.Errorf("%+v", r)
				}
				PrintFailure(log, errors.Wrap(err, "failed to write in async logger"))
			}
			wg.Done()
		}()
		l.Logger.Log(log)
	}(&l.wg, log)
}

// Close waits for all logs to flush before returning. It will wait no longer
// than 30s, and any calls to Log() while Close() is executing may or may not
// be honoured. An error is returned if the timeout is exceeded while waiting.
// The underlying logger is closed after flushing.
func (l *asyncLogger) Close() error {
	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	// Between the call to wg.Wait() and Logger.Close(), new calls to Log() may
	// have been made. We are trusting the underlying logger not to panic on
	// log-after-close.
	select {
	case <-time.After(30 * time.Second):
		l.Logger.Close() // ignore errors
		return errors.New("timed out closing async logger")
	case <-done:
		return l.Logger.Close()
	}
}
