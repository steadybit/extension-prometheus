// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extinstance

import (
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_commons"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-prometheus/config"
)

func RegisterInstanceDiscoveryHandlers() {
	exthttp.RegisterHttpHandler("/prometheus/instance/discovery", exthttp.GetterAsHandler(getPrometheusInstanceDiscoveryDescription))
	exthttp.RegisterHttpHandler("/prometheus/instance/discovery/target-description", exthttp.GetterAsHandler(getPrometheusInstanceTargetDescription))
	exthttp.RegisterHttpHandler("/prometheus/instance/discovery/attribute-descriptions", exthttp.GetterAsHandler(getPrometheusInstanceAttributeDescriptions))
	exthttp.RegisterHttpHandler("/prometheus/instance/discovery/discovered-targets", exthttp.GetterAsHandler(getPrometheusInstanceDiscoveryResults))
}

func getPrometheusInstanceDiscoveryDescription() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id:         PrometheusInstanceTargetId,
		RestrictTo: extutil.Ptr(discovery_kit_api.LEADER),
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			Method:       "GET",
			Path:         "/prometheus/instance/discovery/discovered-targets",
			CallInterval: extutil.Ptr("30s"),
		},
	}
}

func getPrometheusInstanceTargetDescription() discovery_kit_api.TargetDescription {
	return discovery_kit_api.TargetDescription{
		Id:       PrometheusInstanceTargetId,
		Label:    discovery_kit_api.PluralLabel{One: "Prometheus instance", Other: "Prometheus instances"},
		Category: extutil.Ptr("monitoring"),
		Version:  extbuild.GetSemverVersionStringOrUnknown(),
		Icon:     extutil.Ptr(PrometheusIcon),
		Table: discovery_kit_api.Table{
			Columns: []discovery_kit_api.Column{
				{Attribute: "prometheus.instance.name"},
				{Attribute: "prometheus.instance.url"},
			},
			OrderBy: []discovery_kit_api.OrderBy{
				{
					Attribute: "prometheus.instance.name",
					Direction: "ASC",
				},
			},
		},
	}
}

func getPrometheusInstanceAttributeDescriptions() discovery_kit_api.AttributeDescriptions {
	return discovery_kit_api.AttributeDescriptions{
		Attributes: []discovery_kit_api.AttributeDescription{
			{
				Attribute: "prometheus.instance.name",
				Label: discovery_kit_api.PluralLabel{
					One:   "Prometheus instance name",
					Other: "Prometheus instance names",
				},
			}, {
				Attribute: "prometheus.instance.url",
				Label: discovery_kit_api.PluralLabel{
					One:   "Prometheus instance URL",
					Other: "Prometheus instance URLs",
				},
			},
		},
	}
}

func getPrometheusInstanceDiscoveryResults() discovery_kit_api.DiscoveredTargets {
	targets := make([]discovery_kit_api.Target, len(Instances))

	for i, instance := range Instances {
		targets[i] = discovery_kit_api.Target{
			Id:         instance.Name,
			Label:      instance.Name,
			TargetType: PrometheusInstanceTargetId,
			Attributes: map[string][]string{
				"prometheus.instance.name": {instance.Name},
				"prometheus.instance.url":  {instance.BaseUrl},
			},
		}
	}

	return discovery_kit_api.DiscoveredTargets{Targets: discovery_kit_commons.ApplyAttributeExcludes(targets, config.Config.DiscoveryAttributesExcludesInstance)}
}
