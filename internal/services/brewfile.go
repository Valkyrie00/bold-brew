// Package services provides Brewfile support for Bold Brew.
//
// This file handles parsing Brewfile entries (taps, formulae, casks),
// loading packages from third-party taps, and installing missing taps
// at application startup.
//
// NOTE: These methods are only active in Brewfile mode (bbrew -f <file>).
// In normal mode, these functions are not called.
//
// Execution sequence (Brewfile mode only):
//
//  1. Boot() → loadBrewfilePackages()
//     Initial load using cached tap data for fast startup.
//
//  2. BuildApp() → goroutine:
//     a) installBrewfileTapsAtStartup()
//     Installs any missing taps from the Brewfile.
//     b) updateHomeBrew() → forceRefreshResults()
//     Refreshes Homebrew data and reloads packages.
//
//  3. forceRefreshResults() → fetchTapPackages() + loadBrewfilePackages()
//     Fetches fresh tap package info and rebuilds the package list.
package services

import (
	"bbrew/internal/models"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ResolveBrewfilePath resolves a Brewfile path which can be local or a remote URL.
// Returns the local file path and a cleanup function to call when done.
// For local files, cleanup is a no-op. For remote files, cleanup removes the temp file.
func ResolveBrewfilePath(pathOrURL string) (localPath string, cleanup func(), err error) {
	// Check if it's a remote URL (HTTPS only for security)
	if strings.HasPrefix(pathOrURL, "https://") {
		localPath, err = downloadBrewfile(pathOrURL)
		if err != nil {
			return "", nil, err
		}
		// Return cleanup function that removes the temp file
		cleanup = func() { os.Remove(localPath) }
		return localPath, cleanup, nil
	}

	// Local file - validate it exists
	if _, err := os.Stat(pathOrURL); os.IsNotExist(err) {
		return "", nil, fmt.Errorf("brewfile not found: %s", pathOrURL)
	} else if err != nil {
		return "", nil, fmt.Errorf("cannot access Brewfile: %w", err)
	}

	// No cleanup needed for local files
	return pathOrURL, func() {}, nil
}

// downloadBrewfile downloads a remote Brewfile to a temporary file.
func downloadBrewfile(url string) (string, error) {
	fmt.Fprintf(os.Stderr, "Downloading Brewfile from %s...\n", url)

	resp, err := http.Get(url) // #nosec G107 - URL is user-provided, HTTPS enforced
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Create temp file
	tempFile, err := os.CreateTemp(os.TempDir(), "bbrew-remote-*.brewfile")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Copy content
	if _, err = io.Copy(tempFile, resp.Body); err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to save Brewfile: %w", err)
	}

	return filepath.Clean(tempFile.Name()), nil
}

// parseBrewfileWithTaps parses a Brewfile and returns taps and packages separately.
func parseBrewfileWithTaps(filepath string) (*models.BrewfileResult, error) {
	// #nosec G304 -- filepath is user-provided via CLI flag
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Brewfile: %w", err)
	}

	result := &models.BrewfileResult{
		Taps:     []string{},
		Packages: []models.BrewfileEntry{},
	}
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse tap entries: tap "user/repo"
		if strings.HasPrefix(line, "tap ") {
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start != -1 && end != -1 && start < end {
				tapName := line[start+1 : end]
				result.Taps = append(result.Taps, tapName)
			}
		}

		// Parse brew entries: brew "package-name"
		if strings.HasPrefix(line, "brew ") {
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start != -1 && end != -1 && start < end {
				packageName := line[start+1 : end]
				result.Packages = append(result.Packages, models.BrewfileEntry{
					Name:   packageName,
					IsCask: false,
				})
			}
		}

		// Parse cask entries: cask "package-name"
		if strings.HasPrefix(line, "cask ") {
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start != -1 && end != -1 && start < end {
				packageName := line[start+1 : end]
				result.Packages = append(result.Packages, models.BrewfileEntry{
					Name:   packageName,
					IsCask: true,
				})
			}
		}
	}

	return result, nil
}

