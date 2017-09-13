package hatchet_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/zenreach/hatchet"
)

func TestJSON(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	l := hatchet.JSON(buf)

	timestamp := time.Now()
	l.Log(hatchet.L{
		"message": "test log",
		"time":    timestamp,
		"extra":   3,
		"error":   errors.New("oops"),
	})

	have := buf.String()
	want := fmt.Sprintf(
		"{\"error\":\"oops\",\"extra\":3,\"message\":\"test log\",\"time\":\"%s\"}\n",
		timestamp.Format(time.RFC3339Nano),
	)
	if have != want {
		t.Error("message is incorrect")
		t.Errorf("have: %s", have)
		t.Errorf("want: %s", want)
	}
}
