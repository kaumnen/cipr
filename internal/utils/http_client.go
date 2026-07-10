package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/viper"
)

const httpTimeout = 30 * time.Second

// ValidateProxyURL verifies that rawURL is an absolute HTTP(S) proxy URL.
// An empty value is valid and means that the standard proxy environment
// variables should be used.
func ValidateProxyURL(rawURL string) error {
	if rawURL == "" {
		return nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid proxy URL %q", rawURL)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("unsupported proxy URL scheme %q (only http and https are supported)", parsed.Scheme)
	}
	return nil
}

// NewHTTPClient returns a client configured from the effective proxy setting.
// A configured proxy overrides the standard environment proxy. With no
// configured value, ProxyFromEnvironment supplies HTTP_PROXY, HTTPS_PROXY,
// and NO_PROXY behavior.
func NewHTTPClient() (*http.Client, error) {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("unsupported default HTTP transport %T", http.DefaultTransport)
	}
	cloned := transport.Clone()

	configured := viper.GetString("proxy")
	if err := ValidateProxyURL(configured); err != nil {
		return nil, err
	}
	if configured != "" {
		proxyURL, _ := url.Parse(configured)
		cloned.Proxy = func(req *http.Request) (*url.URL, error) {
			Debugf("proxy: configured %s for %s", SanitizeURL(configured), SanitizeURL(req.URL.String()))
			return proxyURL, nil
		}
	} else {
		cloned.Proxy = func(req *http.Request) (*url.URL, error) {
			proxyURL, err := http.ProxyFromEnvironment(req)
			switch {
			case err != nil:
				Debugf("proxy: environment lookup failed for %s", SanitizeURL(req.URL.String()))
			case proxyURL != nil:
				Debugf("proxy: environment %s for %s", SanitizeURL(proxyURL.String()), SanitizeURL(req.URL.String()))
			default:
				Debugf("proxy: direct connection for %s", SanitizeURL(req.URL.String()))
			}
			return proxyURL, err
		}
	}

	return &http.Client{Transport: cloned, Timeout: httpTimeout}, nil
}
