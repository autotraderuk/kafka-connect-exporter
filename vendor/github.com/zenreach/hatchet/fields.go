package hatchet

// Fields creates a logger that sets the given fields on each log message. The
// logger replaces existing fields if replace is set to true.
func Fields(logger Logger, fields map[string]interface{}, replace bool) Logger {
	return &fieldLogger{
		Logger:  logger,
		Fields:  fields,
		Replace: replace,
	}
}

type fieldLogger struct {
	Logger
	Fields  map[string]interface{}
	Replace bool
}

// Log the given message. Set the provided fields on each message.
func (fl *fieldLogger) Log(in map[string]interface{}) {
	if in == nil {
		return
	}
	// don't modify the original log
	out := L(in).Copy()
	for k, v := range fl.Fields {
		if _, exists := out[k]; !exists || fl.Replace {
			out[k] = v
		}
	}
	fl.Logger.Log(out)
}
