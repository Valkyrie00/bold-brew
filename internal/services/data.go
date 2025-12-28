package services

import (
	"bbrew/internal/models"
	"fmt"
)

// SetupData initializes the BrewService by loading installed packages, remote formulae, casks, and analytics data.
// Uses the DataProvider for all data retrieval operations.
func (s *BrewService) SetupData(forceDownload bool) error {
	// Load installed formulae
	installed, err := s.dataProvider.LoadInstalledFormulae(forceDownload)
	if err != nil {
		return fmt.Errorf("failed to load installed formulae: %w", err)
	}
	*s.installed = installed

	// Load remote formulae
	remote, err := s.dataProvider.LoadRemoteFormulae(forceDownload)
	if err != nil {
		return fmt.Errorf("failed to load remote formulae: %w", err)
	}
	*s.remote = remote

	// Load formulae analytics
	analytics, err := s.dataProvider.LoadFormulaeAnalytics(forceDownload)
	if err != nil {
		return fmt.Errorf("failed to load formulae analytics: %w", err)
	}
	s.analytics = analytics

	// Load installed casks
	installedCasks, err := s.dataProvider.LoadInstalledCasks(forceDownload)
	if err != nil {
		return fmt.Errorf("failed to load installed casks: %w", err)
	}
	*s.installedCasks = installedCasks

	// Load remote casks
	remoteCasks, err := s.dataProvider.LoadRemoteCasks(forceDownload)
	if err != nil {
		return fmt.Errorf("failed to load remote casks: %w", err)
	}
	*s.remoteCasks = remoteCasks

	// Load cask analytics
	caskAnalytics, err := s.dataProvider.LoadCaskAnalytics(forceDownload)
	if err != nil {
		return fmt.Errorf("failed to load cask analytics: %w", err)
	}
	s.caskAnalytics = caskAnalytics

	return nil
}

// LoadTapPackagesCache loads cached tap packages from disk.
// Delegates to the DataProvider.
func (s *BrewService) LoadTapPackagesCache() map[string]models.Package {
	return s.dataProvider.LoadTapPackagesCache()
}

// SaveTapPackagesToCache saves tap packages to disk cache.
// Delegates to the DataProvider.
func (s *BrewService) SaveTapPackagesToCache(packages []models.Package) error {
	return s.dataProvider.SaveTapPackagesToCache(packages)
}
