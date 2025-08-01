// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extinstance

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/extension-prometheus/v2/config"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

type Instance struct {
	Name        string `json:"name"`
	BaseUrl     string `json:"baseUrl"`
	HeaderKey   string `json:"headerKey"`
	HeaderValue string `json:"headerValue"`
}

func (i *Instance) IsAuthenticated() bool {
	return len(i.HeaderKey) > 0 && len(i.HeaderValue) > 0
}

// headerRoundTripper is a custom transport that adds headers to each request
type headerRoundTripper struct {
	headers map[string][]string
	rt      http.RoundTripper
}

// RoundTrip implements the http.RoundTripper interface
func (h *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, values := range h.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return h.rt.RoundTrip(req)
}

type loggingRoundTripper struct {
	rt http.RoundTripper
}

func (l *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyStr string
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to read request body")
		} else {
			bodyStr = string(bodyBytes)
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}
	resp, err := l.rt.RoundTrip(req)
	if err != nil {
		log.Warn().Err(err).Str("method", req.Method).Msg("Error during HTTP request")
		return nil, err
	}
	log.Debug().
		Str("method", req.Method).
		Str("url", req.URL.String()).
		Int("status", resp.StatusCode).
		Str("request_body", bodyStr).
		Str("response_body", func() string {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to read response body")
				return ""
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			return string(bodyBytes)
		}()).
		Msg("Received HTTP response")
	return resp, nil
}

func (i *Instance) GetApiClient() (prometheus.API, error) {

	headers := map[string][]string{
		"User-Agent": {"steadybit-extension-prometheus"},
	}

	if i.IsAuthenticated() {
		headers[i.HeaderKey] = []string{i.HeaderValue}
	}

	transport := http.Transport{
		//custom timeouts:
		ResponseHeaderTimeout: 5 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
		//from default roundtripper:
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.Config.InsecureSkipVerify,
		},
	}

	var rt http.RoundTripper
	if config.Config.EnableRequestLogging {
		rt = &loggingRoundTripper{
			rt: &transport,
		}
	} else {
		rt = &transport
	}

	roundTripper := &headerRoundTripper{
		headers: headers,
		rt:      rt,
	}

	apiClient, err := api.NewClient(api.Config{
		Address:      i.BaseUrl,
		RoundTripper: roundTripper,
	})
	if err != nil {
		return nil, err
	}
	client := prometheus.NewAPI(apiClient)
	return client, nil
}

var (
	Instances []Instance
)

func init() {
	name := getInstanceName(0)
	for len(name) > 0 {
		index := len(Instances)
		Instances = append(Instances, Instance{
			Name:        name,
			BaseUrl:     getInstanceOrigin(index),
			HeaderKey:   getAuthHeaderKey(index),
			HeaderValue: getAuthHeaderValue(index),
		})
		name = getInstanceName(len(Instances))
	}
}

func getInstanceName(n int) string {
	return os.Getenv(fmt.Sprintf("STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_%d_NAME", n))
}

func getInstanceOrigin(n int) string {
	return os.Getenv(fmt.Sprintf("STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_%d_ORIGIN", n))
}

func getAuthHeaderKey(n int) string {
	return os.Getenv(fmt.Sprintf("STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_%d_HEADER_KEY", n))
}

func getAuthHeaderValue(n int) string {
	return os.Getenv(fmt.Sprintf("STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_%d_HEADER_VALUE", n))
}

func FindInstanceByName(name string) (*Instance, error) {
	for _, i := range Instances {
		if i.Name == name {
			return &i, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
