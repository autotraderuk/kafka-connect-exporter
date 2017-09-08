package prometheus

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/pkg/errors"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var promHandler = promhttp.HandlerFor(
	prom.DefaultGatherer,
	promhttp.HandlerOpts{
		ErrorHandling: promhttp.HTTPErrorOnError,
	},
)

// NewHandler returns an http handler for serving metrics.
func NewHandler(m *Metrics) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.err != nil {
			w.WriteHeader(500)
			w.Write([]byte(errors.WithStack(m.err).Error()))
			return
		}
		m.PauseUpdates()
		defer m.ResumeUpdates()
		promHandler.ServeHTTP(w, r)
	})
}

// NewServer returns a new http server listening at the given addr with
// the given handler.
func NewServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{Addr: addr, Handler: handler}
}

// Graceful starts the given server, and handles interrupt and kill signals gracefully, If a signal is encountered, the server will shutdown within the timeout given.
func Graceful(srv *http.Server, timeout time.Duration) error {
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, os.Kill)

	errC := make(chan error)
	go func() {
		defer close(errC)
		if err := srv.ListenAndServe(); err != nil {
			errC <- err
		}
	}()

	select {
	case err := <-errC:
		return err
	case <-stop:
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return srv.Shutdown(ctx)
	}
}
