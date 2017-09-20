package hatchet_test

import (
	"testing"

	"github.com/zenreach/hatchet"
)

func TestNull(t *testing.T) {
	t.Parallel()

	logger := hatchet.Null()
	logger.Log(hatchet.L{"message": "into the void"})
	if err := logger.Close(); err != nil {
		t.Errorf("close returned error: %s", err)
	}
}
