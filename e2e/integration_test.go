// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package e2e

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_test/e2e"
	actValidate "github.com/steadybit/action-kit/go/action_kit_test/validate"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	disValidate "github.com/steadybit/discovery-kit/go/discovery_kit_test/validate"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-prometheus/v2/extinstance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestWithMinikube(t *testing.T) {
	extlogging.InitZeroLog()

	extFactory := e2e.HelmExtensionFactory{
		Name: "extension-prometheus",
		Port: 8087,
		ExtraArgs: func(m *e2e.Minikube) []string {
			return []string{
				"--set", "logging.level=debug",
				"--set", "extraEnv[0].name=STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_0_NAME",
				"--set", "extraEnv[0].value=Test_Prometheus",
				"--set", "extraEnv[1].name=STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_0_ORIGIN",
				"--set", "extraEnv[1].value=http://host.minikube.internal",
			}
		},
	}

	e2e.WithDefaultMinikube(t, &extFactory, []e2e.WithMinikubeTestCase{
		{
			Name: "validate discovery",
			Test: validateDiscovery,
		},
		{
			Name: "target discovery",
			Test: testDiscovery,
		},
		{
			Name: "validate Actions",
			Test: validateActions,
		},
	})
}

func validateDiscovery(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	assert.NoError(t, disValidate.ValidateEndpointReferences("/", e.Client))
}

func validateActions(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	assert.NoError(t, actValidate.ValidateEndpointReferences("/", e.Client))
}

func testDiscovery(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	log.Info().Msg("Starting testDiscovery")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	target, err := e2e.PollForTarget(ctx, e, extinstance.PrometheusInstanceTargetId, func(target discovery_kit_api.Target) bool {
		log.Info().Msgf("Checking target: %v", target)
		return e2e.HasAttribute(target, "prometheus.instance.name", "Test_Prometheus") && e2e.HasAttribute(target, "prometheus.instance.url", "http://host.minikube.internal")
	})

	require.NoError(t, err)
	assert.Equal(t, target.TargetType, extinstance.PrometheusInstanceTargetId)
	assert.Contains(t, target.Attributes, "prometheus.instance.name")
	assert.Contains(t, target.Attributes, "prometheus.instance.url")
}
