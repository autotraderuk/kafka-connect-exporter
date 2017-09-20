package rollbar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/adler32"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/zenreach/hatchet"
)

const (
	rollbarEndpoint = "https://api.rollbar.com/api/1/item/"
	rollbarTimeout  = 10 * time.Second
)

// Rollbar defines the following common fields:
var (
	Fingerprint = "fingerprint"
)

// New creates a new logger which persists messages to Rollbar. The Rollbar
// logger treats certain log fields specially. These are:
//
// Message: The title of the message.
//
// Level: The log level. Defaults to "info" for messages or "error" for
// traces/errors.
//
// Timestamp: The time the message occurred. Either a `time.Time` or a Unix
// timestamp as int64. Defaults to `time.Now().UTC()`.
//
// Error: An error containing an optional stack trace (from
// github.com/pkg/errors).
//
// Fingerprint: An optional fingerprint under which to group messages.
//
// Hostname: The hostname to report in the server section. Defaults to
// os.Hostname.
func New(token, env string) hatchet.Logger {
	return &rollbarLogger{
		env:      env,
		token:    token,
		endpoint: rollbarEndpoint,
		client: &http.Client{
			Timeout: rollbarTimeout,
		},
	}
}

type httpClient interface {
	Post(string, string, io.Reader) (*http.Response, error)
}

type rollbarLogger struct {
	env      string
	token    string
	endpoint string
	client   httpClient
}

func (l *rollbarLogger) Log(log map[string]interface{}) {
	type errorResponse struct {
		Err     int    `json:"err"`
		Message string `json:"message"`
	}

	if len(log) == 0 {
		// discard empty logs
		return
	}

	ll := hatchet.L(log)
	payload := l.getPayload(ll)
	buf, err := json.Marshal(payload)
	if err != nil {
		hatchet.PrintFailure(ll, errors.Wrap(err, "failed to encode Rollbar payload"))
		return
	}

	resp, err := l.client.Post(l.endpoint, "application/json", bytes.NewReader(buf))
	if err != nil {
		hatchet.PrintFailure(ll, errors.Wrap(err, "failed to send Rollbar payload"))
		return
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		var err error
		errResp := &errorResponse{}
		dec := json.NewDecoder(resp.Body)
		if decodeErr := dec.Decode(errResp); decodeErr == nil {
			err = errors.Errorf("status code %d, %s", resp.StatusCode, errResp.Message)
		} else {
			err = errors.Errorf("status code %d", resp.StatusCode)
		}
		hatchet.PrintFailure(ll, err)
	}
}

func (*rollbarLogger) Close() error {
	return nil
}

func (l *rollbarLogger) getPayload(log hatchet.L) hatchet.L {
	// extract what we need to construct the rollbar message
	log = log.Copy()
	message := log.Message()
	level := log.Level()
	timestamp := getTimestamp(log)
	server := getServer(log)
	fingerprint := getFingerprint(log)
	err := log.Error()

	delete(log, hatchet.Level)
	delete(log, Fingerprint)
	delete(log, hatchet.Hostname)
	delete(log, hatchet.Time)

	var body hatchet.L
	var custom hatchet.L
	if err == nil {
		if level == "" {
			level = "info"
		}
		log["body"] = message
		delete(log, hatchet.Message)
		body = hatchet.L{
			"message": log,
		}
	} else {
		if level == "" {
			level = "error"
		}
		body = hatchet.L{
			"trace_chain": getTraceChain(err),
		}
		delete(log, hatchet.Error)
		custom = log
	}

	data := hatchet.L{
		"uuid":        fmt.Sprintf("%x", uuid.NewV4().Bytes()),
		"title":       message,
		"level":       level,
		"timestamp":   timestamp,
		"platform":    runtime.GOOS,
		"language":    "go",
		"environment": l.env,
		"body":        body,
		"notifier": hatchet.L{
			"name": "logutil",
		},
	}
	if fingerprint != "" {
		data[Fingerprint] = fingerprint
	}
	if len(server) > 0 {
		data["server"] = server
	}
	if len(custom) > 0 {
		data["custom"] = custom
	}
	return hatchet.L{
		"access_token": l.token,
		"data":         data,
	}
}

func getTimestamp(log hatchet.L) int64 {
	timestamp := log.Time()
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	return timestamp.Unix()
}

func getFingerprint(log hatchet.L) string {
	fingerprint := log[Fingerprint]
	if fingerprint == nil {
		return ""
	}
	return fmt.Sprintf("%s", fingerprint)
}

func getServer(log hatchet.L) hatchet.L {
	server := hatchet.L{}
	host, _ := log["hostname"].(string)
	if host == "" {
		host, _ = os.Hostname()
	}
	if host != "" {
		server["host"] = host
	}
	return server
}

// stolen unashamedly from https://github.com/stvp/roll/blob/master/client.go
func getErrorClass(err error) string {
	class := reflect.TypeOf(err).String()
	if class == "" {
		return "panic"
	} else if class == "*errors.errorString" || class == "*errors.fundamental" {
		checksum := adler32.Checksum([]byte(err.Error()))
		return fmt.Sprintf("{%x}", checksum)
	}
	return strings.TrimPrefix(class, "*")
}

func getTrace(err error) hatchet.L {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	frames := []hatchet.L{}
	if tracer, ok := err.(stackTracer); ok {
		stack := tracer.StackTrace()
		frames = make([]hatchet.L, len(stack))
		for n, frame := range stack {
			lineno, _ := strconv.Atoi(fmt.Sprintf("%d", frame)) // use zero on failure
			methodFmt := "%n"                                   // broken out to trick govet
			frames[n] = hatchet.L{
				"filename": fmt.Sprintf("%s", frame),
				"lineno":   lineno,
				"method":   fmt.Sprintf(methodFmt, frame),
			}
		}
	}
	return hatchet.L{
		"frames": frames,
		"exception": hatchet.L{
			"class":   getErrorClass(err),
			"message": err.Error(),
		},
	}
}

func getTraceChain(err error) []hatchet.L {
	type causer interface {
		Cause() error
	}

	chain := []hatchet.L{}
	for err != nil {
		chain = append(chain, getTrace(err))
		if errCauser, ok := err.(causer); ok {
			err = errCauser.Cause()
		} else {
			err = nil
		}
	}
	return chain
}
