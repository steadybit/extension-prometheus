// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extinstance

import (
	"context"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_commons"
	"github.com/steadybit/discovery-kit/go/discovery_kit_sdk"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-prometheus/config"
	"time"
)

type instanceDiscovery struct {
}

var (
	_ discovery_kit_sdk.TargetDescriber    = (*instanceDiscovery)(nil)
	_ discovery_kit_sdk.AttributeDescriber = (*instanceDiscovery)(nil)
)

func NewInstanceDiscovery() discovery_kit_sdk.TargetDiscovery {
	discovery := &instanceDiscovery{}
	return discovery_kit_sdk.NewCachedTargetDiscovery(discovery,
		discovery_kit_sdk.WithRefreshTargetsNow(),
		discovery_kit_sdk.WithRefreshTargetsInterval(context.Background(), 30*time.Second),
	)
}

func (d *instanceDiscovery) Describe() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id:         PrometheusInstanceTargetId,
		RestrictTo: extutil.Ptr(discovery_kit_api.LEADER),
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			CallInterval: extutil.Ptr("30s"),
		},
	}
}

func (d *instanceDiscovery) DescribeTarget() discovery_kit_api.TargetDescription {
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

func (d *instanceDiscovery) DescribeAttributes() []discovery_kit_api.AttributeDescription {
	return []discovery_kit_api.AttributeDescription{
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
	}
}

func (d *instanceDiscovery) DiscoverTargets(_ context.Context) ([]discovery_kit_api.Target, error) {
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

	return discovery_kit_commons.ApplyAttributeExcludes(targets, config.Config.DiscoveryAttributesExcludesInstance), nil
}
