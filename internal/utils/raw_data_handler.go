package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

func GetRawData(provider string) string {
	endpointURL := ""
	localFile := ""

	if strings.HasPrefix(provider, "https://") || strings.HasPrefix(provider, "http://") {
		endpointURL = provider
	} else if strings.Contains(provider, "/") {
		localFile = provider
	} else {
		endpointURL = viper.GetString(provider + "_endpoint")
		localFile = viper.GetString(provider + "_local_file")
	}

	var ipRanges string
	var err error

	if localFile != "" {

		fmt.Println("Fetching IP ranges from local file:", localFile)
		ipRanges, err = loadFromFile(localFile)

		if err != nil {
			fmt.Println("Error reading local file:", err)
			os.Exit(1)

		}
	} else {
		if endpointURL == "" {
			fmt.Println("No endpoint URL or local file specified.")
			os.Exit(1)
		}

		fmt.Println("Fetching IP ranges from endpoint:", endpointURL)

		ipRanges, err = loadFromEndpoint(endpointURL)

		if err != nil {
			fmt.Println("Error fetching from endpoint:", err)
			os.Exit(1)
		}
	}

	return ipRanges
}

func loadFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func loadFromEndpoint(url string) (string, error) {
	response, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return "", fmt.Errorf("unexpected status %d from %s", response.StatusCode, url)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(responseData), nil
}
