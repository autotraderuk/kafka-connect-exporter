package hatchet

import (
	"fmt"
)

// Hatchet defines some common log levels. These are used throughout the
// package and should be used by packages that extend Hatchet. they may be set
// to suit the developer's needs. These levels are:
var (
	CriticalLevel = "critical"
	ErrorLevel    = "error"
	WarningLevel  = "warning"
	InfoLevel     = "info"
	DebugLevel    = "debug"
)

var logLevels = map[string]int{
	CriticalLevel: 50,
	ErrorLevel:    40,
	WarningLevel:  30,
	InfoLevel:     20,
	DebugLevel:    10,
}

// LevelValue returns an integer value for the level. Higher values are higher
// in severity than lower values. A value of 0 is returned for levels which are
// unrecognized.
func LevelValue(level string) int {
	value, ok := logLevels[level]
	if !ok {
		return 0
	}
	return value
}

// Levelize creates a LeveledLogger which logs messages to the given logger.
func Levelize(logger Logger) *LeveledLogger {
	return &LeveledLogger{
		Logger: logger,
	}
}

// LeveledLogger ensures all message moving through it have a log level attached.
type LeveledLogger struct {
	Logger
	DefaultLevel string
}

// LogWithLevel logs a message at the given level. If the message already has a
// level it is overwitten.
func (l *LeveledLogger) LogWithLevel(level string, log map[string]interface{}) {
	if len(log) == 0 {
		return
	}
	log[Level] = level
	l.Logger.Log(log)
}

// Log a message. Set the log level if it is not already set. The level is set
// to ErrorLevel if the log contains an error or InfoLevel otherwise.
func (l *LeveledLogger) Log(log map[string]interface{}) {
	if len(log) == 0 {
		return
	}
	level, ok := log[Level]
	if !ok || level == nil {
		if L(log).Error() == nil {
			log[Level] = InfoLevel
		} else {
			log[Level] = ErrorLevel
		}
	}
	l.Logger.Log(log)
}

// Print works like `fmt.Print` but sends the message to the logger.
func (l *LeveledLogger) Print(level string, a ...interface{}) {
	l.LogWithLevel(level, L{
		Message: fmt.Sprint(a...),
	})
}

// Printf works like `fmt.Printf` but sends the message to the logger.
func (l *LeveledLogger) Printf(level, format string, a ...interface{}) {
	l.LogWithLevel(level, L{
		Message: fmt.Sprintf(format, a...),
	})
}

// Critical calls `Print` with `CriticalLevel`.
func (l *LeveledLogger) Critical(a ...interface{}) {
	l.Print(CriticalLevel, a...)
}

// Criticalf calls `Printf` with `CriticalLevel`.
func (l *LeveledLogger) Criticalf(format string, a ...interface{}) {
	l.Printf(CriticalLevel, format, a...)
}

// Error calls `Print` with `ErrorLevel`.
func (l *LeveledLogger) Error(a ...interface{}) {
	l.Print(ErrorLevel, a...)
}

// Errorf calls `Printf` with `ErrorLevel`.
func (l *LeveledLogger) Errorf(format string, a ...interface{}) {
	l.Printf(ErrorLevel, format, a...)
}

// Warning calls `Print` with `WarningLevel`.
func (l *LeveledLogger) Warning(a ...interface{}) {
	l.Print(WarningLevel, a...)
}

// Warningf calls `Printf` with `WarningLevel`.
func (l *LeveledLogger) Warningf(format string, a ...interface{}) {
	l.Printf(WarningLevel, format, a...)
}

// Info calls `Print` with `InfoLevel`.
func (l *LeveledLogger) Info(a ...interface{}) {
	l.Print(InfoLevel, a...)
}

// Infof calls `Printf` with `InfoLevel`.
func (l *LeveledLogger) Infof(format string, a ...interface{}) {
	l.Printf(InfoLevel, format, a...)
}

// Debug calls `Print` with `DebugLevel`.
func (l *LeveledLogger) Debug(a ...interface{}) {
	l.Print(DebugLevel, a...)
}

// Debugf calls `Printf` with `DebugLevel`.
func (l *LeveledLogger) Debugf(format string, a ...interface{}) {
	l.Printf(DebugLevel, format, a...)
}
