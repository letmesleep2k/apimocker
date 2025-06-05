package main

import (
	// "bytes"
	// "encoding/base64"
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
