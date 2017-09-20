package hatchet_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/zenreach/hatchet"
)

func TestWrapperWrite(t *testing.T) {
	t.Parallel()

	l := hatchet.Mock()
	w := &hatchet.StandardLogger{Logger: l}

	msg := "test log"
	n, err := w.Write([]byte(msg))
	if err != nil {
		t.Error(err)
	} else if n != len(msg) {
		t.Errorf("wrong length: %d != %d", n, len(msg))
	}
	if err := w.Close(); err != nil {
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

func TestWrapperPrint(t *testing.T) {
	t.Parallel()

	l := hatchet.Mock()
	w := &hatchet.StandardLogger{Logger: l}

	msg := "test log"
	w.Print(msg)
	if err := w.Close(); err != nil {
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

func TestWrapperPrintln(t *testing.T) {
	t.Parallel()

	l := hatchet.Mock()
	w := &hatchet.StandardLogger{Logger: l}

	msg := "test log"
	w.Println(msg)
	if err := w.Close(); err != nil {
		t.Error(err)
	}

	want := hatchet.L{"message": fmt.Sprintf("%s\n", msg)}
	if len(l.Logs) != 1 {
		t.Errorf("wrong number of logs sent %d", len(l.Logs))
	} else if !reflect.DeepEqual(l.Logs[0], want) {
		t.Error("message is incorrect")
		t.Errorf("  have %+v", l.Logs[0])
		t.Errorf("  want %+v", want)
	}
}

func TestWrapperPrintf(t *testing.T) {
	t.Parallel()

	l := hatchet.Mock()
	w := &hatchet.StandardLogger{Logger: l}

	w.Printf("I have %d %s.", 3, "frogs")
	if err := w.Close(); err != nil {
		t.Error(err)
	}

	want := hatchet.L{"message": "I have 3 frogs."}
	if len(l.Logs) != 1 {
		t.Errorf("wrong number of logs sent %d", len(l.Logs))
	} else if !reflect.DeepEqual(l.Logs[0], want) {
		t.Error("message is incorrect")
		t.Errorf("  have %+v", l.Logs[0])
		t.Errorf("  want %+v", want)
	}
}
