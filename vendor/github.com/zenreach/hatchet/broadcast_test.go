package hatchet_test

import (
	"reflect"
	"testing"

	"github.com/zenreach/hatchet"
)

func TestBroadcastLoggers(t *testing.T) {
	t.Parallel()

	mock1 := hatchet.Mock()
	mock2 := hatchet.Mock()
	logger := hatchet.Broadcast(mock1, mock2)

	logs := []hatchet.L{
		{"message": "first log"},
		{"message": "second log"},
	}
	for _, log := range logs {
		logger.Log(log)
	}

	if err := logger.Close(); err != nil {
		t.Fatalf("logger close failed: %s", err)
	}
	if !reflect.DeepEqual(mock1.Logs, logs) {
		t.Errorf("logger 1 has incorrect logs: %+v != %+v", mock1.Logs, logs)
	}
	if !reflect.DeepEqual(mock2.Logs, logs) {
		t.Errorf("logger 2 has incorrect logs: %+v != %+v", mock2.Logs, logs)
	}
}

func TestBroadcastNone(t *testing.T) {
	t.Parallel()

	logger := hatchet.Broadcast()
	logger.Log(hatchet.L{"message": "into the void"})
	if err := logger.Close(); err != nil {
		t.Errorf("logger close failed: %s", err)
	}
}
