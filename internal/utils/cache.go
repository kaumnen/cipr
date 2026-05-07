package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

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
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
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
