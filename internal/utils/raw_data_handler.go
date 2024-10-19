package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/viper"
)

func GetRawData(provider string) string {
	endpointURL := viper.GetString(provider + "_endpoint")
	localFile := viper.GetString(provider + "_local_file")

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
	logger := GetCiprLogger()
	response, err := http.Get(url)

	if err != nil {
		logger.Println(err.Error())
		os.Exit(1)
	}

	responseData, err := io.ReadAll(response.Body)

	if err != nil {
		logger.Fatal(err)
	}

	return string(responseData), nil
}
