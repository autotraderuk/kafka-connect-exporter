package hatchet_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/zenreach/hatchet"
)

func TestFilterIsSet(t *testing.T) {
	t.Parallel()

	mock := hatchet.Mock()
	logger := hatchet.Filter(mock, hatchet.IsSet("message"))
	logs := []hatchet.L{
		{"message": "valid log"},
		{"error": "an error occurred"},
		{"message": "another valid log"},
	}
	for _, log := range logs {
		logger.Log(log)
	}

	want := []hatchet.L{
		{"message": "valid log"},
		{"message": "another valid log"},
	}
	if !reflect.DeepEqual(mock.Logs, want) {
		t.Errorf("%+v != %+v", mock.Logs, want)
	}
}

func TestFilterIsEqual(t *testing.T) {
	t.Parallel()

	mock := hatchet.Mock()
	logger := hatchet.Filter(mock, hatchet.IsEqual("level", "error"))
	logs := []hatchet.L{
		{"message": "info log", "level": "info"},
		{"message": "warn log", "level": "warn"},
		{"message": "error log", "level": "error"},
	}
	for _, log := range logs {
		logger.Log(log)
	}

	want := []hatchet.L{
		{"message": "error log", "level": "error"},
	}
	if !reflect.DeepEqual(mock.Logs, want) {
		t.Errorf("%+v != %+v", mock.Logs, want)
	}
}

func TestFilterNotIsEqual(t *testing.T) {
	t.Parallel()

	mock := hatchet.Mock()
	logger := hatchet.Filter(mock, hatchet.Not(hatchet.IsEqual("level", "error")))
	logs := []hatchet.L{
		{"message": "info log", "level": "info"},
		{"message": "warn log", "level": "warn"},
		{"message": "error log", "level": "error"},
	}
	for _, log := range logs {
		logger.Log(log)
	}

	want := []hatchet.L{
		{"message": "info log", "level": "info"},
		{"message": "warn log", "level": "warn"},
	}
	if !reflect.DeepEqual(mock.Logs, want) {
		t.Errorf("%+v != %+v", mock.Logs, want)
	}
}

func TestFilters(t *testing.T) {
	t.Parallel()

	type Test struct {
		predicate hatchet.Predicate
		log       hatchet.L
		result    bool
	}

	tests := []Test{
		// IsLevelAtLeast
		{ // log > level
			predicate: hatchet.IsLevelAtLeast(hatchet.InfoLevel),
			log: hatchet.L{
				"message": "a warning",
				"level":   hatchet.WarningLevel,
			},
			result: true,
		},
		{ // log < level
			predicate: hatchet.IsLevelAtLeast(hatchet.InfoLevel),
			log: hatchet.L{
				"message": "debug info",
				"level":   hatchet.DebugLevel,
			},
			result: false,
		},
		{ // log == level
			predicate: hatchet.IsLevelAtLeast(hatchet.WarningLevel),
			log: hatchet.L{
				"message": "a warning",
				"level":   hatchet.WarningLevel,
			},
			result: true,
		},
		{ // missing log, info > level
			predicate: hatchet.IsLevelAtLeast(hatchet.DebugLevel),
			log: hatchet.L{
				"message": "missing level",
			},
			result: true,
		},
		{ // missing log, info < level
			predicate: hatchet.IsLevelAtLeast(hatchet.WarningLevel),
			log: hatchet.L{
				"message": "missing level",
			},
			result: false,
		},
		{ // missing log, info == level
			predicate: hatchet.IsLevelAtLeast(hatchet.InfoLevel),
			log: hatchet.L{
				"message": "missing level",
			},
			result: true,
		},
		{ // invalid log, level is lowest (debug)
			predicate: hatchet.IsLevelAtLeast(hatchet.DebugLevel),
			log: hatchet.L{
				"message": "invalid level",
				"level":   "unknown",
			},
			result: false,
		},

		// IsLevelAtMost
		{ // log > level
			predicate: hatchet.IsLevelAtMost(hatchet.InfoLevel),
			log: hatchet.L{
				"message": "a warning",
				"level":   hatchet.WarningLevel,
			},
			result: false,
		},
		{ // log < level
			predicate: hatchet.IsLevelAtMost(hatchet.InfoLevel),
			log: hatchet.L{
				"message": "debug info",
				"level":   hatchet.DebugLevel,
			},
			result: true,
		},
		{ // log == level
			predicate: hatchet.IsLevelAtMost(hatchet.WarningLevel),
			log: hatchet.L{
				"message": "a warning",
				"level":   hatchet.WarningLevel,
			},
			result: true,
		},
		{ // missing log, info > level
			predicate: hatchet.IsLevelAtMost(hatchet.DebugLevel),
			log: hatchet.L{
				"message": "missing level",
			},
			result: false,
		},
		{ // missing log, info < level
			predicate: hatchet.IsLevelAtMost(hatchet.WarningLevel),
			log: hatchet.L{
				"message": "missing level",
			},
			result: true,
		},
		{ // missing log, info == level
			predicate: hatchet.IsLevelAtMost(hatchet.InfoLevel),
			log: hatchet.L{
				"message": "missing level",
			},
			result: true,
		},
		{ // invalid log, level is lowest (debug)
			predicate: hatchet.IsLevelAtMost(hatchet.DebugLevel),
			log: hatchet.L{
				"message": "invalid level",
				"level":   "unknown",
			},
			result: true,
		},
	}

	for n := range tests {
		test := tests[n]
		t.Run(fmt.Sprintf("test%d", n), func(t *testing.T) {
			have := test.predicate(test.log)
			if have != test.result {
				t.Errorf("result incorrect: %t != %t", have, test.result)
			}
		})
	}
}
