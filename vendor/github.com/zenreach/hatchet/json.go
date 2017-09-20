package hatchet

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/pkg/errors"
)

// JSON creates a logger which encodes messages as JSON and writes them to the
// given writer, one message per line. The "error" field is converted to a
// string before writing. Empty logs are discarded. The writer is closed when
// `Close` is called on the logger.
func JSON(wr io.Writer) Logger {
	return &jsonLogger{
		writer:  wr,
		encoder: json.NewEncoder(wr),
	}
}

type jsonLogger struct {
	writer  io.Writer
	encoder *json.Encoder
	mu      sync.Mutex // controls access to writer and encoder
}

// Log writes log as a JSON object, followed by a newline character.
func (l *jsonLogger) Log(log map[string]interface{}) {
	if len(log) == 0 {
		return
	}

	ll := L(log)
	err := ll.Error()
	if err != nil {
		ll = ll.Copy()
		ll[Error] = err.Error()
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.encoder.Encode(ll); err != nil {
		PrintFailure(ll, errors.Wrap(err, "failed to write in json logger"))
	}
}

// Close the logger.
func (l *jsonLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// We are trusting the underlying writer to not panic on write-after-close.
	if closer, ok := l.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
