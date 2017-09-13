package goa_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	goalib "github.com/goadesign/goa"
	"github.com/zenreach/hatchet"
	"github.com/zenreach/hatchet/goa"
)

func TestGoaInterface(t *testing.T) {
	var logger hatchet.Logger = goa.Adapt(hatchet.Mock())
	_ = logger.(goalib.LogAdapter)
}

func TestGoaFunctions(t *testing.T) {
	type Test struct {
		fn  func()
		log hatchet.L
	}

	mock := hatchet.Mock()
	logger := goa.Adapt(mock)

	tests := []Test{
		{
			fn: func() {
				logger.Info("hello", "color", "circle")
			},
			log: hatchet.L{
				"message": "hello",
				"color":   "circle",
				"level":   "info",
			},
		},
		{
			fn: func() {
				logger.Error("oops", "value", 42)
			},
			log: hatchet.L{
				"message": "oops",
				"value":   42,
				"level":   "error",
			},
		},
		{
			fn: func() {
				logger.Info("hi", "value", 15, "cat")
			},
			log: hatchet.L{
				"message": "hi",
				"value":   15,
				"level":   "info",
				"cat":     goalib.ErrMissingLogValue,
			},
		},
		{
			fn: func() {
				logger.Error("oops", "err", errors.New("oops"))
			},
			log: hatchet.L{
				"message": "oops",
				"error":   errors.New("oops"),
				"level":   "error",
			},
		},
	}

	for n, test := range tests {
		test.fn()
		have := mock.Last()
		if !reflect.DeepEqual(have, test.log) {
			t.Errorf("test %d incorrect: %+v != %+v", n, have, test.log)
		}
	}
}

func TestGoaNew(t *testing.T) {
	mock := hatchet.Mock()
	logger := goa.Adapt(mock)

	logger.Info("hello")
	want := hatchet.L{
		"message": "hello",
		"level":   "info",
	}
	if !reflect.DeepEqual(mock.Last(), want) {
		t.Errorf("incorrect: %+v != %+v", mock.Last(), want)
	}

	newLogger := logger.New("shape", "blue")
	newLogger.Info("meow")
	want = hatchet.L{
		"message": "meow",
		"shape":   "blue",
		"level":   "info",
	}
	if !reflect.DeepEqual(mock.Last(), want) {
		t.Errorf("incorrect: %+v != %+v", mock.Last(), want)
	}
}

func TestExtractMock(t *testing.T) {
	logger := hatchet.Mock()
	ctx := context.Background()
	ctx = goalib.WithLogger(ctx, goa.Adapt(logger))

	log := hatchet.L{"message": "hello"}
	extractedLogger := goa.Extract(ctx)
	extractedLogger.Log(log)
	if !reflect.DeepEqual(logger.Last(), log) {
		t.Errorf("extracted logger incorrect: %+v != %+v", extractedLogger, logger)
	}
}

func TestExtractNull(t *testing.T) {
	ctx := context.Background()
	log := hatchet.L{"message": "hello"}
	extractedLogger := goa.Extract(ctx)
	extractedLogger.Log(log)
}
