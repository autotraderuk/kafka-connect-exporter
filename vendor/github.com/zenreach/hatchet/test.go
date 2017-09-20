package hatchet

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

// T is used to adapt a testing.T or testing.B object for logging.
type T interface {
	Log(args ...interface{})
}

// Test creates a logger which encode logs as JSON and writes them to the given
// T object.
func Test(t T) Logger {
	return &testLogger{t}
}

type testLogger struct {
	t T
}

// Log encodes the log to JSON and sends it to t.Log().
func (l *testLogger) Log(log map[string]interface{}) {
	if len(log) == 0 {
		return
	}
	ll := L(log)
	err := ll.Error()
	if err != nil {
		ll = ll.Copy()
		ll[Error] = err.Error()
	}
	out, err := json.MarshalIndent(ll, "", "  ")
	if err != nil {
		PrintFailure(ll, errors.Wrap(err, "failed to encode log as json"))
		return
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		l.t.Log(line)
	}
}

// Close does nothing.
func (l *testLogger) Close() error {
	return nil
}
