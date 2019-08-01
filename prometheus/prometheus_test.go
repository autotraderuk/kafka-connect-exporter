package prometheus_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/go-kafka/connect"
	"github.com/snahelou/kafka-connect-exporter/prometheus"
)

func TestMetricsUpdateErr(t *testing.T) {
	testCases := []updateTestCase{
		{
			name: "error on list connectors",
			client: &mockConnectClient{
				listConnectorErr: true,
			},
			expectErrOnUpdate: true,
		},
		{
			name: "error on connector status",
			client: &mockConnectClient{
				connectorStatusErr: true,
				connectors:         []string{"example-connector"},
				statuses: map[string]*connect.ConnectorStatus{
					"example-connector": &connect.ConnectorStatus{
						Name: "example-connector",
						Connector: connect.ConnectorState{
							State:    "RUNNING",
							WorkerID: "example.com:8083",
						},
						Tasks: []connect.TaskState{
							{
								ID:       0,
								State:    "RUNNING",
								WorkerID: "example.com:8083",
							},
						},
					},
				},
			},
			expectErrOnUpdate: true,
		},
		{
			name:   "no connectors",
			client: &mockConnectClient{},
		},
		{
			name: "no tasks",
			client: &mockConnectClient{
				connectors: []string{"example-connector"},
				statuses: map[string]*connect.ConnectorStatus{
					"example-connector": &connect.ConnectorStatus{
						Name: "example-connector",
						Connector: connect.ConnectorState{
							State:    "RUNNING",
							WorkerID: "example.com:8083",
						},
						Tasks: []connect.TaskState{},
					},
				},
			},
			expectErrOnUpdate: false,
		},
		{
			name: "running task",
			client: &mockConnectClient{
				connectors: []string{"example-connector"},
				statuses: map[string]*connect.ConnectorStatus{
					"example-connector": &connect.ConnectorStatus{
						Name: "example-connector",
						Connector: connect.ConnectorState{
							State:    "RUNNING",
							WorkerID: "example.com:8083",
						},
						Tasks: []connect.TaskState{
							{
								ID:       0,
								State:    "RUNNING",
								WorkerID: "example.com:8083",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.assert)
	}
}

type updateTestCase struct {
	name                string
	client              *mockConnectClient
	expectErrOnUpdate   bool
	expectListCallCount int
}

func (tc updateTestCase) assert(t *testing.T) {
	// set up metrics
	metrics := prometheus.NewMetrics(tc.client)

	if err := metrics.Update(); err != nil {
		if !tc.expectErrOnUpdate {
			t.Fatal(err)
		}
	} else if tc.expectErrOnUpdate {
		t.Error("expected error on update")
	}
}

type mockConnectClient struct {
	listConnectorErr   bool
	listCallCount      int
	connectorStatusErr bool
	connectors         []string
	statuses           map[string]*connect.ConnectorStatus
}

func (c *mockConnectClient) ListConnectors() ([]string, *http.Response, error) {
	c.listCallCount++
	if c.listConnectorErr {
		return nil, nil, errors.New("error listing connectors")
	}
	return c.connectors, &http.Response{StatusCode: 200}, nil
}

func (c *mockConnectClient) GetConnectorStatus(connector string) (*connect.ConnectorStatus, *http.Response, error) {
	if c.connectorStatusErr {
		return nil, nil, errors.New("error getting connector status")
	}
	status, ok := c.statuses[connector]
	if !ok {
		return nil, &http.Response{StatusCode: 404}, nil
	}
	if status == nil {
		return nil, &http.Response{StatusCode: 500}, nil
	}
	return status, &http.Response{StatusCode: 200}, nil
}
