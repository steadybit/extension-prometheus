// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package main

import (
	"github.com/rs/zerolog"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthealth"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-kit/extruntime"
	"github.com/steadybit/extension-prometheus/config"
	"github.com/steadybit/extension-prometheus/extinstance"
	"github.com/steadybit/extension-prometheus/extmetric"
)

func main() {
	extlogging.InitZeroLog()
	extbuild.PrintBuildInformation()
	extruntime.LogRuntimeInformation(zerolog.DebugLevel)

	exthealth.SetReady(false)
	exthealth.StartProbes(8088)

	// Most extensions require some form of configuration. These calls exist to parse and validate the
	// configuration obtained from environment variables.
	config.ParseConfiguration()
	config.ValidateConfiguration()

	exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(getExtensionList))
	extinstance.RegisterInstanceDiscoveryHandlers()
	action_kit_sdk.RegisterAction(extmetric.NewMetricCheckAction())

	action_kit_sdk.InstallSignalHandler()

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
		ActionList: action_kit_sdk.GetActionList(),
		DiscoveryList: discovery_kit_api.DiscoveryList{
			Discoveries: []discovery_kit_api.DescribingEndpointReference{
				{
					Method: "GET",
					Path:   "/prometheus/instance/discovery",
				},
			},
			TargetTypes: []discovery_kit_api.DescribingEndpointReference{
				{
					Method: "GET",
					Path:   "/prometheus/instance/discovery/target-description",
				},
			},
			TargetAttributes: []discovery_kit_api.DescribingEndpointReference{
				{
					Method: "GET",
					Path:   "/prometheus/instance/discovery/attribute-descriptions",
				},
			},
		},
	}
}
