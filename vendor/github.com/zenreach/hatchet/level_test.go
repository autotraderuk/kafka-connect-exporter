package hatchet_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/zenreach/hatchet"
)

func TestLeveledLoggerDefault(t *testing.T) {
	t.Parallel()

	tests := [][2]hatchet.L{
		{
			{"message": "hello, world"},
			{"message": "hello, world", "level": hatchet.InfoLevel},
		},
		{
			{"message": "oops", "error": errors.New("oops")},
			{"message": "oops", "error": errors.New("oops"), "level": hatchet.ErrorLevel},
		},
		{
			{"message": "hello, world", "level": "not a level"},
			{"message": "hello, world", "level": "not a level"},
		},
	}

	for n := range tests {
		test := tests[n]
		t.Run(fmt.Sprintf("test%d", n), func(t *testing.T) {
			mock := hatchet.Mock()
			logger := hatchet.Levelize(mock)
			logger.Log(test[0])
			have := mock.Last()
			want := test[1]
			if !reflect.DeepEqual(have, want) {
				t.Errorf("incorrect: %+v != %+v", have, want)
			}
		})
	}
}

func TestLeveledLoggerFunctions(t *testing.T) {
	t.Parallel()

	type Test struct {
		fn  func(logger *hatchet.LeveledLogger)
		log hatchet.L
	}
	tests := []Test{
		{
			fn: func(logger *hatchet.LeveledLogger) {
				logger.LogWithLevel(hatchet.InfoLevel, hatchet.L{
					"message":  "hello",
					"followup": "it's me",
				})
			},
			log: hatchet.L{
				"message":  "hello",
				"followup": "it's me",
				"level":    "info",
			},
		},
		{
			fn: func(logger *hatchet.LeveledLogger) {
				logger.Print(hatchet.InfoLevel, "hello")
			},
			log: hatchet.L{
				"message": "hello",
				"level":   "info",
			},
		},
		{
			fn: func(logger *hatchet.LeveledLogger) {
				logger.Printf(hatchet.WarningLevel, "something %s", "fishy")
			},
			log: hatchet.L{
				"message": "something fishy",
				"level":   "warning",
			},
		},
		{
			fn: func(logger *hatchet.LeveledLogger) {
				logger.Critical("halt!")
			},
			log: hatchet.L{
				"message": "halt!",
				"level":   "critical",
			},
		},
		{
			fn: func(logger *hatchet.LeveledLogger) {
				logger.Criticalf("system error, code %d", 1234)
			},
			log: hatchet.L{
				"message": "system error, code 1234",
				"level":   "critical",
			},
		},
		{
			fn: func(logger *hatchet.LeveledLogger) {
				logger.Error("oops")
			},
			log: hatchet.L{
				"message": "oops",
				"level":   "error",
			},
		},
		{
			fn: func(logger *hatchet.LeveledLogger) {
				logger.Errorf("a more specific oops: %s", "red")
			},
			log: hatchet.L{
				"message": "a more specific oops: red",
				"level":   "error",
			},
		},
		{
			fn: func(logger *hatchet.LeveledLogger) {
				logger.Warning("danger")
			},
			log: hatchet.L{
				"message": "danger",
				"level":   "warning",
			},
		},
		{
			fn: func(logger *hatchet.LeveledLogger) {
				logger.Warningf("danger %s", "Will Robinson")
			},
			log: hatchet.L{
				"message": "danger Will Robinson",
				"level":   "warning",
			},
		},
		{
			fn: func(logger *hatchet.LeveledLogger) {
				logger.Info("purely informational")
			},
			log: hatchet.L{
				"message": "purely informational",
				"level":   "info",
			},
		},
		{
			fn: func(logger *hatchet.LeveledLogger) {
				logger.Infof("fyi: %s", "this is a message")
			},
			log: hatchet.L{
				"message": "fyi: this is a message",
				"level":   "info",
			},
		},
		{
			fn: func(logger *hatchet.LeveledLogger) {
				logger.Debug("six legged creature removal")
			},
			log: hatchet.L{
				"message": "six legged creature removal",
				"level":   "debug",
			},
		},
		{
			fn: func(logger *hatchet.LeveledLogger) {
				logger.Debugf("bug zapper %d", 9000)
			},
			log: hatchet.L{
				"message": "bug zapper 9000",
				"level":   "debug",
			},
		},
	}

	for n := range tests {
		test := tests[n]
		t.Run(fmt.Sprintf("test%d", n), func(t *testing.T) {
			mock := hatchet.Mock()
			logger := hatchet.Levelize(mock)

			test.fn(logger)
			have := mock.Last()
			if !reflect.DeepEqual(have, test.log) {
				t.Errorf("incorrect: %+v != %+v", have, test.log)
			}
		})
	}
}
