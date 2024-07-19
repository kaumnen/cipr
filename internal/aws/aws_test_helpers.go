package aws

import (
	"os"
	"path/filepath"
)

func mockGetReq() string {
	data, err := os.ReadFile(filepath.Join("..", "testdata", "mock_ip_ranges_response.json"))
	if err != nil {
		panic(err)
	}
	return string(data)
}
