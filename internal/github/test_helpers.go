package github

import (
	"os"
	"path/filepath"
)

func mockGetReq() string {
	data, err := os.ReadFile(filepath.Join("..", "testdata", "github_meta_sample.json"))
	if err != nil {
		panic(err)
	}
	return string(data)
}
