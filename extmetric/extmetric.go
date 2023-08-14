// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extmetric

import (
	"context"
	"fmt"
	"github.com/prometheus/common/model"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-prometheus/extinstance"
	"time"
)

type MetricCheckAction struct {
}

type MetricCheckState struct {
	Command         []string `json:"command"`
	Pid             int      `json:"pid"`
	CmdStateID      string   `json:"cmdStateId"`
	Timestamp       string   `json:"timestamp"`
	StdOutLineCount int      `json:"stdOutLineCount"`
}

func NewMetricCheckAction() action_kit_sdk.Action[MetricCheckState] {
	return MetricCheckAction{}
}

// Make sure PostmanAction implements all required interfaces
var _ action_kit_sdk.Action[MetricCheckState] = (*MetricCheckAction)(nil)
var _ action_kit_sdk.ActionWithMetricQuery[MetricCheckState] = (*MetricCheckAction)(nil)

func (f MetricCheckAction) NewEmptyState() MetricCheckState {
	return MetricCheckState{}
}

func (f MetricCheckAction) Describe() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          fmt.Sprintf("%s.metrics", extinstance.PrometheusInstanceTargetId),
		Label:       "Prometheus metrics",
		Description: "Gather and check on Prometheus metrics",
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Icon:        extutil.Ptr(extinstance.PrometheusIcon),
		TargetSelection: extutil.Ptr(action_kit_api.TargetSelection{
			TargetType:          extinstance.PrometheusInstanceTargetId,
			QuantityRestriction: extutil.Ptr(action_kit_api.ExactlyOne),
			SelectionTemplates: extutil.Ptr([]action_kit_api.TargetSelectionTemplate{
				{
					Label:       "by instance-name",
					Description: extutil.Ptr("Find prometheus-instance by instance-name"),
					Query:       "prometheus.instance.name=\"\"",
				},
			}),
		}),
		Category:    extutil.Ptr("monitoring"),
		Kind:        action_kit_api.Check,
		TimeControl: action_kit_api.TimeControlExternal,
		Parameters: []action_kit_api.ActionParameter{
			{
				Label:        "Duration",
				Name:         "duration",
				Type:         "duration",
				Advanced:     extutil.Ptr(false),
				Required:     extutil.Ptr(true),
				DefaultValue: extutil.Ptr("30s"),
			},
		},
		Prepare: action_kit_api.MutatingEndpointReference{},
		Start:   action_kit_api.MutatingEndpointReference{},
		Metrics: extutil.Ptr(action_kit_api.MetricsConfiguration{
			Query: extutil.Ptr(action_kit_api.MetricsQueryConfiguration{
				Endpoint: action_kit_api.MutatingEndpointReferenceWithCallInterval{
					CallInterval: extutil.Ptr("1s"),
				},
				Parameters: []action_kit_api.ActionParameter{
					{
						Name:     "query",
						Label:    "PromQL Query",
						Required: extutil.Ptr(true),
						Type:     action_kit_api.String,
					},
				},
			}),
		}),
	}
}

func (f MetricCheckAction) Prepare(_ context.Context, _ *MetricCheckState, _ action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	return nil, nil
}

func (f MetricCheckAction) Start(_ context.Context, _ *MetricCheckState) (*action_kit_api.StartResult, error) {
	return nil, nil
}

func (f MetricCheckAction) QueryMetrics(ctx context.Context, request action_kit_api.QueryMetricsRequestBody) (*action_kit_api.QueryMetricsResult, error) {
	instance, err := extinstance.FindInstanceByName(request.Target.Name)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError(fmt.Sprintf("Failed to find Prometheus instance named '%s'", request.Target.Name), err))
	}

	client, err := instance.GetApiClient()
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to initialize Prometheus API client", err))
	}

	query := request.Config["query"]
	if query == nil {
		return nil, extutil.Ptr(extension_kit.ToError("No PromQL query defined", nil))
	}

	result, _, err := client.Query(ctx, query.(string), request.Timestamp)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError(fmt.Sprintf("Failed to execute Prometheus query against instance '%s' at timestamp %s with query '%s'",
			request.Target.Name,
			request.Timestamp,
			query),
			err))
	}

	vector, ok := result.(model.Vector)
	if !ok {
		return nil, extutil.Ptr(extension_kit.ToError("PromQL query returned unexpect result. Only vectors are supported as query results", nil))
	}

	metrics := make([]action_kit_api.Metric, len(vector))
	for i, sample := range vector {
		metric := make(map[string]string, len(sample.Metric))
		for key, value := range sample.Metric {
			metric[string(key)] = string(value)
		}
		metrics[i] = action_kit_api.Metric{
			Timestamp: sample.Timestamp.Time(),
			Metric:    metric,
			Value:     float64(sample.Value),
		}
	}

	return extutil.Ptr(action_kit_api.QueryMetricsResult{
		Metrics: extutil.Ptr(metrics),
	}), nil
}

type Metric struct {
	Timestamp time.Time         `json:"timestamp"`
	Metric    map[string]string `json:"metric"`
	Value     float64           `json:"value"`
}
