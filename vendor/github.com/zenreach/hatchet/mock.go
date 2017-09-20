package hatchet

// Mock returns a logger suitable for use in mocks.
func Mock() *MockLogger {
	return &MockLogger{}
}

// MockLogger stores logs for testing.
type MockLogger struct {
	Logs []L
}

// First returns the first log sent through the mock or nil of none have been
// sent.
func (l *MockLogger) First() L {
	if len(l.Logs) > 0 {
		return l.Logs[0]
	}
	return nil
}

// Last returns the first log sent through the mock or nil of none have been
// sent.
func (l *MockLogger) Last() L {
	if len(l.Logs) > 0 {
		return l.Logs[len(l.Logs)-1]
	}
	return nil
}

// Log appends a log to the mock's array of logs.
func (l *MockLogger) Log(log map[string]interface{}) {
	l.Logs = append(l.Logs, L(log))
}

// Close does nothing and returns nil.
func (*MockLogger) Close() error {
	return nil
}
