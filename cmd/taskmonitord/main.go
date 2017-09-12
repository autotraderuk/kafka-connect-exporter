package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-kafka/connect"
	prom "github.com/prometheus/client_golang/prometheus"
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
	viper.SetDefault("prometheus.port", "9400")

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
	viper.BindEnv("prometheus.port")

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

func main() {
	// set up logging
	logger := hatchet.JSON(os.Stderr)
	loggers := make([]hatchet.Logger, 0, 2)
	var leLogger hatchet.Logger
	var err error
	leToken := viper.GetString("logging.logentries.token")
	if leToken != "" {
		if leLogger, err = logentries.New(leToken); err != nil {
			log.Fatal(err)
		}
		loggers = append(loggers, leLogger)
	}
	rbToken := viper.GetString("logging.rollbar.token")
	rbEnv := viper.GetString("logging.rollbar.env")
	loggers = append(loggers, rollbar.New(rbToken, rbEnv))
	if len(loggers) > 0 {
		hatchet.Broadcast(loggers...)
	}

	// set up connect api refresh
	connectHost := viper.GetString("connect.host")
	if connectHost == "" {
		log.Fatal("no configured connect host")
	}
	client := connect.NewClient(connectHost)
	metrics := prometheus.NewMetrics(prom.DefaultRegisterer, client)
	ival := viper.GetInt64("connect.poll-interval")
	go func() {
		for {
			<-time.After(time.Duration(ival) * time.Second)
			if err := metrics.Update(); err != nil {
				logger.Log(hatchet.L{
					"message": "updating metrics",
					"error":   err,
				})
			}
		}
	}()

	// expose metrics via http
	promPort := viper.GetString("prometheus.port")
	if promPort == "" {
		log.Fatal("no configured prometheus port")
	}
	addr := fmt.Sprintf(":%s", promPort)
	handler := prometheus.NewHandler(metrics)
	srv := prometheus.NewServer(addr, handler)
	timeout := time.Duration(10) * time.Second
	if err := prometheus.Graceful(srv, timeout); err != nil {
		log.Fatal(err)
	}
}
