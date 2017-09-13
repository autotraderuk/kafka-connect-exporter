package hatchet

// Null discards all logs.
func Null() Logger {
	var logger *nullLogger
	return logger
}

type nullLogger struct{}

func (*nullLogger) Log(map[string]interface{}) {
}

func (*nullLogger) Close() error {
	return nil
}
