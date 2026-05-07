package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// UserAgent is sent with every outgoing HTTP request. cmd populates it
// with the build version when one is set.
var UserAgent = "cipr"

var httpClient = &http.Client{Timeout: 30 * time.Second}

// GetRawData fetches IP-range data for the given source. The source may be
// a config-key prefix (e.g. "aws", looked up in viper), an http(s) URL, or
// a filesystem path containing a slash.
func GetRawData(ctx context.Context, source string) (string, error) {
	var endpointURL, localFile string

	switch {
	case strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "http://"):
		endpointURL = source
	case strings.Contains(source, "/"):
		localFile = source
	default:
		endpointURL = viper.GetString(source + "_endpoint")
		localFile = viper.GetString(source + "_local_file")
		if endpointURL == "" && localFile == "" {
			endpointURL = DefaultEndpoints[source]
		}
	}

	if localFile != "" {
		fmt.Fprintln(os.Stderr, "Fetching IP ranges from local file:", localFile)
		return loadFromFile(localFile)
	}

	if endpointURL == "" {
		return "", fmt.Errorf("no endpoint URL or local file specified for source %q", source)
	}

	fmt.Fprintln(os.Stderr, "Fetching IP ranges from endpoint:", endpointURL)
	return loadFromEndpoint(ctx, endpointURL)
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

	if response.StatusCode >= 400 {
		return "", fmt.Errorf("unexpected status %d from %s", response.StatusCode, url)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("read body from %s: %w", url, err)
	}
	return string(body), nil
}
