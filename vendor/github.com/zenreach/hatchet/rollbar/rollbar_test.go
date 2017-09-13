package rollbar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/adler32"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/zenreach/hatchet"
)

type errorTracer interface {
	Error() string
	StackTrace() errors.StackTrace
}

type mockRequest struct {
	URL         string
	ContentType string
	Body        *bytes.Buffer
	Payload     map[string]interface{}
}

type mockClient struct {
	T *testing.T
	// The request passed to the client.
	Request *mockRequest
	// Set to respond with a 500 error.
	ErrorMessage string
	// Set to return an error from Post.
	PostError error
}

func (c *mockClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	type errorResponse struct {
		Err     int    `json:"err"`
		Message string `json:"message"`
	}

	buf := &bytes.Buffer{}
	io.Copy(buf, body)
	if closer, ok := body.(io.Closer); ok {
		closer.Close()
	}

	payload := map[string]interface{}{}
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		c.T.Errorf("unable to decode payload: %s", err)
		c.T.Errorf("payload: %s", buf.String())
	}

	c.Request = &mockRequest{
		URL:         url,
		ContentType: contentType,
		Body:        buf,
		Payload:     payload,
	}
	if c.PostError != nil {
		return nil, c.PostError
	}

	respBody := &bytes.Buffer{}
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Body:       ioutil.NopCloser(respBody),
	}
	if c.ErrorMessage != "" {
		errorBytes, _ := json.Marshal(errorResponse{
			Err:     1,
			Message: c.ErrorMessage,
		})
		respBody.Write(errorBytes)
		resp.Status = "500 Internal Server Error"
		resp.StatusCode = http.StatusInternalServerError
	}
	resp.ContentLength = int64(respBody.Len())
	return resp, nil
}

func UUID() string {
	return uuid.NewV4().String()
}

func TestRollbarMessage(t *testing.T) {
	env := "testing"
	token := UUID()
	client := &mockClient{T: t}
	logutilLogger := New(token, env)
	logger, ok := logutilLogger.(*rollbarLogger)
	if !ok {
		t.Fatal("New returned wrong logger type")
	}
	logger.client = client

	message := "useful information to log"
	level := "warning"
	timestamp := float64(time.Now().Unix())
	fingerprint := UUID()
	hostname := "server.example.net"
	extra := "extra info"

	log := hatchet.L{
		"message":     message,
		"level":       level,
		"time":        timestamp,
		"fingerprint": fingerprint,
		"hostname":    hostname,
		"extra":       extra,
	}

	wantData := map[string]interface{}{
		"uuid":        "",
		"timestamp":   timestamp,
		"fingerprint": fingerprint,
		"title":       message,
		"level":       level,
		"platform":    runtime.GOOS,
		"language":    "go",
		"environment": env,
		"body": map[string]interface{}{
			"message": map[string]interface{}{
				"body":  message,
				"extra": extra,
			},
		},
		"notifier": map[string]interface{}{
			"name": "logutil",
		},
		"server": map[string]interface{}{
			"host": hostname,
		},
	}
	wantPayload := map[string]interface{}{
		"access_token": token,
		"data":         wantData,
	}

	logger.Log(log)
	if client.Request.URL != rollbarEndpoint {
		t.Errorf("request URL incorrect: %s != %s", client.Request.URL, rollbarEndpoint)
	}
	if client.Request.ContentType != "application/json" {
		t.Errorf("request content type incorrect: %s != application/json", client.Request.ContentType)
	}

	if haveData, ok := client.Request.Payload["data"].(map[string]interface{}); ok && haveData != nil {
		wantData["uuid"] = haveData["uuid"]
		wantData["timestamp"] = haveData["timestamp"]

		haveTimestamp, ok := haveData["timestamp"].(float64)
		if !ok {
			t.Errorf("timestamp has wrong type: %t", haveData["timestamp"])
		}

		if haveTimestamp != timestamp {
			t.Errorf("incorrect timestamp: %f != %f", haveTimestamp, timestamp)
		}
	}

	if !reflect.DeepEqual(client.Request.Payload, wantPayload) {
		t.Error("request payload incorrect:")
		t.Errorf("  have: %+v", client.Request.Payload)
		t.Errorf("  want: %+v", wantPayload)
	}
}

func TestRollbarError(t *testing.T) {
	env := "testing"
	token := UUID()
	client := &mockClient{T: t}
	logutilLogger := New(token, env)
	logger, ok := logutilLogger.(*rollbarLogger)
	if !ok {
		t.Fatal("New returned wrong logger type")
	}
	logger.client = client

	message := "an error occurred"
	err := errors.New("oops").(errorTracer)
	errClass := fmt.Sprintf("{%x}", adler32.Checksum([]byte(err.Error())))
	trace := err.StackTrace()
	hostname, _ := os.Hostname()
	extra := "some things went wrong"

	log := hatchet.L{
		"message": message,
		"error":   err,
		"extra":   extra,
	}

	wantFrames := make([]interface{}, len(trace))
	for n, frame := range trace {
		lineno, _ := strconv.Atoi(fmt.Sprintf("%d", frame))
		methodFmt := "%n" // govet doesn't recognize %n as a printf verb
		wantFrames[n] = map[string]interface{}{
			"filename": fmt.Sprintf("%s", frame),
			"lineno":   float64(lineno),
			"method":   fmt.Sprintf(methodFmt, frame),
		}
	}
	wantData := map[string]interface{}{
		"uuid":        "",
		"timestamp":   nil,
		"title":       message,
		"level":       "error",
		"platform":    runtime.GOOS,
		"language":    "go",
		"environment": env,
		"body": map[string]interface{}{
			"trace_chain": []interface{}{
				map[string]interface{}{
					"frames": wantFrames,
					"exception": map[string]interface{}{
						"class":   errClass,
						"message": err.Error(),
					},
				},
			},
		},
		"notifier": map[string]interface{}{
			"name": "logutil",
		},
		"server": map[string]interface{}{
			"host": hostname,
		},
		"custom": map[string]interface{}{
			"message": message,
			"extra":   extra,
		},
	}
	wantPayload := map[string]interface{}{
		"access_token": token,
		"data":         wantData,
	}

	logger.Log(log)
	if client.Request.URL != rollbarEndpoint {
		t.Errorf("request URL incorrect: %s != %s", client.Request.URL, rollbarEndpoint)
	}
	if client.Request.ContentType != "application/json" {
		t.Errorf("request content type incorrect: %s != application/json", client.Request.ContentType)
	}

	if haveData, ok := client.Request.Payload["data"].(map[string]interface{}); ok && haveData != nil {
		wantData["uuid"] = haveData["uuid"]
		wantData["timestamp"] = haveData["timestamp"]

		haveTimestamp, ok := haveData["timestamp"].(float64)
		if !ok {
			t.Errorf("timestamp has wrong type: %t", haveData["timestamp"])
		}

		wantTimestamp := float64(time.Now().Unix())
		if math.Abs(haveTimestamp-wantTimestamp) > 1 {
			t.Errorf("incorrect timestamp: %f !~ %f", haveTimestamp, wantTimestamp)
		}
	}

	if !reflect.DeepEqual(client.Request.Payload, wantPayload) {
		t.Error("request payload incorrect:")
		t.Errorf("  have: %+v", client.Request.Payload)
		t.Errorf("  want: %+v", wantPayload)
	}
}
