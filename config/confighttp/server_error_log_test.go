// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package confighttp

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestHTTPServerErrorLog(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	settings := componenttest.NewNopTelemetrySettings()
	settings.Logger = zap.New(core)

	server, err := (&ServerConfig{}).ToServer(t.Context(), nil, settings, http.NewServeMux())
	require.NoError(t, err)

	tests := []struct {
		name    string
		message string
		level   zapcore.Level
	}{
		{
			name:    "TLS handshake EOF",
			message: "http: TLS handshake error from 127.0.0.1:12345: EOF",
			level:   zapcore.DebugLevel,
		},
		{
			name:    "TLS handshake error",
			message: "http: TLS handshake error from 127.0.0.1:12345: remote error: tls: bad certificate",
			level:   zapcore.ErrorLevel,
		},
		{
			name:    "other server error",
			message: "http: panic serving 127.0.0.1:12345: boom",
			level:   zapcore.ErrorLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.ErrorLog.Print(tt.message)

			entries := observed.All()
			require.Len(t, entries, 1)
			assert.Equal(t, tt.level, entries[0].Level)
			assert.Equal(t, tt.message, entries[0].Message)
			observed.TakeAll()
		})
	}
}
