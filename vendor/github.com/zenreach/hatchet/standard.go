package hatchet

import (
	"fmt"
	"os"
)

// Standardize creates a StandardLogger from the provided Logger.
func Standardize(l Logger) *StandardLogger {
	return &StandardLogger{Logger: l}
}

// StandardLogger extends a Logger with functions from the standard library
// logger.
type StandardLogger struct {
	Logger
}

// Write bytes to the logger. Converts `p` to a string using `Sprintf` and
// called `Output`.
func (l *StandardLogger) Write(p []byte) (n int, err error) {
	return len(p), l.Output(2, fmt.Sprintf("%s", p))
}

// Output writes the output for a logging event. It formats the message as a
// JSON object with a single field, `message`, set to `s`.
func (l *StandardLogger) Output(calldepth int, s string) error {
	l.Logger.Log(L{Message: s})
	return nil
}

// Print calls l.Output to print to the logger. Arguments are handled in the
// manner of fmt.Print.
func (l *StandardLogger) Print(v ...interface{}) {
	l.Output(2, fmt.Sprint(v...))
}

// Printf calls l.Output to print to the logger. Arguments are handled in the
// manner of fmt.Printf.
func (l *StandardLogger) Printf(f string, v ...interface{}) {
	l.Output(2, fmt.Sprintf(f, v...))
}

// Println calls l.Output to print to the logger. Arguments are handled in the
// manner of fmt.Println.
func (l *StandardLogger) Println(v ...interface{}) {
	l.Output(2, fmt.Sprintln(v...))
}

// Panic is equivalent to l.Print() followed by a call to panic().
func (l *StandardLogger) Panic(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.Output(2, msg)
	l.Logger.Close()
	panic(msg)
}

// Panicf is equivalent to l.Printf() followed by a call to panic().
func (l *StandardLogger) Panicf(f string, v ...interface{}) {
	msg := fmt.Sprintf(f, v...)
	l.Output(2, msg)
	l.Logger.Close()
	panic(msg)
}

// Panicln is equivalent to l.Println() followed by a call to panic().
func (l *StandardLogger) Panicln(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	l.Output(2, msg)
	l.Logger.Close()
	panic(msg)
}

// Fatal is equivalent to l.Print() followed by a call to os.Exit(1).
func (l *StandardLogger) Fatal(v ...interface{}) {
	l.Output(2, fmt.Sprint(v...))
	l.Close()
	os.Exit(1)
}

// Fatalf is equivalent to l.Printf() followed by a call to os.Exit(1).
func (l *StandardLogger) Fatalf(f string, v ...interface{}) {
	l.Output(2, fmt.Sprintf(f, v...))
	l.Close()
	os.Exit(1)
}

// Fatalln is equivalent to l.Println() followed by a call to os.Exit(1).
func (l *StandardLogger) Fatalln(v ...interface{}) {
	l.Output(2, fmt.Sprintln(v...))
	l.Close()
	os.Exit(1)
}
