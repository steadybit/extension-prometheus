// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extinstance

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/steadybit/extension-prometheus/v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstance_IsAuthenticated(t *testing.T) {
	tests := []struct {
		name        string
		headerKey   string
		headerValue string
		expected    bool
	}{
		{
			name:        "both key and value present",
			headerKey:   "Authorization",
			headerValue: "Bearer token123",
			expected:    true,
		},
		{
			name:        "empty key",
			headerKey:   "",
			headerValue: "Bearer token123",
			expected:    false,
		},
		{
			name:        "empty value",
			headerKey:   "Authorization",
			headerValue: "",
			expected:    false,
		},
		{
			name:        "both empty",
			headerKey:   "",
			headerValue: "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := &Instance{
				HeaderKey:   tt.headerKey,
				HeaderValue: tt.headerValue,
			}
			assert.Equal(t, tt.expected, instance.IsAuthenticated())
		})
	}
}

func TestHeaderRoundTripper_RoundTrip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-value", r.Header.Get("Test-Header"))
		assert.Equal(t, "steadybit-extension-prometheus", r.Header.Get("User-Agent"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	headers := map[string][]string{
		"Test-Header": {"test-value"},
		"User-Agent":  {"steadybit-extension-prometheus"},
	}

	rt := &headerRoundTripper{
		headers: headers,
		rt:      http.DefaultTransport,
	}

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLoggingRoundTripper_RoundTrip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	rt := &loggingRoundTripper{
		rt: http.DefaultTransport,
	}

	req, err := http.NewRequest("GET", server.URL, strings.NewReader("test body"))
	require.NoError(t, err)

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestInstance_GetApiClient(t *testing.T) {
	config.Config = config.Specification{
		InsecureSkipVerify:   false,
		EnableRequestLogging: false,
	}

	instance := &Instance{
		Name:        "test-instance",
		BaseUrl:     "http://localhost:9090",
		HeaderKey:   "Authorization",
		HeaderValue: "Bearer test-token",
	}

	client, err := instance.GetApiClient()
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestInstance_GetApiClient_WithLogging(t *testing.T) {
	config.Config = config.Specification{
		InsecureSkipVerify:   false,
		EnableRequestLogging: true,
	}

	instance := &Instance{
		Name:    "test-instance",
		BaseUrl: "http://localhost:9090",
	}

	client, err := instance.GetApiClient()
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestFindInstanceByName(t *testing.T) {
	originalInstances := Instances
	defer func() { Instances = originalInstances }()

	Instances = []Instance{
		{Name: "prometheus-1", BaseUrl: "http://localhost:9090"},
		{Name: "prometheus-2", BaseUrl: "http://localhost:9091"},
	}

	t.Run("found", func(t *testing.T) {
		instance, err := FindInstanceByName("prometheus-1")
		require.NoError(t, err)
		assert.Equal(t, "prometheus-1", instance.Name)
		assert.Equal(t, "http://localhost:9090", instance.BaseUrl)
	})

	t.Run("not found", func(t *testing.T) {
		instance, err := FindInstanceByName("nonexistent")
		assert.Error(t, err)
		assert.Nil(t, instance)
		assert.Equal(t, "not found", err.Error())
	})
}
