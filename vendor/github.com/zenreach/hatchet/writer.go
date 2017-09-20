package hatchet

import (
	"bytes"
)

const writerBufferSize = 256

// Writer is an io.Writer which sends each written line to the underlying
// Logger.
//
// By default the emitted log contains the line in the "message" field. This
// may be changed by setting Field to a non-empty value.
//
// Writes are buffered in order to allow a line to be written over multiple
// writes. By default the maximum size of this buffer is 256 bytes. This may be
// changed by setting BufferSize. When the buffer reaches its max size without
// seeing a newline its contents are emitted to a log.
//
// The buffer may be flushed to a log by calling `Flush`. `Flush` is called by
// `Close` prior to closing the underlying logger.
type Writer struct {
	Logger
	Field      string
	BufferSize int
	buffer     bytes.Buffer
}

// NewWriter creates a log writer with default field and buffer size.
func NewWriter(logger Logger) *Writer {
	return &Writer{Logger: logger}
}

// Write appends data to the buffer. The buffer is flushed when a newline is
// written or when the buffer's capacity is reached.
func (w *Writer) Write(b []byte) (int, error) {
	written := 0

	for len(b) > 0 {
		var line []byte
		index := bytes.IndexRune(b, '\n')
		if index == -1 {
			line = b
			b = b[:0]
		} else {
			line = b[:index]
			b = b[index+1:]
		}

		for {
			capacity := w.BufferSize - w.buffer.Len()
			if capacity > len(line) {
				capacity = len(line)
			}
			n, err := w.buffer.Write(line[:capacity])
			written += n
			if err != nil {
				return written, err
			}
			line = line[capacity:]
			if len(line) == 0 {
				break
			}
			w.Flush()
		}

		if index != -1 {
			written++
			w.Flush()
		}
	}
	return written, nil
}

// Flush emits the current contents of the buffer to a log and clears the
// buffer.
func (w *Writer) Flush() {
	field := w.Field
	if field == "" {
		field = Message
	}
	w.Logger.Log(L{
		field: w.buffer.String(),
	})
	w.buffer.Reset()
}

// Close flushes the buffer and closes the underlying logger.
func (w *Writer) Close() error {
	w.Flush()
	return w.Logger.Close()
}
