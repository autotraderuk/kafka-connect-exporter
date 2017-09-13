package hatchet_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/zenreach/hatchet"
)

type mockTest struct {
	logs []interface{}
}

func (m *mockTest) Log(args ...interface{}) {
	m.logs = append(m.logs, args...)
}

func TestTest(t *testing.T) {
	t.Parallel()

	m := &mockTest{}
	l := hatchet.Test(m)

	timestamp := time.Now()
	l.Log(hatchet.L{
		"message": "test log",
		"time":    timestamp,
		"extra":   3,
		"error":   errors.New("oops"),
	})

	want := []interface{}{
		"{",
		"  \"error\": \"oops\",",
		"  \"extra\": 3,",
		"  \"message\": \"test log\",",
		fmt.Sprintf("  \"time\": \"%s\"", timestamp.Format(time.RFC3339Nano)),
		"}",
	}
	if !reflect.DeepEqual(want, m.logs) {
		t.Error("messages are incorrect")
		t.Errorf("have: %+v", m.logs)
		t.Errorf("want: %+v", want)
	}
}
