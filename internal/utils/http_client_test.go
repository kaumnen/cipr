package utils

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateProxyURL(t *testing.T) {
	for _, valid := range []string{"", "http://proxy.example:8080", "https://user:pass@proxy.example"} {
		assert.NoError(t, ValidateProxyURL(valid))
	}
	for _, invalid := range []string{"proxy.example:8080", "ftp://proxy.example", "https://"} {
		assert.Error(t, ValidateProxyURL(invalid))
	}
}

func TestConfiguredProxyRoutesEndpointRequest(t *testing.T) {
	oldProxy := viper.Get("proxy")
	t.Cleanup(func() { viper.Set("proxy", oldProxy) })

	var requestedURL string
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedURL = r.URL.String()
		_, _ = w.Write([]byte("proxied response"))
	}))
	defer proxy.Close()
	viper.Set("proxy", proxy.URL)

	body, err := loadFromEndpoint(context.Background(), "http://upstream.invalid/ranges?token=secret")
	require.NoError(t, err)
	assert.Equal(t, "proxied response", body)
	assert.Equal(t, "http://upstream.invalid/ranges?token=secret", requestedURL)
}

func TestHTTPClientFallsBackToProxyEnvironment(t *testing.T) {
	oldProxy := viper.Get("proxy")
	t.Cleanup(func() { viper.Set("proxy", oldProxy) })
	viper.Set("proxy", "")

	client, err := NewHTTPClient()
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodGet, "https://example.test/ranges", nil)
	require.NoError(t, err)

	transport := client.Transport.(*http.Transport)
	got, gotErr := transport.Proxy(req)
	want, wantErr := http.ProxyFromEnvironment(req)
	assert.Equal(t, wantErr, gotErr)
	if want == nil {
		assert.Nil(t, got)
	} else {
		require.NotNil(t, got)
		assert.Equal(t, want.String(), got.String())
	}
}

func TestDebugHTTPLoggingRedactsSecrets(t *testing.T) {
	oldProxy := viper.Get("proxy")
	t.Cleanup(func() { viper.Set("proxy", oldProxy) })
	viper.Set("proxy", "http://proxy-user:proxy-secret@proxy.example:8080")

	var output bytes.Buffer
	restoreWriter := setDebugWriter(&output)
	t.Cleanup(restoreWriter)
	SetDebug(true)
	t.Cleanup(func() { SetDebug(false) })

	client, err := NewHTTPClient()
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodGet, "https://endpoint-user:endpoint-secret@example.test/ranges?token=query-secret", nil)
	require.NoError(t, err)
	_, err = client.Transport.(*http.Transport).Proxy(req)
	require.NoError(t, err)

	log := output.String()
	assert.Contains(t, log, "[debug] proxy: configured")
	for _, secret := range []string{"proxy-user", "proxy-secret", "endpoint-user", "endpoint-secret", "query-secret"} {
		assert.NotContains(t, log, secret)
	}
}

func TestDebugDisabledProducesNoOutput(t *testing.T) {
	var output bytes.Buffer
	restoreWriter := setDebugWriter(&output)
	t.Cleanup(restoreWriter)
	SetDebug(false)

	Debugf("secret diagnostic")
	assert.Empty(t, output.String())
}

func TestSanitizeURL(t *testing.T) {
	got := SanitizeURL("https://user:password@example.test/ranges?token=secret#fragment")
	assert.Equal(t, "https://xxxxx:xxxxx@example.test/ranges", got)
}
