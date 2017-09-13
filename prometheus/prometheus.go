// Package prometheus provides an API for gathering prometheus metrics on tasks deployed to a kafka connect cluster.
package prometheus

import (
	"net/http"
	"sync"

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

// NewMetrics returns a new instance of prometheus metrics using the given client.
func NewMetrics(client ConnectClient) *Metrics {
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
	}
	return m
}

// Update will update all metrics for the monitored set of kafka connect configs. It
// returns an error if any underlying API calls to kafka connect fail, either by connection
// or non-2XX status code.
//
// NOTE: Any metrics readers should use the PauseUpdates and ResumeUpdates methods to ensure
// metrics are exposed in a consistent way.
// If Update returns an error, all exported metrics will be deleted until another successful call to Update.
func (m *Metrics) Update() error {
	m.err = nil
	conns, res, err := m.client.ListConnectors()
	if err != nil {
		m.err = errors.Wrap(err, "listing connectors")
		return m.err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		m.err = errors.Errorf("status code %d from listing connectors", res.StatusCode)
		return m.err
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
			m.err = errors.Wrapf(err, "getting status for connector %s", conn)
			return m.err
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			m.err = errors.Errorf("status code %d from getting status for connector %s", res.StatusCode, conn)
			return m.err
		}
		if len(connStatus.Tasks) == 0 {
			m.err = errors.Errorf("no tasks for connector %s", conn)
			return m.err
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

// PauseUpdates blocks metrics refreshes.
func (m *Metrics) PauseUpdates() {
	m.lock.RLock()
}

// ResumeUpdates unblocks metrics refreshes.
func (m *Metrics) ResumeUpdates() {
	m.lock.RUnlock()
}

// Err returns an error if the last call to Update returned a non-nil error,
// or nil otherwise.
func (m *Metrics) Err() error {
	return m.err
}
