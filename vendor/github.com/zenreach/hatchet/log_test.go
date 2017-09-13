package hatchet_test

import (
	"reflect"
	"testing"

	"github.com/zenreach/hatchet"
)

func TestLCopy(t *testing.T) {
	t.Parallel()

	original := "original"
	log := hatchet.L{
		"message": original,
	}
	cp := log.Copy()

	if !reflect.DeepEqual(log, cp) {
		t.Errorf("log is not a copy: %+v != %+v", log, cp)
	}

	cp["message"] = "copy"
	if log["message"] != original {
		t.Errorf("original log modified: %s != %s", log["message"], original)
	}
}

func TestLCopyTo(t *testing.T) {
	t.Parallel()

	original := "original"
	log := hatchet.L{
		"message": original,
	}
	cp := hatchet.L{}
	log.CopyTo(cp)

	if !reflect.DeepEqual(log, cp) {
		t.Errorf("log is not a copy: %+v != %+v", log, cp)
	}

	cp["message"] = "copy"
	if log["message"] != original {
		t.Errorf("original log modified: %s != %s", log["message"], original)
	}
}

func TestSetField(t *testing.T) {
	// not parallel because it modifies global state
	before := hatchet.Message
	defer func() {
		hatchet.Message = before
	}()

	msg := "hello, hatchet!"
	ll := hatchet.L{
		hatchet.Message: msg,
	}

	if ll.Message() != msg {
		t.Errorf("message not set")
	}

	want := hatchet.L{
		"message": msg,
	}
	if !reflect.DeepEqual(ll, want) {
		t.Error("message set incorrectly")
		t.Errorf("  have: %+v", ll)
		t.Errorf("  want: %+v", want)
	}

	hatchet.Message = "msg"
	if ll.Message() != "" {
		t.Errorf("message field not changed")
	}

	ll[hatchet.Message] = msg
	if ll.Message() != msg {
		t.Errorf("message field not changed")
	}

	want = hatchet.L{
		"message": msg,
		"msg":     msg,
	}
	if !reflect.DeepEqual(ll, want) {
		t.Error("message set incorrectly")
		t.Errorf("  have: %+v", ll)
		t.Errorf("  want: %+v", want)
	}
}
