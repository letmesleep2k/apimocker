package main

import (
	// "bytes"
	"encoding/base64"
	// "encoding/json"
	// "fmt"
	// "io"
	// "net/http"
	// "net/http/httptest"
	// "net/url"
	"os"
	// "strings"
	"testing"
	"path/filepath"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configData  string
		filename    string
		wantErr     bool
		expectedPort int
	}{
		{
			name: "valid YAML config",
			configData: `
port: 8080
logging:
  enabled: true
  format: json
  output: stdout
endpoints:
  - path: /users
    method: GET
    count: 5
    data: '{"id": "uuid", "name": "name"}'
`,
			filename:    "test.yaml",
			wantErr:     false,
			expectedPort: 8080,
		},
		{
			name: "valid JSON config",
			configData: `{
  "port": 9090,
  "logging": {
    "enabled": false,
    "format": "plain",
    "output": "test.log"
  },
  "endpoints": [
    {
      "path": "/api/data",
      "method": "POST",
      "status": 201,
      "count": 10
    }
  ]
}`,
			filename:    "test.json",
			wantErr:     false,
			expectedPort: 9090,
		},
		{
			name:       "invalid config file",
			configData: "invalid: yaml: content:",
			filename:   "test.yaml",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "config-*"+filepath.Ext(tt.filename))
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tt.configData)
			require.NoError(t, err)
			tmpFile.Close()

			config, err := loadConfig(tmpFile.Name())

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedPort, config.Port)

			for _, endpoint := range config.Endpoints {
				if endpoint.Status == 0 {
					t.Errorf("Status should be set to default 200, got %d", endpoint.Status)
				}
				if endpoint.Count == 0 {
					t.Errorf("Count should be set to default 1, got %d", endpoint.Count)
				}
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input string
		expected time.Duration
	}{
		{"", 0},
		{"100ms", 100 * time.Millisecond},
		{"5s", 5 * time.Second},
		{"2m", 2 * time.Minute},
		{"1h30m", 90 * time.Minute},
		{"invalid", 0},
		{"100", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseDuration(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldTriggerError(t *testing.T) {
	tests := []struct {
		name string
		errors []ErrorConfig
		runs int
	}{
		{
			name: "no errors configured",
			errors: []ErrorConfig{},
			runs: 10,
		},
		{
			name: "100% probability error",
			errors: []ErrorConfig{
				{Probability: 1.0, Status: 500, Message: "Always fails"},
			},
			runs: 5,
		},
		{
			name: "0% probability error",
			errors: []ErrorConfig{
				{Probability: 0.0, Status: 500, Message: "Never fails"},
			},
			runs: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorCount := 0
			for i := 0; i < tt.runs; i++ {
				triggered, _ := shouldTriggerError(tt.errors)
				if triggered {
					errorCount++
				}
			}

			if len(tt.errors) == 0 {
				assert.Equal(t, 0, errorCount, "No error should be triggered")
			} else if tt.errors[0].Probability == 1.0 {
				assert.Equal(t, tt.runs, errorCount, "All requests should trigger error")
			} else if tt.errors[0].Probability == 0.0 {
				assert.Equal(t, 0, errorCount, "No requests should trigger error")
			}
		})
	}
}

func TestAuthenticateBasic(t *testing.T) {
	authConfig := &AuthConfig{
		Type: "basic",
		Username: "testuser",
		Password: "testpass",
	}

	tests := []struct{
		name string
		authHeader string
		wantAuth bool
		wantResult string
	}{
		{
			name: "valid credentials",
			authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("testuser:testpass")),
			wantAuth: true,
			wantResult: "success",
		},
		{
			name: "invalid credentials",
			authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("wrong:wrong")),
			wantAuth: false,
			wantResult: "invalid-credentials",
		},
		{
			name: "invalid format",
			authHeader: "Bearer token123",
			wantAuth: false,
			wantResult: "invalid-basic-format",
		},
		{
			name: "invalid base64",
			authHeader: "Basic invalid-base64!",
			wantAuth: false,
			wantResult: "invalid-base64",
		},
		{
			name: "missing colon",
			authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("userpass")),
			wantAuth: false,
			wantResult: "invalid-credentials-format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			success, authType, result := authenticateBasic(tt.authHeader, authConfig)
			assert.Equal(t, tt.wantAuth, success)
			assert.Equal(t, "basic", authType)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}

func TestAuthenticateBearer(t *testing.T) {
	authConfig := &AuthConfig{
		Type: "bearer",
		Token: "valid-token-123",
	}

	tests := []struct {
		name string
		authHeader string
		wantAuth bool
		wantResult string
	}{
		{
			name: "valid token",
			authHeader: "Bearer valid-token-123",
			wantAuth: true,
			wantResult: "success",
		},
		{
			name: "invalid token",
			authHeader: "Bearer wrong-token",
			wantAuth: false,
			wantResult: "invalid-token",
		},
		{
			name: "invalid format",
			authHeader: "Basic dGVzdA==",
			wantAuth: false,
			wantResult: "invalid-bearer-format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			success, authType, result := authenticateBearer(tt.authHeader, authConfig)
			assert.Equal(t, tt.wantAuth, success)
			assert.Equal(t, "bearer", authType)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
