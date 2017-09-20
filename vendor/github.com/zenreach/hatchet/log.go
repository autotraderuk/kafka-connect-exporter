package hatchet

import (
	"fmt"
	"time"
)

// Hatchet globally defines some common field names. They are used throughout
// the package and should be packages that extend Hatchet. They may be set to
// suit the developer's needs. These fields are:
var (
	// Holds the log message. Value is typically a string.
	Message = "message"

	// Holds the timestamp. Value is typically a time.Time.
	Time = "time"

	// Holds the log level. Value is typically a string.
	Level = "level"

	// Holds the error. Value is typically an error. Usually accompanied by a
	// level of warning or higher.
	Error = "error"

	// Holds the PID of the process. Typically an int.
	PID = "pid"

	// Holds the name of the process. Typically a string.
	Process = "process"

	// Holds the name of the host running hte process. Typically a string.
	Hostname = "hostname"
)

// L is used to pass structured log data to a Logger.
type L map[string]interface{}

// getString retrieves a field value and returns it as a string. An empty
// string is returned if the field is missing or nil.
func (l L) getString(field string) string {
	valueInt, ok := l[field]
	if !ok || valueInt == nil {
		return ""
	}
	return fmt.Sprintf("%s", valueInt)
}

// Message returns the value of the log's message field.
func (l L) Message() string {
	return l.getString(Message)
}

// Time returns the timestamp of the log. If the log has no timestamp the
// time.Time zero value is returned. The time is extracted from the TimeField
// log field. It is expected to be either a time.Time or an integer containing
// a Unix timestamp.
func (l L) Time() time.Time {
	switch timestamp := l[Time].(type) {
	case time.Time:
		return timestamp
	case int:
		return time.Unix(int64(timestamp), 0)
	case int32:
		return time.Unix(int64(timestamp), 0)
	case int64:
		return time.Unix(timestamp, 0)
	}
	return time.Time{}
}

// Level returns the value of the log's level field. An empty string is
// returned if it is not set.
func (l L) Level() string {
	return l.getString(Level)
}

// LevelValue returns the integer value of a log level. A 0 is returned if not
// set or unrecognized. See `LevelValue` for more detail.
func (l L) LevelValue() int {
	return LevelValue(l.Level())
}

// Error returns the value of the log's error field.
func (l L) Error() error {
	err, ok := l[Error].(error)
	if !ok {
		return nil
	}
	return err
}

// CopyTo copies fields from this log to another.
func (l L) CopyTo(cp L) {
	for k, v := range l {
		cp[k] = v
	}
}

// Copy returns a shallow copy of the log.
func (l L) Copy() L {
	cp := make(L, len(l))
	l.CopyTo(cp)
	return cp
}
