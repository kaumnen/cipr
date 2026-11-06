package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheDir_XDGSet(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", xdg)

	dir, err := cacheDir()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(xdg, "cipr"), dir)
}

func TestCacheDir_XDGUnset(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "")

	dir, err := cacheDir()
	assert.NoError(t, err)
	home, _ := os.UserHomeDir()
	assert.Equal(t, filepath.Join(home, ".cache", "cipr"), dir)
}

func TestReadCache_Miss(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	data, hit, err := readCache("absent", time.Hour)
	assert.NoError(t, err)
	assert.False(t, hit)
	assert.Nil(t, data)
}

func TestWriteAndReadCache_RoundTrip(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	want := []byte("payload")
	assert.NoError(t, writeCache("rt", want))

	got, hit, err := readCache("rt", time.Hour)
	assert.NoError(t, err)
	assert.True(t, hit)
	assert.Equal(t, want, got)
}

func TestReadCache_StaleByMtime(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	assert.NoError(t, writeCache("stale", []byte("payload")))

	path, err := cachePath("stale")
	assert.NoError(t, err)
	old := time.Now().Add(-2 * time.Hour)
	assert.NoError(t, os.Chtimes(path, old, old))

	_, hit, err := readCache("stale", time.Hour)
	assert.NoError(t, err)
	assert.False(t, hit, "entry older than TTL should miss")
}

func TestReadCache_TTLZero(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	assert.NoError(t, writeCache("zero", []byte("payload")))

	_, hit, err := readCache("zero", 0)
	assert.NoError(t, err)
	assert.False(t, hit, "ttl=0 should always miss")
}

func TestWriteCache_CreatesDir(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", root)

	assert.NoError(t, writeCache("nested", []byte("x")))

	info, err := os.Stat(filepath.Join(root, "cipr", "nested.cache"))
	assert.NoError(t, err)
	assert.False(t, info.IsDir())
}

func TestDropCache(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	assert.NoError(t, writeCache("aws", []byte("bad")))
	DropCache("aws")
	_, hit, _ := readCache("aws", time.Hour)
	assert.False(t, hit, "configured source cache should be removed")

	assert.NoError(t, writeCache("ranges.json", []byte("kept")))
	DropCache("ranges.json")
	_, hit, _ = readCache("ranges.json", time.Hour)
	assert.True(t, hit, "unconfigured source must not touch the cache")
}
