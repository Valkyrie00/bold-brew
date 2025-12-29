package services

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
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

// readCacheFile reads a cached file if it exists and meets minimum size requirements.
// Returns nil if cache should not be used.
func readCacheFile(filename string, minSize int64) []byte {
	cacheFile := filepath.Join(getCacheDir(), filename)
	fileInfo, err := os.Stat(cacheFile)
	if err != nil || fileInfo.Size() < minSize {
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
