package hatchet

import (
	"io"
	"sync"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

const (
	defaultConsoleTemplate = "{{\"2006-01-02T15:04:05Z07:00\" | .time.Format}} {{.level}}: {{.message}}{{if .error }}: {{ .error }}{{end}}\n"
)

// Console creates a logger that formats messages for logging to the console.
// The logger does not close the underlying writer when Close is called.
func Console(wr io.Writer) Logger {
	logger, err := ConsoleWithTemplate(wr, defaultConsoleTemplate)
	if err != nil {
		panic(err)
	}
	return logger
}

// ConsoleWithTemplate returns a console logger that formats its messages with
// the given template.
func ConsoleWithTemplate(wr io.Writer, templateText string) (Logger, error) {
	tpl, err := template.New("console").Option("missingkey=zero").Parse(templateText)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse template")
	}
	return &consoleLogger{
		writer:   wr,
		template: tpl,
	}, nil
}

type consoleLogger struct {
	writer   io.Writer
	template *template.Template
	mu       sync.Mutex
}

// Log the error to the console.
func (l *consoleLogger) Log(log map[string]interface{}) {
	ll := L(log).Copy()
	if _, ok := ll[Level]; !ok {
		ll[Level] = InfoLevel
	}
	timestamp := ll.Time()
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}
	ll[Time] = timestamp
	l.mu.Lock()
	if err := l.template.Execute(l.writer, ll); err != nil {
		PrintFailure(ll, err)
	}
	l.mu.Unlock()
}

func (l *consoleLogger) Close() error {
	return nil
}
