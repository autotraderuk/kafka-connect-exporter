// package prometheus provides a prometheus backend for task monitoring. It contains a Metrics type for synchronizing connect API requests for updating metrics, and reads (typically via an http handler). Also included, is an http server for exposing the metrics.
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
	client ConnectClient
	lock   *sync.RWMutex
	tasks  *prom.GaugeVec
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
func NewMetrics(reg prom.Registerer, client ConnectClient) *Metrics {
	m := &Metrics{
		client: client,
		lock:   new(sync.RWMutex),
		tasks: prom.NewGaugeVec(
			prom.GaugeOpts{
				Namespace: "kafka",
				Subsystem: "connect",
				Name:      "tasks",
				Help:      "deployed tasks",
			},
			[]string{"connector", "state", "worker"},
		),
	}
	reg.MustRegister(m.tasks)
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

	m.tasks.Reset()

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
			m.tasks.With(prom.Labels{
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
