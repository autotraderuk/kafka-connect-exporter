package hatchet_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/zenreach/hatchet"
)

// Table test the LineWriter for common cases.
func TestLineWriterTable(t *testing.T) {
	t.Parallel()

	// severely limit the buffer size for testing
	bufferSize := 16

	type Test struct {
		Writes []string
		Logs   []hatchet.L
	}

	tests := []Test{
		{
			// single write, no newline
			Writes: []string{
				"this is a test",
			},
			Logs: []hatchet.L{
				{"message": "this is a test"},
			},
		},
		{
			// single write, one newline
			Writes: []string{
				"hello\nthere",
			},
			Logs: []hatchet.L{
				{"message": "hello"},
				{"message": "there"},
			},
		},
		{
			// single write, multiple newlines
			Writes: []string{
				"hi\nthere\nworld",
			},
			Logs: []hatchet.L{
				{"message": "hi"},
				{"message": "there"},
				{"message": "world"},
			},
		},
		{
			// two writes, no newlines
			Writes: []string{
				"hello, ",
				"world",
			},
			Logs: []hatchet.L{
				{"message": "hello, world"},
			},
		},
		{
			// multiple writes, multiple newlines
			Writes: []string{
				"hi there\nlook ",
				"at me\n",
				"i'm a star\nshine",
			},
			Logs: []hatchet.L{
				{"message": "hi there"},
				{"message": "look at me"},
				{"message": "i'm a star"},
				{"message": "shine"},
			},
		},
		{
			// single write, exceeds buffer size 1x
			Writes: []string{
				"this is a seemingly long lin",
			},
			Logs: []hatchet.L{
				{"message": "this is a seemin"},
				{"message": "gly long lin"},
			},
		},
		{
			// single write, exceeds buffer size 2x
			Writes: []string{
				"this is a surprisingly long line. you might say it is written like a novel.",
			},
			Logs: []hatchet.L{
				{"message": "this is a surpri"},
				{"message": "singly long line"},
				{"message": ". you might say "},
				{"message": "it is written li"},
				{"message": "ke a novel."},
			},
		},
		{
			// multiple writes, exceeds buffer size
			Writes: []string{
				"this is a surprisingly long line. ",
				"this time it is split into multiple writes.",
			},
			Logs: []hatchet.L{
				{"message": "this is a surpri"},
				{"message": "singly long line"},
				{"message": ". this time it i"},
				{"message": "s split into mul"},
				{"message": "tiple writes."},
			},
		},
	}

	for n := range tests {
		test := tests[n]
		t.Run(fmt.Sprintf("test%d", n), func(t *testing.T) {
			mock := hatchet.Mock()
			wr := &hatchet.Writer{
				Logger:     mock,
				BufferSize: bufferSize,
			}
			for j, write := range test.Writes {
				l := len([]byte(write))
				if n, err := wr.Write([]byte(write)); err != nil {
					t.Errorf("error on write %d: %s", j, err)
				} else if n != len([]byte(write)) {
					t.Errorf("write %d has incorrect length: %d != %d", j, n, l)
				}
			}
			if err := wr.Close(); err != nil {
				t.Errorf("error on close: %s", err)
			}
			if !reflect.DeepEqual(mock.Logs, test.Logs) {
				t.Errorf("emitted incorrect logs:")
				t.Errorf("  have: %+v", mock.Logs)
				t.Errorf("  want: %+v", test.Logs)
			}
		})
	}
}
