// Package prometheus provides an API for gathering prometheus metrics on tasks deployed to a kafka connect cluster.
package prometheus

import (
	"net/http"

	"github.com/go-kafka/connect"
	"github.com/pkg/errors"
	prom "github.com/prometheus/client_golang/prometheus"
)

// Metrics encapsulates prom metrics for kafka connect tasks.
type Metrics struct {
	*prom.GaugeVec
	client ConnectClient
}

// ConnectClient is an abstraction for a kafka connect REST Client.
//
// NOTE: This may change, since it is currently designed around the go-kafka/connect package.
type ConnectClient interface {
	// List connectors returns a list of connector names.
	ListConnectors() ([]string, *http.Response, error)

	// GetConnectorStatus returns the status of a single connector.
	GetConnectorStatus(string) (*connect.ConnectorStatus, *http.Response, error)
}

// NewMetrics returns a new instance of prometheus metrics using the given client, and
// it will start polling the connect API at the given pollInterval.
func NewMetrics(client ConnectClient) *Metrics {
	return &Metrics{
		client: client,
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
}

// update will update all metrics for the monitored set of kafka connect configs. It
// returns an error if any underlying API calls to kafka connect fail, either by connection
// or non-2XX status code.
func (m *Metrics) Update() error {
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
			//return errors.Errorf("no tasks for connector %s", conn)
			m.With(prom.Labels{
				"connector": conn,
				"state":     "EMPTY_TASKS",
				"worker":    "-1",
			}).Inc()
		}
		m.With(prom.Labels{
			"connector": conn,
			"state":     connStatus.Connector.State,
			"worker":    "toplevel:" + connStatus.Connector.WorkerID,
		}).Inc()
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
