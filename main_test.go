package main

import (
	// "bytes"
	"encoding/base64"
	// "encoding/json"
	// "fmt"
	"io"
	// "net/http"
	"net/http/httptest"
	"net/url"
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

func TestAuthenticateRequest(t *testing.T) {
	tests := []struct{
		name string
		authConfig *AuthConfig
		authHeader string
		wantAuth bool
		wantType string
		wantResult string
	}{
		{
			name: "no auth required",
			authConfig: nil,
			authHeader: "",
			wantAuth: true,
			wantType: "",
			wantResult: "no-auth",
		},
		{
			name: "valid basic auth",
			authConfig: &AuthConfig{
				Type: "basic",
				Username: "user",
				Password: "pass",
			},
			authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass")),
			wantAuth: true,
			wantType: "basic",
			wantResult: "success",
		},
		{
			name: "missing auth header",
			authConfig: &AuthConfig{
				Type: "bearer",
				Token: "token123",
			},
			authHeader: "",
			wantAuth: false,
			wantType: "bearer",
			wantResult: "missing-auth",
		},
		{
			name: "invalid auth type",
			authConfig: &AuthConfig{
				Type: "invalid",
			},
			authHeader: "Bearer token",
			wantAuth: false,
			wantType: "invalid",
			wantResult: "invalid-auth-type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			success, authType, result := authenticateRequest(req, tt.authConfig)
			assert.Equal(t, tt.wantAuth, success)
			assert.Equal(t, tt.wantType, authType)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}

func TestGenerateFakeData(t *testing.T) {
	tests := []struct {
		name string
		schema string
		count int
	}{
		{
			name: "valid JSON schema",
			schema: `{"id": "uuid", "name": "name", "email": "email"}`,
			count: 3,
		},
		{
			name: "invalid JSON schema - fallbakc to faker",
			schema: "invalid json",
			count: 2,
		},
		{
			name: "supported field types",
			schema: `{"id": "uuid", "flag": "bool", "number": "int", "location": "lat"}`,
			count: 1,
		},
		{
			name: "unsupported field type",
			schema: `{"id": "uuid", "unknown": "unsupported"}`,
			count: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := generateFakeData(tt.schema, tt.count)
			require.NoError(t, err)
			assert.Len(t, data, tt.count)

			if tt.count > 0 {
				assert.IsType(t, []map[string]interface{}{}, data)
				firstItem := data[0]
				assert.NotEmpty(t, firstItem)
			}
		})
	}
}

func TestApplyQueryFilters(t *testing.T) {
	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice", "age": 30},
		{"id": 2, "name": "Bob", "age": 25},
		{"id": 3, "name": "Charlie", "age": 35},
		{"id": 4, "name": "Alice", "age": 28},
	}

	tests := []struct{
		name string
		params map[string]string
		expected int
	}{
		{
			name: "no filters",
			params: map[string]string{},
			expected: 4,
		},
		{
			name: "filter by name",
			params: map[string]string{"filter": "name:Alice"},
			expected: 2,
		},
		{
			name: "limit count",
			params: map[string]string{"count": "2"},
			expected: 2,
		},
		{
			name: "limit with alias",
			params: map[string]string{"limit": "3"},
			expected: 3,
		},
		{
			name: "offset",
			params: map[string]string{"offset": "1", "count": "2"},
			expected: 2,
		},
		{
			name: "sort ascending",
			params: map[string]string{"sort": "age", "order": "asc"},
			expected: 4,
		},
		{
			name: "sort descending",
			params: map[string]string{"sort": "age", "order": "desc"},
			expected: 4,
		},
		{
			name: "offset beyond data",
			params: map[string]string{"offset": "10"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := url.Values{}
			for k, v := range tt.params {
				values.Set(k, v)
			}

			result := applyQueryFilters(testData, values)
			assert.Len(t, result, tt.expected)

			if tt.params["sort"] == "age" {
				if len(result) >= 2 {
					if tt.params["order"] == "desc" {
						assert.True(t, result[0]["age"].(int) >= result[1]["age"].(int))
					} else {
						assert.True(t, result[0]["age"].(int) <= result[1]["age"].(int))
					}
				}
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name string
		config LogConfig
		wantErr bool
	} {
		{
			name: "disabled logger",
			config: LogConfig{
				Enabled: false,
				Format: "json",
				Output: "stdout",
			},
			wantErr: false,
		},
		{
			name: "stdout logger",
			config: LogConfig{
				Enabled: true,
				Format: "plain",
				Output: "stdout",
			},
			wantErr: false,
		},
		{
			name: "file logger",
			config: LogConfig{
				Enabled: true,
				Format: "json",
				Output: "/tmp/test.log",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)

			if tt.wantErr{
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, logger)

			reqLog := RequestLog{
				Timestamp: time.Now().Format(time.RFC3339),
				Method: "GET",
				Path: "/test",
				StatusCode: 200,
				ResponseTime: "10ms",
				RemoteAddr: "127.0.0.1",
			}

			logger.LogRequest(reqLog)
		})
	}
}

func TestCreateLoggingHandler(t *testing.T) {
	logger := &Logger{writer: io.Discard, format: "json"}

	tests := []struct {
		name string
		endpoint Endpoint
		method string
		path string
		authHeader string
		expectedStatus int
	}{
		{
			name: "successful request",
			endpoint: Endpoint{
				Path: "/test",
				Method: "GET",
				Status: 200,
				Data: `{"id": "uuid"}`,
				Count: 1,
			},
			method: "GET",
			path: "/test",
			expectedStatus: 200,
		},
		{
			name: "method not allowed",
			endpoint: Endpoint{
				Path: "/test",
				Method: "POST",
				Status: 200,
			},
			method: "GET",
			path: "/test",
			expectedStatus: 405,
		},
		{
			name: "unauthorized request",
			endpoint: Endpoint{
				Path: "/secure",
				Method: "GET",
				Status: 200,
				Auth: &AuthConfig{
					Type: "bearer",
					Token: "secret",
				},
			},
			method: "GET",
			path: "/secure",
			expectedStatus: 401,
		},
		{
			name: "authorized request",
			endpoint: Endpoint{
				Path: "/secure",
				Method: "GET",
				Status: 200,
				Data: `{"message": "success"}`,
				Auth: &AuthConfig{
					Type: "bearer",
					Token: "secret",
				},
			},
			method: "GET",
			path: "/secure",
			authHeader: "Bearer secret",
			expectedStatus: 200,
		},
		{
			name: "request with delay",
			endpoint: Endpoint{
				Path: "/slow",
				Method: "GET",
				Status: 200,
				Data: `{"slow": true}`,
				Delay: "10ms",
			},
			method: "GET",
			path: "/slow",
			expectedStatus: 200,
		},
		{
			name: "request with custom headers",
			endpoint: Endpoint{
				Path: "/headers",
				Method: "GET",
				Status: 200,
				Data: `{"test": true}`,
				Headers: map[string]string{
					"X-Custom-Header": "test-value",
					"X-API-Version": "v1",
				},
			},
			method: "GET",
			path: "/headers",
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := createLoggingHandler(tt.endpoint, logger)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == 200 {
				for key, expectedValue := range tt.endpoint.Headers {
					assert.Equal(t, expectedValue, rr.Header().Get(key))
				}

				if tt.endpoint.Data != "" {
					assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				}
			}
		})
	}
}
