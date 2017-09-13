package goa

import (
	"context"
	"fmt"

	"github.com/goadesign/goa"
	"github.com/zenreach/hatchet"
)

// Adapt creates a Goa log adapter which logs to the provided logger.
func Adapt(logger hatchet.Logger) *Adapter {
	return &Adapter{
		Logger: logger,
	}
}

// Adapter wraps a Logger with Goa Log Adapter functionality.
type Adapter struct {
	hatchet.Logger
	context hatchet.L
}

// Info logs a message at info level.
func (a *Adapter) Info(msg string, keyvals ...interface{}) {
	a.log(hatchet.InfoLevel, msg, keyvals)
}

// Error logs a message at error level.
func (a *Adapter) Error(msg string, keyvals ...interface{}) {
	a.log(hatchet.ErrorLevel, msg, keyvals)
}

// New creates an Adapter with additional log context.
func (a *Adapter) New(keyvals ...interface{}) goa.LogAdapter {
	context := a.context.Copy()
	a.addKeyvals(context, keyvals)
	return &Adapter{
		Logger:  a.Logger,
		context: context,
	}
}

func (a *Adapter) log(level, msg string, keyvals []interface{}) {
	log := hatchet.L{
		hatchet.Message: msg,
		hatchet.Level:   level,
	}
	a.context.CopyTo(log)
	a.addKeyvals(log, keyvals)
	a.Logger.Log(log)
}

func (a *Adapter) addKeyvals(log hatchet.L, keyvals []interface{}) {
	for i := 0; i < len(keyvals); i += 2 {
		key := keyvals[i]
		var value interface{}
		if i+1 < len(keyvals) {
			value = keyvals[i+1]
		} else {
			value = goa.ErrMissingLogValue
		}
		if key == nil {
			continue
		}
		if key == "err" {
			key = hatchet.Error
		}
		log[fmt.Sprintf("%s", key)] = value
	}

}

// Extract the underlying Logger from a Goa context. If the logger is an
// Adapter then the Adapter's wrapped logger is returned. Otherwise a null
// logger is returned.
func Extract(ctx context.Context) hatchet.Logger {
	adapter, ok := goa.ContextLogger(ctx).(*Adapter)
	if !ok {
		return hatchet.Null()
	}
	return adapter.Logger
}
