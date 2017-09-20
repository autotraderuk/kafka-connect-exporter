package hatchet_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/zenreach/hatchet"
)

func TestConsole(t *testing.T) {
	t.Parallel()

	type Test struct {
		log hatchet.L
		out string
		tpl string
	}

	timestamp := time.Now()
	timeStr := timestamp.Format(time.RFC3339)
	tests := []Test{
		{
			log: hatchet.L{
				"message": "hello",
				"level":   hatchet.InfoLevel,
				"time":    timestamp,
			},
			out: fmt.Sprintf("%s info: hello\n", timeStr),
		},
		{
			log: hatchet.L{
				"message": "oops",
				"level":   hatchet.ErrorLevel,
				"time":    timestamp,
				"error":   errors.New("something broke"),
			},
			out: fmt.Sprintf("%s error: oops: something broke\n", timeStr),
		},
		{
			log: hatchet.L{
				"message": "custom template",
				"prefix":  "message",
				"level":   hatchet.DebugLevel,
				"time":    timestamp,
			},
			tpl: "{{ .prefix }}: {{ .message }}\n",
			out: "message: custom template\n",
		},
	}

	for n := range tests {
		test := tests[n]
		t.Run(fmt.Sprintf("test%d", n), func(t *testing.T) {
			buf := &bytes.Buffer{}
			var logger hatchet.Logger
			if test.tpl == "" {
				logger = hatchet.Console(buf)
			} else {
				var err error
				logger, err = hatchet.ConsoleWithTemplate(buf, test.tpl)
				if err != nil {
					t.Errorf("could not create logger: %s", err)
					return
				}
			}
			logger.Log(test.log)
			if err := logger.Close(); err != nil {
				t.Errorf("close returned error: %s", err)
			}
			out := buf.String()
			if out != test.out {
				t.Errorf("logged wrong output:")
				t.Errorf("  have: %s", out)
				t.Errorf("  want: %s", test.out)
			}
		})
	}
}
