// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH
package extmetric

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	dcontainer "github.com/docker/docker/api/types/container"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-prometheus/v2/config"
	"github.com/steadybit/extension-prometheus/v2/extinstance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestQueryMetrics(t *testing.T) {
	// Given
	container, err := setupTestContainers(context.Background())
	require.Nil(t, err)

	instance := extinstance.Instance{Name: "test-prom", BaseUrl: container.baseUrl}
	extinstance.Instances = []extinstance.Instance{instance}

	require.Eventually(t, func() bool {
		result, err := getTestMetric(instance)
		return err == nil && len(*result.Metrics) > 0
	}, time.Minute, time.Millisecond*200)

	// When
	result, exterr := getTestMetric(instance)
	require.Nil(t, exterr)

	assert.Len(t, *result.Metrics, 1)

	metric := (*result.Metrics)[0]
	assert.NotNil(t, metric.Timestamp)
	assert.Nil(t, metric.Name)
	assert.Equal(t, float64(1), metric.Value)
	assert.Equal(t, "up", metric.Metric["__name__"])
	assert.Equal(t, "localhost:9090", metric.Metric["instance"])
	assert.Equal(t, "prometheus", metric.Metric["job"])
}

func TestQueryRetries(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		retries int
	}{
		{
			name:    "NoRetries",
			retries: 0,
			wantErr: true,
		},
		{
			name:    "WithRetries",
			wantErr: false,
			retries: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flakyPrometheusURL := setupFlakyInstance(t)
			// Cleanup so we don't impact other tests, regardless of the order.
			prevRetries := config.Config.Retries
			t.Cleanup(func() {
				config.Config.Retries = prevRetries
			})
			config.Config.Retries = tt.retries
			instance := extinstance.Instance{Name: "flaky-prom", BaseUrl: flakyPrometheusURL}
			extinstance.Instances = []extinstance.Instance{instance}

			_, err := getTestMetric(instance)

			if tt.wantErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
		})
	}
}

func getTestMetric(instance extinstance.Instance) (*action_kit_api.QueryMetricsResult, error) {
	timestamp := time.Now()
	action := NewMetricCheckAction().(action_kit_sdk.ActionWithMetricQuery[MetricCheckState])

	return action.QueryMetrics(context.Background(), action_kit_api.QueryMetricsRequestBody{
		Target: extutil.Ptr(action_kit_api.Target{
			Name: instance.Name,
		}),
		Timestamp: timestamp,
		Config: map[string]interface{}{
			"query": "up",
		},
	})

}

type testContainer struct {
	container *testcontainers.Container
	baseUrl   string
}

func setupTestContainers(ctx context.Context) (*testContainer, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	containerReq := testcontainers.ContainerRequest{
		Image:        "prom/prometheus:v2.38.0",
		Name:         "test-prometheus",
		ExposedPorts: []string{"9090/tcp"},
		WaitingFor:   wait.ForHTTP("/-/ready").WithPort("9090"),
		HostConfigModifier: func(hostConfig *dcontainer.HostConfig) {
			hostConfig.Binds = append(hostConfig.Binds, path.Join(wd, "prometheus-test-config")+":/etc/prometheus")
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: containerReq,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "9090")
	if err != nil {
		return nil, err
	}

	origin := fmt.Sprintf("http://%s:%s", ip, port.Port())

	return &testContainer{
		container: &container,
		baseUrl:   origin,
	}, nil
}

func setupFlakyInstance(t *testing.T) (url string) {
	t.Helper()

	count := 0
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count += 1
			// Fail half the requests, starting with the first one.
			if count%2 == 1 {
				http.Error(w, "Temporary error", http.StatusInternalServerError)
				return
			}
			t.Log("Successful request after", count, "attempts")

			// Return a proper Prometheus API JSON response for the "up" metric
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{
  "status": "success",
  "data": {
    "resultType": "matrix",
    "result": [
      {
        "metric": {
          "__name__": "up",
          "instance": "localhost:9090",
          "job": "prometheus"
        },
        "values": [
          [1675956970.123, "1"]
        ]
      }
    ]
  }
}`)); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
			}
		}),
	)
	t.Cleanup(server.Close)

	return server.URL
}
