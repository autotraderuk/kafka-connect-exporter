// Package prometheus provides an API for gathering prometheus metrics on tasks deployed to a kafka connect cluster.
package prometheus

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-kafka/connect"
	"github.com/pkg/errors"
	prom "github.com/prometheus/client_golang/prometheus"
)

// Metrics encapsulates prom metrics for kafka connect tasks.
type Metrics struct {
	*prom.GaugeVec
	client ConnectClient
	lock   *sync.RWMutex
	err    error
	close  chan struct{}
}

// ConnectClient is an abstraction for a kafka connect REST Client.
//
// NOTE: This is bound to change, since it is currently designed around the go-kafka/connect package.
type ConnectClient interface {
	// List connectors returns a list of connector names.
	ListConnectors() ([]string, *http.Response, error)

	// GetConnectorStatus returns the status of a single connector.
	GetConnectorStatus(string) (*connect.ConnectorStatus, *http.Response, error)
}

// Clock is a type for abstracting the time.After method.
// Like ConnectClient, this is mostly used for testing purposes.
type Clock interface {
	// After returns a channel that sends the current time after d.
	After(d time.Duration) <-chan time.Time
}

// NewMetrics returns a new instance of prometheus metrics using the given client, and
// it will start polling the connect API at the given pollInterval.
func NewMetrics(client ConnectClient, clock Clock, pollInterval time.Duration) *Metrics {
	m := &Metrics{
		client: client,
		lock:   new(sync.RWMutex),
		GaugeVec: prom.NewGaugeVec(
			prom.GaugeOpts{
				Namespace: "kafka",
				Subsystem: "connect",
				Name:      "tasks",
				Help:      "deployed tasks",
			},
			[]string{"connector", "state", "worker"},
		),
		close: make(chan struct{}),
	}
	go func() {
		for {
			select {
			case <-clock.After(pollInterval):
				m.err = m.update()
			case <-m.close:
				return
			}
		}
	}()
	return m
}

// update will update all metrics for the monitored set of kafka connect configs. It
// returns an error if any underlying API calls to kafka connect fail, either by connection
// or non-2XX status code.
func (m *Metrics) update() error {
	conns, res, err := m.client.ListConnectors()
	if err != nil {
		return errors.Wrap(err, "listing connectors")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return errors.Errorf("status code %d from listing connectors", res.StatusCode)
	}

	if len(conns) == 0 {
		return nil
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	m.Reset()

	for _, conn := range conns {
		connStatus, res, err := m.client.GetConnectorStatus(conn)
		if err != nil {
			return errors.Wrapf(err, "getting status for connector %s", conn)
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return errors.Errorf("status code %d from getting status for connector %s", res.StatusCode, conn)
		}
		if len(connStatus.Tasks) == 0 {
			return errors.Errorf("no tasks for connector %s", conn)
		}
		for _, tStatus := range connStatus.Tasks {
			m.With(prom.Labels{
				"connector": conn,
				"state":     tStatus.State,
				"worker":    tStatus.WorkerID,
			}).Inc()
		}
	}
	return nil
}

// Err returns a non-nil error if the last call to the connect API
// resulted in an error or a non-2XX status code.
func (m *Metrics) Err() error {
	return m.err
}

// Collect overrides the Collect method to ensure metrics
// are only collected after any updates have finished.
func (m *Metrics) Collect(ch chan<- prom.Metric) {
	m.lock.RLock()
	m.GaugeVec.Collect(ch)
	m.lock.RUnlock()
}

// Close stops metrics from polling the connect API, and the prometheus metrics
// will no longer update. It always returns nil.
func (m *Metrics) Close() error {
	m.close <- struct{}{}
	return nil
}
