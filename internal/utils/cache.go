package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// GetCached wraps fetch with the standard cache-aside policy: read on hit,
// run fetch on miss, write the result. If key is "" or --no-cache is set,
// fetch runs unconditionally and the cache is left untouched. Cache write
// failures are logged but never fatal. Suitable for providers (e.g. azure)
// whose fetch path doesn't reduce to a single utils.GetRawData call.
func GetCached(ctx context.Context, key string, fetch func(context.Context) (string, error)) (string, error) {
	if key == "" {
		Debugf("cache: bypassed for unkeyed source")
		return fetch(ctx)
	}
	if viper.GetBool("no_cache") {
		Debugf("cache: bypassed for %s by configuration", key)
		return fetch(ctx)
	}
	ttl := resolveCacheTTL(key)
	if ttl > 0 {
		if data, ok, _ := readCache(key, ttl); ok {
			Debugf("cache: hit for %s (age %s, ttl %s)", key, cacheAge(key), ttl)
			fmt.Fprintf(os.Stderr, "Using cached IP ranges for %s (age %s)\n", key, cacheAge(key))
			return string(data), nil
		}
		Debugf("cache: miss or stale entry for %s (ttl %s)", key, ttl)
	} else {
		Debugf("cache: disabled for %s", key)
	}
	body, err := fetch(ctx)
	if err != nil {
		return "", err
	}
	if werr := writeCache(key, []byte(body)); werr != nil {
		fmt.Fprintln(os.Stderr, "Warning: cache write failed:", werr)
		Debugf("cache: write failed for %s: %v", key, werr)
	} else {
		Debugf("cache: wrote %d bytes for %s", len(body), key)
	}
	return body, nil
}

func resolveCacheTTL(key string) time.Duration {
	raw := viper.GetString(key + "_cache_ttl")
	if raw == "" {
		return defaultCacheTTL
	}
	ttl, err := time.ParseDuration(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: invalid %s_cache_ttl=%q, using %s\n", key, raw, defaultCacheTTL)
		return defaultCacheTTL
	}
	return ttl
}

func cacheDir() (string, error) {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, "cipr"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve cache dir: %w", err)
	}
	return filepath.Join(home, ".cache", "cipr"), nil
}

func cachePath(key string) (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, key+".cache"), nil
}

// readCache returns (data, true, nil) on a fresh hit, (nil, false, nil) on
// miss / stale / unreadable / corrupt — callers should fetch in that case.
// Errors are returned only for genuinely unexpected conditions.
func readCache(key string, ttl time.Duration) ([]byte, bool, error) {
	path, err := cachePath(key)
	if err != nil {
		return nil, false, err
	}
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, nil
	}
	if time.Since(info.ModTime()) >= ttl {
		return nil, false, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false, nil
	}
	return data, true, nil
}

// writeCache writes data atomically (tmp + rename). Caller logs and
// continues on failure — never fatal.
func writeCache(key string, data []byte) error {
	path, err := cachePath(key)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create cache temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()
	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close cache: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("finalize cache: %w", err)
	}
	return nil
}

// cacheAge returns how old the cache entry for key is, or 0 if absent.
// Used only for log messages; never returns an error.
func cacheAge(key string) time.Duration {
	path, err := cachePath(key)
	if err != nil {
		return 0
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return time.Since(info.ModTime()).Round(time.Second)
}
