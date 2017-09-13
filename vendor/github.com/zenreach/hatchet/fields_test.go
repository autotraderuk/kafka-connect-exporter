package hatchet_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/zenreach/hatchet"
)

// Table test the fields logger with replace enabled.
func TestFieldsReplace(t *testing.T) {
	t.Parallel()

	fields := hatchet.L{
		"level": "info",
		"count": 42,
	}

	tests := [][2]hatchet.L{
		{ // fields are added to empty log
			{},
			{"level": "info", "count": 42},
		},
		{ // fields are added existing log
			{"message": "hello"},
			{"message": "hello", "level": "info", "count": 42},
		},
		{ // fields are replaced
			{"message": "hi", "level": "error"},
			{"message": "hi", "level": "info", "count": 42},
		},
	}

	for n := range tests {
		test := tests[n]
		t.Run(fmt.Sprintf("test%d", n), func(t *testing.T) {
			mock := hatchet.Mock()
			logger := hatchet.Fields(mock, fields, true)
			logger.Log(test[0])
			log := mock.Last()
			if !reflect.DeepEqual(log, test[1]) {
				t.Errorf("%v != %v", log, test[1])
			}
		})
	}
}

// Table test the fields logger with replace disabled.
func TestFieldsNoReplace(t *testing.T) {
	t.Parallel()

	fields := hatchet.L{
		"level": "info",
		"count": 42,
	}

	tests := [][2]hatchet.L{
		{ // fields are added to empty log
			{},
			{"level": "info", "count": 42},
		},
		{ // new field is added
			{"message": "hello", "level": "warning"},
			{"message": "hello", "level": "warning", "count": 42},
		},
		{ // no fields are added
			{"message": "hi", "level": "error", "count": 13},
			{"message": "hi", "level": "error", "count": 13},
		},
	}

	for n := range tests {
		test := tests[n]
		t.Run(fmt.Sprintf("test%d", n), func(t *testing.T) {
			mock := hatchet.Mock()
			logger := hatchet.Fields(mock, fields, false)
			logger.Log(test[0])
			log := mock.Last()
			if !reflect.DeepEqual(log, test[1]) {
				t.Errorf("%v != %v", log, test[1])
			}
		})
	}
}

// Ensure the fields logger does not modify the input log.
func TestFieldsNoModify(t *testing.T) {
	t.Parallel()

	mock := hatchet.Mock()
	logger := hatchet.Fields(mock, hatchet.L{
		"level": "info",
		"count": 42,
	}, true)

	want := hatchet.L{"level": "error", "count": 15}
	in := hatchet.L{"level": "error", "count": 15}
	logger.Log(in)
	if !reflect.DeepEqual(in, want) {
		t.Errorf("field logger modified input: %v != %v", in, want)
	}
}
