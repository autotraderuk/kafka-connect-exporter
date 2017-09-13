package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-kafka/connect"
	"github.com/pkg/errors"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"github.com/zenreach/hatchet"
	"github.com/zenreach/hatchet/logentries"
	"github.com/zenreach/hatchet/rollbar"
	"github.com/zenreach/kafka-connect-exporter/prometheus"
)

func init() {
	// defaults
	viper.SetDefault("logging.logentries.token", "")
	viper.SetDefault("logging.rollbar.token", "")
	viper.SetDefault("logging.rollbar.env", "")
	viper.SetDefault("config.file.path", "")
	viper.SetDefault("config.consul.url", "")
	viper.SetDefault("config.consul.path", "")
	viper.SetDefault("connect.host", "")
	viper.SetDefault("connect.poll-interval", "10")
	viper.SetDefault("prometheus.listen", ":9400")

	// env vars
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.BindEnv("logging.logentries.token")
	viper.BindEnv("logging.rollbar.token")
	viper.BindEnv("logging.rollbar.env")
	viper.BindEnv("config.file.name")
	viper.BindEnv("config.file.path")
	viper.BindEnv("config.consul.url")
	viper.BindEnv("config.consul.path")
	viper.BindEnv("connect.host")
	viper.BindEnv("connect.poll-interval")
	viper.BindEnv("prometheus.listen")

	// config file
	fPath := viper.GetString("config.file.path")
	if fPath != "" {
		viper.SetConfigFile(fPath)
		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}
	}

	// consul
	remURL := viper.GetString("config.consul.url")
	remPath := viper.GetString("config.consul.path")
	if remURL != "" && remPath != "" {
		viper.AddRemoteProvider("consul", remURL, remPath)
		if err := viper.ReadRemoteConfig(); err != nil {
			panic(err)
		}
	}
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

type clock struct{}

func (c *clock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func main() {
	// set up logging
	logger := hatchet.Standardize(hatchet.JSON(os.Stderr))
	loggers := make([]hatchet.Logger, 0, 2)
	var leLogger hatchet.Logger
	var err error
	leToken := viper.GetString("logging.logentries.token")
	if leToken != "" {
		if leLogger, err = logentries.New(leToken); err != nil {
			logger.Fatal(err)
		}
		loggers = append(loggers, leLogger)
	}
	rbToken := viper.GetString("logging.rollbar.token")
	rbEnv := viper.GetString("logging.rollbar.env")
	loggers = append(loggers, rollbar.New(rbToken, rbEnv))
	if len(loggers) > 0 {
		logger = hatchet.Standardize(hatchet.Broadcast(loggers...))
	}

	// set up connect api refresh
	connectHost := viper.GetString("connect.host")
	if connectHost == "" {
		logger.Fatal("no configured connect host")
	}
	client := connect.NewClient(connectHost)
	ival := viper.GetInt64("connect.poll-interval")
	metrics := prometheus.NewMetrics(client, new(clock), time.Duration(ival)*time.Second)
	defer metrics.Close()
	prom.MustRegister(metrics)

	// expose metrics via http
	addr := viper.GetString("prometheus.listen")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := metrics.Err(); err != nil {
			logger.Log(hatchet.L{
				"message": "calling kafka connect API",
				"error":   err,
			})
			w.WriteHeader(500)
			w.Write([]byte(errors.Cause(err).Error()))
			return
		}
		promhttp.Handler().ServeHTTP(w, r)
	})
	timeout := time.Duration(10) * time.Second
	if err := graceful(&http.Server{Addr: addr, Handler: handler}, timeout); err != nil {
		logger.Fatal(err)
	}
}
