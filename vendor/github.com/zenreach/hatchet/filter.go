package hatchet

// Predicate is a function used to match logs.
type Predicate func(map[string]interface{}) bool

// Filter discards logs which do not match the predicate.
func Filter(logger Logger, predicate Predicate) Logger {
	return &filterLogger{
		Logger:    logger,
		Predicate: predicate,
	}
}

type filterLogger struct {
	Logger
	Predicate Predicate
}

func (f *filterLogger) Log(log map[string]interface{}) {
	if f.Predicate(log) {
		f.Logger.Log(log)
	}
}

// Not creates a predicate with negates the result of the given predicate.
func Not(predicate Predicate) Predicate {
	return func(log map[string]interface{}) bool {
		return !predicate(log)
	}
}

// IsSet is a predicate which returns true when a log message has the given
// field.
func IsSet(field string) Predicate {
	return func(log map[string]interface{}) bool {
		_, ok := log[field]
		return ok
	}
}

// IsEqual creates a predicate which checks the equality of a field against
// the given value. This wraps a basic Go equality check.
func IsEqual(field string, value interface{}) Predicate {
	return func(log map[string]interface{}) bool {
		have, ok := log[field]
		if !ok {
			return false
		}
		return have == value
	}
}

// IsLevelAtLeast creates a predicate that checks if a log's level is at least
// as severe as the provided level. If no level exists it is assumed to be
// InfoLevel.
func IsLevelAtLeast(level string) Predicate {
	levelValue := LevelValue(level)
	return func(log map[string]interface{}) bool {
		logLevel := L(log).Level()
		if logLevel == "" {
			logLevel = InfoLevel
		}
		return LevelValue(logLevel) >= levelValue
	}
}

// IsLevelAtMost creates a predicate that checks if a log's level is no more
// severe as the provided level. If no level exists it is assumed to be
// InfoLevel.
func IsLevelAtMost(level string) Predicate {
	levelValue := LevelValue(level)
	return func(log map[string]interface{}) bool {
		logLevel := L(log).Level()
		if logLevel == "" {
			logLevel = InfoLevel
		}
		return LevelValue(logLevel) <= levelValue
	}
}
