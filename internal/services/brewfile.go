package services

import (
	"bbrew/internal/models"
	"fmt"
	"sort"
)

// loadBrewfilePackages parses the Brewfile and creates a filtered package list.
// Packages not found in s.packages are loaded from cache, or show "(loading...)" if not cached.
func (s *AppService) loadBrewfilePackages() error {
	result, err := s.brewService.ParseBrewfileWithTaps(s.brewfilePath)
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

	// Track which packages were found in the main package list
	foundPackages := make(map[string]bool)

	// Get actual installed packages (2 calls total, much faster than per-package checks)
	installedCasks := s.brewService.GetInstalledCaskNames()
	installedFormulae := s.brewService.GetInstalledFormulaNames()

	// Filter packages to only include those in the Brewfile
	*s.brewfilePackages = []models.Package{}
	for _, pkg := range *s.packages {
		if pkgType, exists := packageMap[pkg.Name]; exists && pkgType == pkg.Type {
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

	// Load tap packages cache for packages not found in main list
	tapCache := s.brewService.LoadTapPackagesCache()

	// For packages not found, try cache first, then show "(loading...)"
	for _, entry := range result.Packages {
		if foundPackages[entry.Name] {
			continue
		}

		// Try to get from cache
		if cachedPkg, exists := tapCache[entry.Name]; exists {
			// Update installation status from local system
			if entry.IsCask {
				cachedPkg.LocallyInstalled = s.brewService.IsPackageInstalled(entry.Name, true)
			} else {
				cachedPkg.LocallyInstalled = s.brewService.IsPackageInstalled(entry.Name, false)
			}
			*s.brewfilePackages = append(*s.brewfilePackages, cachedPkg)
			continue
		}

		// Not in cache - show placeholder
		pkgType := models.PackageTypeFormula
		if entry.IsCask {
			pkgType = models.PackageTypeCask
		}
		*s.brewfilePackages = append(*s.brewfilePackages, models.Package{
			Name:        entry.Name,
			DisplayName: entry.Name,
			Description: "(loading...)",
			Type:        pkgType,
		})
	}

	// Sort by name for consistent display
	sort.Slice(*s.brewfilePackages, func(i, j int) bool {
		return (*s.brewfilePackages)[i].Name < (*s.brewfilePackages)[j].Name
	})

	return nil
}

// fetchTapPackages fetches info for packages from third-party taps and adds them to s.packages.
// This is called after taps are installed so that loadBrewfilePackages can find them.
// It also saves the fetched data to cache for faster startup next time.
func (s *AppService) fetchTapPackages() {
	if !s.IsBrewfileMode() || len(s.brewfileTaps) == 0 {
		return
	}

	result, err := s.brewService.ParseBrewfileWithTaps(s.brewfilePath)
	if err != nil {
		return
	}

	// Build a map of existing packages for quick lookup
	existingPackages := make(map[string]models.Package)
	for _, pkg := range *s.packages {
		existingPackages[pkg.Name] = pkg
	}

	// Collect packages not in s.packages (need to fetch) and packages already present (for cache)
	var missingCasks []string
	var missingFormulae []string
	var presentPackages []models.Package // Packages already in s.packages (installed tap packages)

	for _, entry := range result.Packages {
		if pkg, exists := existingPackages[entry.Name]; exists {
			// Package is already in s.packages (likely installed)
			// Save it to cache so it's available after uninstall
			presentPackages = append(presentPackages, pkg)
		} else {
			// Package is missing, need to fetch
			if entry.IsCask {
				missingCasks = append(missingCasks, entry.Name)
			} else {
				missingFormulae = append(missingFormulae, entry.Name)
			}
		}
	}

	// Collect all tap packages to save to cache (both fetched and already present)
	tapPackages := append([]models.Package{}, presentPackages...)

	// Fetch and add missing casks
	if len(missingCasks) > 0 {
		caskInfo := s.brewService.GetPackagesInfo(missingCasks, true)
		for _, name := range missingCasks {
			if pkg, exists := caskInfo[name]; exists {
				*s.packages = append(*s.packages, pkg)
				tapPackages = append(tapPackages, pkg)
			} else {
				// Add fallback entry if brew info failed
				fallback := models.Package{
					Name:        name,
					DisplayName: name,
					Description: "(unable to load package info)",
					Type:        models.PackageTypeCask,
				}
				*s.packages = append(*s.packages, fallback)
				tapPackages = append(tapPackages, fallback)
			}
		}
	}

	// Fetch and add missing formulae
	if len(missingFormulae) > 0 {
		formulaInfo := s.brewService.GetPackagesInfo(missingFormulae, false)
		for _, name := range missingFormulae {
			if pkg, exists := formulaInfo[name]; exists {
				*s.packages = append(*s.packages, pkg)
				tapPackages = append(tapPackages, pkg)
			} else {
				// Add fallback entry if brew info failed
				fallback := models.Package{
					Name:        name,
					DisplayName: name,
					Description: "(unable to load package info)",
					Type:        models.PackageTypeFormula,
				}
				*s.packages = append(*s.packages, fallback)
				tapPackages = append(tapPackages, fallback)
			}
		}
	}

	// Save ALL tap packages to cache (including already installed ones)
	if len(tapPackages) > 0 {
		_ = s.brewService.SaveTapPackagesToCache(tapPackages)
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

