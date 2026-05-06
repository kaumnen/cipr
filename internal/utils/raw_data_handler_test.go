package utils

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")
	want := "1.1.1.0/24\n2.2.2.0/24\n"
	if err := os.WriteFile(path, []byte(want), 0o644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	got, err := loadFromFile(path)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestLoadFromFile_Missing(t *testing.T) {
	_, err := loadFromFile(filepath.Join(t.TempDir(), "does-not-exist"))
	assert.Error(t, err)
}

func TestLoadFromEndpoint_OK(t *testing.T) {
	body := "1.1.1.0/24\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	got, err := loadFromEndpoint(srv.URL)
	assert.NoError(t, err)
	assert.Equal(t, body, got)
}

func TestLoadFromEndpoint_Non2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := loadFromEndpoint(srv.URL)
	assert.Error(t, err)
}

func TestLoadFromEndpoint_Unreachable(t *testing.T) {
	_, err := loadFromEndpoint("http://127.0.0.1:1")
	assert.Error(t, err)
}
