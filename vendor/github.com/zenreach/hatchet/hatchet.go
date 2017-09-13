package hatchet

import (
	"fmt"
	"os"
)

// Logger implements field-based logging.
type Logger interface {
	// Log send a structured message to the logger.
	Log(map[string]interface{})

	// Close the logger. Cleans up resources associated with the logger
	// including any loggers encapsulated by this logger.
	Close() error
}

// PrintFailure is used to indicate a failure which causes a log to be
// discarded. The error and discarded log are printed to stderr.
func PrintFailure(log map[string]interface{}, err error) {
	fmt.Fprintf(os.Stderr, "logger error: %+v\n", err)
	fmt.Fprintf(os.Stderr, "discard log: %+v\n", log)
}
