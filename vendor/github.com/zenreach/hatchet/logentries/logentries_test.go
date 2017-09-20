package logentries

import (
	"bytes"
	"testing"

	"github.com/zenreach/hatchet"
)

type mockWriter struct {
	bytes.Buffer
	Writes int
	Closed bool
}

func (w *mockWriter) Write(p []byte) (int, error) {
	w.Writes++
	return w.Buffer.Write(p)
}

func (w *mockWriter) Close() error {
	w.Closed = true
	return nil
}

func TestLog(t *testing.T) {
	type Test struct {
		Log    hatchet.L
		Out    string
		Writes int
	}

	tests := []Test{
		{
			Log:    hatchet.L{},
			Out:    "",
			Writes: 0,
		},
		{
			Log:    nil,
			Out:    "",
			Writes: 0,
		},
		{
			Log:    hatchet.L{"message": "hello, world!"},
			Out:    `{"message":"hello, world!"}`,
			Writes: 1,
		},
		{
			Log:    hatchet.L{"error": "oops"},
			Out:    `{"error":"oops"}`,
			Writes: 1,
		},
		{
			Log:    hatchet.L{"message": "item count", "count": 23},
			Out:    `{"count":23,"message":"item count"}`,
			Writes: 1,
		},
	}

	for n, test := range tests {
		buf := &mockWriter{}
		logger := &logentriesLogger{writer: buf}
		logger.Log(test.Log)
		logger.Close()
		if buf.Writes != test.Writes {
			t.Errorf("test %d write count incorrect: %d != %d", n, buf.Writes, test.Writes)
		}
		if !buf.Closed {
			t.Errorf("test %d writer not closed", n)
		}
		written := buf.String()
		if written != test.Out {
			t.Errorf("test %d incorrect output: %s != %s", n, written, test.Out)
		}
	}
}

func TestLogMulti(t *testing.T) {
	logs := []hatchet.L{
		{"message": "hello"},
		nil,
		{"message": "item count", "count": 42},
		{"message": "error", "error": "oops"},
		{},
	}
	out := `{"message":"hello"}{"count":42,"message":"item count"}{"error":"oops","message":"error"}`
	writes := 3

	buf := &mockWriter{}
	logger := &logentriesLogger{writer: buf}
	for _, log := range logs {
		logger.Log(log)
	}
	logger.Close()

	if buf.Writes != writes {
		t.Errorf("write count incorrect: %d != %d", buf.Writes, writes)
	}
	if !buf.Closed {
		t.Error("writer not closed")
	}
	written := buf.String()
	if written != out {
		t.Errorf("incorrect output: %s != %s", written, out)
	}
}

func TestLogClosed(t *testing.T) {
	buf := &mockWriter{}
	logger := &logentriesLogger{writer: buf}

	logger.Close()

	// Write-after-close should not panic.
	logger.Log(hatchet.L{"message": "closed"})
}
