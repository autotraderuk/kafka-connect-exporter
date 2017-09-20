package hatchet_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/zenreach/hatchet"
)

func TestTimestamp(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	mock := hatchet.Mock()
	logger := hatchet.Timestamp(mock)
	logger.Log(hatchet.L{
		"message": "hello!",
	})

	have := mock.Last()
	want := hatchet.L{
		"message": "hello!",
		"time":    have.Time(),
	}
	if !reflect.DeepEqual(have, want) {
		t.Errorf("%v != %v", have, want)
	}
	if now.Sub(have.Time()) > time.Second {
		t.Errorf("timestamp incorrect: %s ~= %s", have.Time(), now)
	}
}

func TestTimestampOverwrite(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	mock := hatchet.Mock()
	logger := hatchet.Timestamp(mock)
	logger.Log(hatchet.L{
		"message": "hello!",
		"time":    time.Now().UTC().Add(-24 * time.Hour),
	})

	have := mock.Last()
	if now.Sub(have.Time()) > time.Second {
		t.Errorf("timestamp incorrect: %s ~= %s", have.Time(), now)
	}
}
