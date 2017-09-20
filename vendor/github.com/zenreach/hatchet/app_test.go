package hatchet_test

import (
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/zenreach/hatchet"
)

// Test FlagFields with all flags.
func TestFlagFieldsAll(t *testing.T) {
	t.Parallel()

	mock := hatchet.Mock()
	logger := hatchet.AppInfo(mock)
	logger.Log(hatchet.L{
		"message": "hello!",
	})

	process := path.Base(os.Args[0])
	pid := os.Getpid()
	hostname, _ := os.Hostname()

	want := hatchet.L{
		"message":  "hello!",
		"pid":      pid,
		"process":  process,
		"hostname": hostname,
	}
	have := mock.Last()
	if !reflect.DeepEqual(have, want) {
		t.Errorf("%v != %v", have, want)
	}
}
