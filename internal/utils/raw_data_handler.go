package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// UserAgent is sent with every outgoing HTTP request. cmd populates it
// with the build version when one is set.
var UserAgent = "cipr"

var httpClient = &http.Client{Timeout: 30 * time.Second}

const defaultCacheTTL = 24 * time.Hour
const maxResponseBytes = 64 << 20

// GetRawData fetches IP-range data for the given source. The source may be
// a config-key prefix (e.g. "aws", looked up in viper), an HTTP(S) URL, or
// a filesystem path.
func GetRawData(ctx context.Context, source string) (string, error) {
	var endpointURL, localFile, cacheKey string

	switch {
	case source == "":
		return "", fmt.Errorf("source cannot be empty")
	case strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "http://"):
		if err := ValidateHTTPURL(source); err != nil {
			return "", err
		}
		endpointURL = source
	case IsConfiguredSource(source):
		endpointURL = viper.GetString(source + "_endpoint")
		localFile = viper.GetString(source + "_local_file")
		if endpointURL == "" && localFile == "" {
			endpointURL = DefaultEndpoints[source]
		}
		if localFile == "" {
			cacheKey = source
		}
	case strings.Contains(source, "://"):
		return "", ValidateHTTPURL(source)
	default:
		localFile = source
	}

	if localFile != "" {
		fmt.Fprintln(os.Stderr, "Fetching IP ranges from local file:", localFile)
		return loadFromFile(localFile)
	}

	if endpointURL == "" {
		return "", fmt.Errorf("no endpoint URL or local file specified for source %q", source)
	}
	if err := ValidateHTTPURL(endpointURL); err != nil {
		return "", fmt.Errorf("invalid endpoint for source %q: %w", source, err)
	}

	return GetCached(ctx, cacheKey, func(ctx context.Context) (string, error) {
		fmt.Fprintln(os.Stderr, "Fetching IP ranges from endpoint:", endpointURL)
		return loadFromEndpoint(ctx, endpointURL)
	})
}

// ValidateHTTPURL verifies that rawURL is an absolute HTTP(S) URL suitable for
// endpoint fetching.
func ValidateHTTPURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid source URL %q", rawURL)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("unsupported source URL scheme %q (only http and https are supported)", parsed.Scheme)
	}
	return nil
}

func loadFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", filePath, err)
	}
	return string(data), nil
}

func loadFromEndpoint(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request for %s: %w", url, err)
	}
	req.Header.Set("User-Agent", UserAgent)

	response, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", url, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("unexpected status %d from %s", response.StatusCode, url)
	}

	body, err := readAllLimited(response.Body, maxResponseBytes)
	if err != nil {
		return "", fmt.Errorf("read body from %s: %w", url, err)
	}
	return string(body), nil
}

func readAllLimited(r io.Reader, limit int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("response exceeds %d byte limit", limit)
	}
	return body, nil
}
