// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extmetric

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	retry "github.com/sethvargo/go-retry"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-prometheus/v2/config"
	"github.com/steadybit/extension-prometheus/v2/extinstance"
)

type MetricCheckAction struct {
}

type MetricCheckState struct {
	Command         []string  `json:"command"`
	Pid             int       `json:"pid"`
	CmdStateID      string    `json:"cmdStateId"`
	Timestamp       string    `json:"timestamp"`
	StdOutLineCount int       `json:"stdOutLineCount"`
	ExecutionId     uuid.UUID `json:"executionId"`
}

func NewMetricCheckAction() action_kit_sdk.Action[MetricCheckState] {
	return MetricCheckAction{}
}

// Make sure PrometheusAction implements all required interfaces
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
		Technology:  extutil.Ptr("Prometheus"),
		Category:    extutil.Ptr("Prometheus"), //Can be removed in Q1/24 - support for backward compatibility of old sidebar
		TargetSelection: extutil.Ptr(action_kit_api.TargetSelection{
			TargetType:          extinstance.PrometheusInstanceTargetId,
			QuantityRestriction: extutil.Ptr(action_kit_api.QuantityRestrictionExactlyOne),
			SelectionTemplates: extutil.Ptr([]action_kit_api.TargetSelectionTemplate{
				{
					Label:       "instance-name",
					Description: extutil.Ptr("Find prometheus-instance by instance-name"),
					Query:       "prometheus.instance.name=\"\"",
				},
			}),
		}),
		Kind:        action_kit_api.Check,
		TimeControl: action_kit_api.TimeControlExternal,
		Parameters: []action_kit_api.ActionParameter{
			{
				Label:        "Duration",
				Name:         "duration",
				Type:         action_kit_api.ActionParameterTypeDuration,
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
						Type:     action_kit_api.ActionParameterTypeString,
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

	retries := config.Config.QueryRetries

	// Use QueryRange instead of Query to get actual metric timestamps
	start := request.Timestamp.Add(-time.Duration(1) * time.Second) // Adjust start time to ensure we capture the last second of data, matching the call interval
	end := request.Timestamp
	step := 1 * time.Second

	r := v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	var result model.Value
	err = retry.Do(ctx, retry.WithMaxRetries(uint64(retries), retry.NewFibonacci(50*time.Millisecond)), func(ctx context.Context) error {
		value, warnings, err := client.QueryRange(ctx, query.(string), r)
		if err != nil {
			return retry.RetryableError(err)
		}
		if len(warnings) > 0 {
			log.Info().Str("query", query.(string)).Strs("warnings", warnings).Msg("Warnings returned from query.")
		}

		result = value
		return nil
	})
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError(fmt.Sprintf("Failed to execute Prometheus range query against instance '%s' from %s to %s with query '%s'",
			request.Target.Name,
			start,
			end,
			query),
			err))
	}

	// QueryRange returns a matrix instead of a vector
	matrix, ok := result.(model.Matrix)
	if !ok {
		return nil, extutil.Ptr(extension_kit.ToError("PromQL range query returned unexpected result. Expected matrix type as query result", nil))
	}

	// Process the matrix result
	var metrics []action_kit_api.Metric
	for _, sampleStream := range matrix {
		// For each time series in the matrix
		metricLabels := make(map[string]string, len(sampleStream.Metric))
		for key, value := range sampleStream.Metric {
			metricLabels[string(key)] = string(value)
		}

		// For each sample in the time series
		if len(sampleStream.Values) == 0 {
			log.Warn().Msgf("No samples found for query '%s'", query)
			continue
		}
		for _, samplePair := range sampleStream.Values {
			metric := action_kit_api.Metric{
				Timestamp:       samplePair.Timestamp.Time(),
				TimestampSource: extutil.Ptr(action_kit_api.TimestampSourceExternal),
				Metric:          metricLabels,
				Value:           float64(samplePair.Value),
			}
			metrics = append(metrics, metric)
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
