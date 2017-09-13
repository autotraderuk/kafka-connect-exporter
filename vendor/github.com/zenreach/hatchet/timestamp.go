package hatchet

import (
	"time"
)

type timestampLogger struct {
	Logger
	location *time.Location
}

// Timestamp creates a logger that adds a timestamp to each log. The timestamp
// is a time.Time holding the current UTC time. It overwrites any existing value.
func Timestamp(logger Logger) Logger {
	return TimestampInLocation(logger, nil)
}

// TimestampInLocation creates a logger that adds a timestamp to each log. The
// timestamp is a time.Time holding the current time in the provided location.
// If the location is nil UTC is used. Existing timestamps are overwritten.
func TimestampInLocation(logger Logger, location *time.Location) Logger {
	if location == nil {
		location = time.UTC
	}
	return &timestampLogger{
		Logger:   logger,
		location: location,
	}
}

// Log adds a timestamp to the log and sends it to the underlying logger.
func (logger *timestampLogger) Log(log map[string]interface{}) {
	ll := L(log).Copy()
	ll[Time] = time.Now().In(logger.location)
	logger.Logger.Log(ll)
}
