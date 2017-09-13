package hatchet_test

import (
	"reflect"
	"testing"

	"github.com/zenreach/hatchet"
)

func TestAsyncLogger(t *testing.T) {
	t.Parallel()

	l := hatchet.Mock()
	a := hatchet.Async(l)

	msg := "test log"
	a.Log(hatchet.L{"message": msg})
	if err := a.Close(); err != nil {
		t.Error(err)
	}

	want := hatchet.L{"message": msg}
	if len(l.Logs) != 1 {
		t.Errorf("wrong number of logs sent %d", len(l.Logs))
	} else if !reflect.DeepEqual(l.Logs[0], want) {
		t.Error("message is incorrect")
		t.Errorf("  have %+v", l.Logs[0])
		t.Errorf("  want %+v", want)
	}
}
