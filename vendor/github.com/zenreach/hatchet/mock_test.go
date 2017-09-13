package hatchet_test

import (
	"reflect"
	"testing"

	"github.com/zenreach/hatchet"
)

func TestMockEmpty(t *testing.T) {
	t.Parallel()

	mock := hatchet.Mock()
	if mock.First() != nil {
		t.Error("first returns non-nil on empty mock")
	}
	if mock.Last() != nil {
		t.Error("last returns non-nil on empty mock")
	}
	if mock.Close() != nil {
		t.Error("close returns non-nil on empty mock")
	}
}

func TestMockLogs(t *testing.T) {
	t.Parallel()

	mock := hatchet.Mock()
	logs := []hatchet.L{
		{
			"message": "first",
		},
		{
			"message": "second",
		},
		{
			"message": "third",
		},
	}
	for _, log := range logs {
		mock.Log(log)
	}
	if !reflect.DeepEqual(mock.Logs, logs) {
		t.Errorf("mock does not contain logs: %+v != %+v", mock.Logs, logs)
	}
	first := mock.First()
	if !reflect.DeepEqual(first, logs[0]) {
		t.Errorf("first returned incorrect log: %+v != %+v", first, logs[0])
	}
	last := mock.Last()
	if !reflect.DeepEqual(last, logs[2]) {
		t.Errorf("last returned incorrect log: %+v != %+v", last, logs[2])
	}
}
