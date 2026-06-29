package services

import (
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
)

// Default cache TTL values
const (
	cacheDefaultTTL = 24 * time.Hour // Remote API data (formulae, casks, analytics)
	cacheShortTTL   = 1 * time.Hour  // Installed package data (refreshed more frequently)
)

// getCacheDir returns the cache directory following XDG Base Directory Specification.
func getCacheDir() string {
	return filepath.Join(xdg.CacheHome, "bbrew")
}

// ensureCacheDir creates the cache directory if it doesn't exist.
func ensureCacheDir() error {
	cacheDir := getCacheDir()
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return os.MkdirAll(cacheDir, 0750)
	}
	return nil
}

// readCacheFile reads a cached file if it exists, meets minimum size requirements,
// and is not older than the specified TTL. Returns nil if cache should not be used.
func readCacheFile(filename string, minSize int64) []byte {
	return readCacheFileWithTTL(filename, minSize, cacheDefaultTTL)
}

// readCacheFileWithTTL reads a cached file with a custom TTL.
func readCacheFileWithTTL(filename string, minSize int64, ttl time.Duration) []byte {
	cacheFile := filepath.Join(getCacheDir(), filename)
	fileInfo, err := os.Stat(cacheFile)
	if err != nil || fileInfo.Size() < minSize {
		return nil
	}
	if time.Since(fileInfo.ModTime()) > ttl {
		return nil
	}
	// #nosec G304 -- cacheFile path is safely constructed from getCacheDir
	data, err := os.ReadFile(cacheFile)
	if err != nil || len(data) == 0 {
		return nil
	}
	return data
}

// writeCacheFile saves data to a cache file.
func writeCacheFile(filename string, data []byte) {
	cacheFile := filepath.Join(getCacheDir(), filename)
	_ = os.WriteFile(cacheFile, data, 0600)
}