// loadBrewfilePackages parses the Brewfile and creates a filtered package list.
// Uses the DataProvider to load tap packages from cache or fetch via brew info.
func (s *AppService) loadBrewfilePackages() error {
	result, err := parseBrewfileWithTaps(s.brewfilePath)
	if err != nil {
		return err
	}

	// Store taps for later installation
	s.brewfileTaps = result.Taps

	// Create a map for quick lookup of Brewfile entries
	packageMap := make(map[string]models.PackageType)
	for _, entry := range result.Packages {
		if entry.IsCask {
			packageMap[entry.Name] = models.PackageTypeCask
		} else {
			packageMap[entry.Name] = models.PackageTypeFormula
		}
	}

	// Track which packages were found (to avoid duplicates)
	foundPackages := make(map[string]bool)

	// Get actual installed packages (2 calls total, much faster than per-package checks)
	installedCasks := s.dataProvider.FetchInstalledCaskNames()
	installedFormulae := s.dataProvider.FetchInstalledFormulaNames()

	// Filter packages to only include those in the Brewfile
	*s.brewfilePackages = []models.Package{}
	for _, pkg := range *s.packages {
		if pkgType, exists := packageMap[pkg.Name]; exists && pkgType == pkg.Type {
			// Skip if already added (prevent duplicates)
			if foundPackages[pkg.Name] {
				continue
			}
			// Verify installation status against actual installed lists
			if pkgType == models.PackageTypeCask {
				pkg.LocallyInstalled = installedCasks[pkg.Name]
			} else {
				pkg.LocallyInstalled = installedFormulae[pkg.Name]
			}
			*s.brewfilePackages = append(*s.brewfilePackages, pkg)
			foundPackages[pkg.Name] = true
		}
	}

	// Collect entries not found in main list (tap packages)
	var tapEntries []models.BrewfileEntry
	for _, entry := range result.Packages {
		if !foundPackages[entry.Name] {
			tapEntries = append(tapEntries, entry)
		}
	}

	// Load tap packages from cache (fast startup)
	if len(tapEntries) > 0 {
		// Build existing packages map
		existingPackages := make(map[string]models.Package)
		for _, pkg := range *s.packages {
			existingPackages[pkg.Name] = pkg
		}

		// Use DataProvider to load tap packages (from cache only at startup, no fetch)
		tapPackages, _ := s.dataProvider.GetTapPackages(tapEntries, existingPackages, false)

		// Add tap packages to brewfilePackages, updating installed status (avoid duplicates)
		for _, pkg := range tapPackages {
			if foundPackages[pkg.Name] {
				continue // Already added
			}
			if pkg.Type == models.PackageTypeCask {
				pkg.LocallyInstalled = installedCasks[pkg.Name]
			} else {
				pkg.LocallyInstalled = installedFormulae[pkg.Name]
			}
			*s.brewfilePackages = append(*s.brewfilePackages, pkg)
			foundPackages[pkg.Name] = true
		}
	}

	// Sort by name for consistent display
	sort.Slice(*s.brewfilePackages, func(i, j int) bool {
		return (*s.brewfilePackages)[i].Name < (*s.brewfilePackages)[j].Name
	})

	return nil
}

// fetchTapPackages fetches info for packages from third-party taps and adds them to s.packages.
// This is called after taps are installed so that loadBrewfilePackages can find them.
// Uses the DataProvider to fetch and cache tap package data.
func (s *AppService) fetchTapPackages() {
	if !s.IsBrewfileMode() || len(s.brewfileTaps) == 0 {
		return
	}

	result, err := parseBrewfileWithTaps(s.brewfilePath)
	if err != nil {
		return
	}

	// Build a map of existing packages for quick lookup
	existingPackages := make(map[string]models.Package)
	for _, pkg := range *s.packages {
		existingPackages[pkg.Name] = pkg
	}

	// Use DataProvider to fetch all tap packages (force download to get fresh data)
	tapPackages, _ := s.dataProvider.GetTapPackages(result.Packages, existingPackages, true)

	// Add tap packages to s.packages (avoiding duplicates)
	for _, pkg := range tapPackages {
		if _, exists := existingPackages[pkg.Name]; !exists {
			*s.packages = append(*s.packages, pkg)
		}
	}
}

// installBrewfileTapsAtStartup installs any missing taps from the Brewfile at app startup.
// This runs before updateHomeBrew, which will then reload all data including the new taps.
func (s *AppService) installBrewfileTapsAtStartup() {
	// Check which taps need to be installed
	var tapsToInstall []string
	for _, tap := range s.brewfileTaps {
		if !s.brewService.IsTapInstalled(tap) {
			tapsToInstall = append(tapsToInstall, tap)
		}
	}

	if len(tapsToInstall) == 0 {
		return // All taps already installed
	}

	// Install missing taps
	for _, tap := range tapsToInstall {
		tap := tap // Create local copy for closures
		s.app.QueueUpdateDraw(func() {
			s.layout.GetNotifier().ShowWarning(fmt.Sprintf("Installing tap %s...", tap))
			fmt.Fprintf(s.layout.GetOutput().View(), "[TAP] Installing %s...\n", tap)
		})

		if err := s.brewService.InstallTap(tap, s.app, s.layout.GetOutput().View()); err != nil {
			s.app.QueueUpdateDraw(func() {
				s.layout.GetNotifier().ShowError(fmt.Sprintf("Failed to install tap %s", tap))
				fmt.Fprintf(s.layout.GetOutput().View(), "[ERROR] Failed to install tap %s\n", tap)
			})
		} else {
			s.app.QueueUpdateDraw(func() {
				s.layout.GetNotifier().ShowSuccess(fmt.Sprintf("Tap %s installed", tap))
				fmt.Fprintf(s.layout.GetOutput().View(), "[SUCCESS] tap %s installed\n", tap)
			})
		}
	}

	s.app.QueueUpdateDraw(func() {
		s.layout.GetNotifier().ShowSuccess("All taps installed")
	})
}
