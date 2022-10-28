// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extmetric

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/prometheus/common/model"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-prometheus/extinstance"
	"net/http"
	"time"
)

func RegisterMetricCheckHandlers() {
	exthttp.RegisterHttpHandler("/prometheus/metrics", exthttp.GetterAsHandler(getMetricCheckDescription))
	exthttp.RegisterHttpHandler("/prometheus/metrics/prepare", noopHandler)
	exthttp.RegisterHttpHandler("/prometheus/metrics/start", noopHandler)
	exthttp.RegisterHttpHandler("/prometheus/metrics/query", query)
}

func noopHandler(_ http.ResponseWriter, _ *http.Request, _ []byte) {

}

func getMetricCheckDescription() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          fmt.Sprintf("%s.metrics", extinstance.PrometheusInstanceTargetId),
		Label:       "Prometheus metrics",
		Description: "Gather and check on Prometheus metrics",
		Version:     "1.1.1",
		Icon:        extutil.Ptr(extinstance.PrometheusIcon),
		TargetType:  extutil.Ptr(extinstance.PrometheusInstanceTargetId),
		Category:    extutil.Ptr("monitoring"),
		Kind:        action_kit_api.Check,
		TimeControl: action_kit_api.External,
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
		Prepare: action_kit_api.MutatingEndpointReference{
			Method: action_kit_api.Post,
			Path:   "/prometheus/metrics/prepare",
		},
		Start: action_kit_api.MutatingEndpointReference{
			Method: action_kit_api.Post,
			Path:   "/prometheus/metrics/start",
		},
		Metrics: extutil.Ptr(action_kit_api.MetricsConfiguration{
			Query: extutil.Ptr(action_kit_api.MetricsQueryConfiguration{
				Endpoint: action_kit_api.MutatingEndpointReferenceWithCallInterval{
					Method:       action_kit_api.Post,
					Path:         "/prometheus/metrics/query",
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

type Metric struct {
	Timestamp time.Time         `json:"timestamp"`
	Metric    map[string]string `json:"metric"`
	Value     float64           `json:"value"`
}

func query(w http.ResponseWriter, _ *http.Request, body []byte) {
	result, err := Query(body)
	if err != nil {
		exthttp.WriteError(w, *err)
	} else {
		exthttp.WriteBody(w, *result)
	}
}

func Query(body []byte) (*action_kit_api.QueryMetricsResult, *extension_kit.ExtensionError) {
	var request action_kit_api.QueryMetricsRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to parse request body", err))
	}

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

	result, _, err := client.Query(context.TODO(), query.(string), request.Timestamp)
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
