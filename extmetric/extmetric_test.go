// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH
package extmetric

import (
	"context"
	"fmt"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-prometheus/extinstance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"path"
	"testing"
	"time"
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
		Mounts: testcontainers.ContainerMounts{
			testcontainers.ContainerMount{
				Source: testcontainers.GenericBindMountSource{
					HostPath: path.Join(wd, "prometheus-test-config"),
				},
				Target:   "/etc/prometheus",
				ReadOnly: true,
			},
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
