// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extinstance

import (
	"crypto/tls"
	"fmt"
	"github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/steadybit/extension-prometheus/v2/config"
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

func (i *Instance) GetApiClient() (prometheus.API, error) {
	roundTripper := &http.Transport{
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
