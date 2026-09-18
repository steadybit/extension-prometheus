// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package main

import (
	_ "github.com/KimMachineGun/automemlimit" // By default, it sets `GOMEMLIMIT` to 90% of cgroup's memory limit.
	"github.com/rs/zerolog"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_sdk"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthealth"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-kit/extotel"
	"github.com/steadybit/extension-kit/extruntime"
	"github.com/steadybit/extension-kit/extsignals"
	"github.com/steadybit/extension-prometheus/v2/config"
	"github.com/steadybit/extension-prometheus/v2/extinstance"
	"github.com/steadybit/extension-prometheus/v2/extmetric"
)

func main() {
	extlogging.InitZeroLog()

	// Export OpenTelemetry traces when an OTLP endpoint is configured, so an
	// operator debugging a slow or timing-out action can see what happened inside
	// this extension. Off, and free, until OTEL_EXPORTER_OTLP_ENDPOINT is set —
	// see the extension-kit README for the full set of variables.
	extotel.InitOpenTelemetry()
	extbuild.PrintBuildInformation()
	extruntime.LogRuntimeInformation(zerolog.DebugLevel)

	exthealth.SetReady(false)
	exthealth.StartProbes(8088)

	// Most extensions require some form of configuration. These calls exist to parse and validate the
	// configuration obtained from environment variables.
	config.ParseConfiguration()
	config.ValidateConfiguration()

	discovery_kit_sdk.Register(extinstance.NewInstanceDiscovery())
	action_kit_sdk.RegisterAction(extmetric.NewMetricCheckAction())

	exthttp.RegisterRevisionedHandler("/", getExtensionList)

	extsignals.ActivateSignalHandlers()

	action_kit_sdk.RegisterCoverageEndpoints()

	exthealth.SetReady(true)
	exthttp.Listen(exthttp.ListenOpts{
		Port: 8087,
	})
}

type ExtensionListResponse struct {
	action_kit_api.ActionList       `json:",inline"`
	discovery_kit_api.DiscoveryList `json:",inline"`
}

func getExtensionList() ExtensionListResponse {
	return ExtensionListResponse{
		ActionList:    action_kit_sdk.GetActionList(),
		DiscoveryList: discovery_kit_sdk.GetDiscoveryList(),
	}
}
