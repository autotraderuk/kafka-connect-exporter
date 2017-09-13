package logentries

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"
	"github.com/zenreach/hatchet"
	"github.com/zenreach/le_go"
)

// TODO Contribute our github.com/zenreach/le_go modifications back to github.com/bsphere/le_go

// New creates a logger which persists messages to Logentries. Log messages are encoded as JSON
// before being sent to Logentries. Authentication requires a valid Logentries token. An error is
// returned on connection failure.
func New(token string) (hatchet.Logger, error) {
	le, err := le_go.Connect(token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to Logentries")
	}

	return &logentriesLogger{
		writer: le,
	}, nil
}

// logentriesLogger is a logutil.Logger which persists messages to Logentries.
type logentriesLogger struct {
	writer io.WriteCloser
}

// Log encodes a message as JSON and sends it to Logentries.
func (l *logentriesLogger) Log(log map[string]interface{}) {
	if len(log) == 0 {
		return
	}

	ll := hatchet.L(log).Copy()
	if err := ll.Error(); err != nil {
		ll[hatchet.Error] = err.Error()
	}
	if timestamp := ll.Time(); !timestamp.IsZero() {
		ll[hatchet.Time] = timestamp
	}

	// Marshal instead of using a central encoder so we don't require locking
	buf, err := json.Marshal(ll)
	if err != nil {
		hatchet.PrintFailure(ll, errors.Wrap(err, "failed to encode Logentries message"))
		return
	}

	// The Logentries writer is thread safe
	if _, err := l.writer.Write(buf); err != nil {
		hatchet.PrintFailure(ll, errors.Wrap(err, "failed to write Logentries message"))
	}
}

// Close the connection to Logentries.
func (l *logentriesLogger) Close() error {
	// Write-after-close is handled gracefully by the Logentries logger
	if err := l.writer.Close(); err != nil {
		return errors.Wrap(err, "failed to close Logentries writer")
	}
	return nil
}
