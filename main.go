package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-kafka/connect"
	"github.com/pkg/errors"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/snahelou/kafka-connect-exporter/prometheus"
)

type config struct {
	KafkaConnectHost string `env:"KAFKA_CONNECT_HOST"`
	Port             int    `env:"PORT" envDefault:"9400"`
}

func graceful(srv *http.Server, timeout time.Duration) error {
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

func main() {
	cfg := new(config)
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}

	// set up connect api refresh
	client := connect.NewClient(cfg.KafkaConnectHost)
	metrics := prometheus.NewMetrics(client)
	prom.MustRegister(metrics)

	// expose metrics via http
	addr := fmt.Sprintf(":%d", cfg.Port)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := metrics.Update(); err != nil {
			log.Print(errors.WithStack(errors.WithMessage(err, "calling kafka connect API")))
			//w.WriteHeader(500)
			//w.Write([]byte(errors.Cause(err).Error()))
			return
		}
		promhttp.Handler().ServeHTTP(w, r)
	})
	timeout := 10 * time.Second
	if err := graceful(&http.Server{Addr: addr, Handler: handler}, timeout); err != nil {
		log.Fatal(err)
	}
}
