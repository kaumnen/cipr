package utils

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestGetRawData_BareRelativePath(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	require.NoError(t, os.WriteFile("ranges.csv", []byte("local"), 0o600))

	got, err := GetRawData(context.Background(), "ranges.csv")
	require.NoError(t, err)
	assert.Equal(t, "local", got)
}

func TestReadAllLimited(t *testing.T) {
	got, err := readAllLimited(bytes.NewBufferString("1234"), 4)
	require.NoError(t, err)
	assert.Equal(t, []byte("1234"), got)

	_, err = readAllLimited(bytes.NewBufferString("12345"), 4)
	require.Error(t, err)
}

func TestLoadFromEndpoint_OK(t *testing.T) {
	body := "1.1.1.0/24\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, UserAgent, r.Header.Get("User-Agent"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	got, err := loadFromEndpoint(context.Background(), srv.URL)
	assert.NoError(t, err)
	assert.Equal(t, body, got)
}

func TestLoadFromEndpoint_Non2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := loadFromEndpoint(context.Background(), srv.URL)
	assert.Error(t, err)
}

func TestLoadFromEndpoint_Unreachable(t *testing.T) {
	_, err := loadFromEndpoint(context.Background(), "http://127.0.0.1:1")
	assert.Error(t, err)
}

func TestGetRawData_URL(t *testing.T) {
	body := "1.1.1.0/24\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	got, err := GetRawData(context.Background(), srv.URL)
	assert.NoError(t, err)
	assert.Equal(t, body, got)
}

func TestGetRawData_LocalPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")
	want := "x"
	if err := os.WriteFile(path, []byte(want), 0o644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	got, err := GetRawData(context.Background(), path)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetRawData_BareFilename(t *testing.T) {
	dir := t.TempDir()
	oldWorkingDir, err := os.Getwd()
	assert.NoError(t, err)
	assert.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(oldWorkingDir) })

	want := "from-bare-filename"
	assert.NoError(t, os.WriteFile("ranges.txt", []byte(want), 0o644))
	got, err := GetRawData(context.Background(), "ranges.txt")
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetRawData_UnsupportedURLScheme(t *testing.T) {
	_, err := GetRawData(context.Background(), "ftp://example.com/ranges.txt")
	assert.ErrorContains(t, err, "unsupported source URL scheme")
}

func TestGetRawData_InvalidURL(t *testing.T) {
	_, err := GetRawData(context.Background(), "https://")
	assert.ErrorContains(t, err, "invalid source URL")
}

func TestGetRawData_UnsupportedConfiguredEndpointScheme(t *testing.T) {
	t.Cleanup(func() { viper.Reset() })
	viper.Set("testprovider_endpoint", "ftp://example.com/ranges.txt")

	_, err := GetRawData(context.Background(), "testprovider")
	assert.ErrorContains(t, err, "unsupported source URL scheme")
}

func TestGetRawData_ViperKey(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Cleanup(func() { viper.Reset() })

	body := "from-viper-key"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	viper.Set("testprovider_endpoint", srv.URL)

	got, err := GetRawData(context.Background(), "testprovider")
	assert.NoError(t, err)
	assert.Equal(t, body, got)
}

func TestGetRawData_NoSource(t *testing.T) {
	t.Cleanup(func() { viper.Reset() })

	_, err := GetRawData(context.Background(), "unknown-key")
	assert.Error(t, err)
}

func TestGetRawData_CacheHitWithinTTL(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Cleanup(func() { viper.Reset() })

	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte("payload"))
	}))
	defer srv.Close()

	viper.Set("cachetest_endpoint", srv.URL)
	viper.Set("cachetest_cache_ttl", "1h")

	for i := 0; i < 2; i++ {
		got, err := GetRawData(context.Background(), "cachetest")
		assert.NoError(t, err)
		assert.Equal(t, "payload", got)
	}
	assert.Equal(t, 1, hits, "second call should hit cache, not server")
}

func TestGetRawData_NoCacheBypass(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Cleanup(func() { viper.Reset() })

	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte("payload"))
	}))
	defer srv.Close()

	viper.Set("cachetest_endpoint", srv.URL)
	viper.Set("no_cache", true)

	for i := 0; i < 2; i++ {
		_, err := GetRawData(context.Background(), "cachetest")
		assert.NoError(t, err)
	}
	assert.Equal(t, 2, hits, "--no-cache should skip cache reads")

	path, err := cachePath("cachetest")
	assert.NoError(t, err)
	_, statErr := os.Stat(path)
	assert.True(t, os.IsNotExist(statErr), "--no-cache should skip cache writes")
}

func TestGetRawData_LocalFileSkipsCache(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)
	t.Cleanup(func() { viper.Reset() })

	dir := t.TempDir()
	path := filepath.Join(dir, "ranges.txt")
	if err := os.WriteFile(path, []byte("local"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	viper.Set("cachetest_local_file", path)

	got, err := GetRawData(context.Background(), "cachetest")
	assert.NoError(t, err)
	assert.Equal(t, "local", got)

	cp, err := cachePath("cachetest")
	assert.NoError(t, err)
	_, statErr := os.Stat(cp)
	assert.True(t, os.IsNotExist(statErr), "local-file source should not write cache")
}

func TestGetRawData_RawURLSkipsCache(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Cleanup(func() { viper.Reset() })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("body"))
	}))
	defer srv.Close()

	_, err := GetRawData(context.Background(), srv.URL)
	assert.NoError(t, err)

	dir, err := cacheDir()
	assert.NoError(t, err)
	entries, _ := os.ReadDir(dir)
	assert.Empty(t, entries, "raw URL source should not write any cache file")
}

func TestGetRawData_TTLZeroAlwaysFetches(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Cleanup(func() { viper.Reset() })

	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte("payload"))
	}))
	defer srv.Close()

	viper.Set("cachetest_endpoint", srv.URL)
	viper.Set("cachetest_cache_ttl", "0s")

	for i := 0; i < 2; i++ {
		_, err := GetRawData(context.Background(), "cachetest")
		assert.NoError(t, err)
	}
	assert.Equal(t, 2, hits, "ttl=0 should refetch every call")
}

func TestGetRawData_BadTTLFallsBack(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Cleanup(func() { viper.Reset() })

	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte("payload"))
	}))
	defer srv.Close()

	viper.Set("cachetest_endpoint", srv.URL)
	viper.Set("cachetest_cache_ttl", "garbage")

	for i := 0; i < 2; i++ {
		_, err := GetRawData(context.Background(), "cachetest")
		assert.NoError(t, err)
	}
	assert.Equal(t, 1, hits, "unparseable TTL should fall back to default and still cache")
}
