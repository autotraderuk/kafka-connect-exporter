package prometheus_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kafka/connect"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/zenreach/kafka-connect-exporter/prometheus"
)

func TestServer(t *testing.T) {
	testCases := []srvTestCase{
		{
			name: "error on list connectors",
			client: &mockConnectClient{
				listConnectorErr: true,
			},
			expectErrOnUpdate: true,
			expectStatus:      500,
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
			expectStatus:      500,
		},
		{
			name:         "no connectors",
			client:       &mockConnectClient{},
			expectStatus: 200,
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
			expectErrOnUpdate: true,
			expectStatus:      500,
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
			expectStatus: 200,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.assert)
	}
}

type srvTestCase struct {
	name              string
	client            *mockConnectClient
	expectErrOnUpdate bool
	expectStatus      int
}

func (tc srvTestCase) assert(t *testing.T) {
	// set up metrics
	metrics := prometheus.NewMetrics(prom.NewRegistry(), tc.client)
	h := prometheus.NewHandler(metrics)

	// start a test server
	ts := httptest.NewServer(h)

	if err := metrics.Update(); err != nil {
		if !tc.expectErrOnUpdate {
			t.Fatal(err)
		}
	} else if tc.expectErrOnUpdate {
		t.Error("expected error on update")
	}

	// call the metrics endpoint
	res, err := http.Get(fmt.Sprintf("%s/metrics", ts.URL))
	if err != nil {
		t.Fatal(err)
	}

	// assert
	if res.StatusCode != tc.expectStatus {
		t.Errorf("Unexpected status code.\nExpected: %d\nActual: %d\n", tc.expectStatus, res.StatusCode)
	}
}

type mockConnectClient struct {
	listConnectorErr   bool
	connectorStatusErr bool
	connectors         []string
	statuses           map[string]*connect.ConnectorStatus
}

func (c *mockConnectClient) ListConnectors() ([]string, *http.Response, error) {
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
