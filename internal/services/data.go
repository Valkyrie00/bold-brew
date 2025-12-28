package services

import (
	"bbrew/internal/models"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// SetupData initializes the BrewService by loading installed packages, remote formulae, casks, and analytics data.
func (s *BrewService) SetupData(forceDownload bool) error {
	// Load formulae
	if err := s.loadInstalled(forceDownload); err != nil {
		return fmt.Errorf("failed to load installed formulae: %w", err)
	}

	if err := s.loadRemote(forceDownload); err != nil {
		return fmt.Errorf("failed to load remote formulae: %w", err)
	}

	if err := s.loadAnalytics(forceDownload); err != nil {
		return fmt.Errorf("failed to load formulae analytics: %w", err)
	}

	// Load casks
	if err := s.loadInstalledCasks(forceDownload); err != nil {
		return fmt.Errorf("failed to load installed casks: %w", err)
	}

	if err := s.loadRemoteCasks(forceDownload); err != nil {
		return fmt.Errorf("failed to load remote casks: %w", err)
	}

	if err := s.loadCaskAnalytics(forceDownload); err != nil {
		return fmt.Errorf("failed to load cask analytics: %w", err)
	}

	return nil
}

// loadInstalled retrieves installed formulae, optionally using cache.
func (s *BrewService) loadInstalled(forceDownload bool) error {
	if err := ensureCacheDir(); err != nil {
		return err
	}

	const cacheFile = "installed.json"
	if !forceDownload {
		if data := readCacheFile(cacheFile, 10); data != nil {
			*s.installed = make([]models.Formula, 0)
			if err := json.Unmarshal(data, &s.installed); err == nil {
				s.markFormulaeAsInstalled()
				return nil
			}
		}
	}

	cmd := exec.Command("brew", "info", "--json=v1", "--installed")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	*s.installed = make([]models.Formula, 0)
	if err := json.Unmarshal(output, &s.installed); err != nil {
		return err
	}

	s.markFormulaeAsInstalled()
	writeCacheFile(cacheFile, output)
	return nil
}

// markFormulaeAsInstalled sets LocallyInstalled and LocalPath for all installed formulae.
func (s *BrewService) markFormulaeAsInstalled() {
	prefix := s.GetPrefixPath()
	for i := range *s.installed {
		(*s.installed)[i].LocallyInstalled = true
		(*s.installed)[i].LocalPath = filepath.Join(prefix, "Cellar", (*s.installed)[i].Name)
	}
}

// loadInstalledCasks retrieves installed casks, optionally using cache.
func (s *BrewService) loadInstalledCasks(forceDownload bool) error {
	if err := ensureCacheDir(); err != nil {
		return err
	}

	const cacheFile = "installed-casks.json"
	if !forceDownload {
		if data := readCacheFile(cacheFile, 10); data != nil {
			var response struct {
				Casks []models.Cask `json:"casks"`
			}
			if err := json.Unmarshal(data, &response); err == nil {
				*s.installedCasks = response.Casks
				s.markCasksAsInstalled()
				return nil
			}
		}
	}

	// Get list of installed cask names
	listCmd := exec.Command("brew", "list", "--cask")
	listOutput, err := listCmd.Output()
	if err != nil {
		*s.installedCasks = make([]models.Cask, 0)
		return nil
	}

	caskNames := strings.Split(strings.TrimSpace(string(listOutput)), "\n")
	if len(caskNames) == 0 || (len(caskNames) == 1 && caskNames[0] == "") {
		*s.installedCasks = make([]models.Cask, 0)
		return nil
	}

	// Get info for each installed cask
	args := append([]string{"info", "--json=v2", "--cask"}, caskNames...)
	infoCmd := exec.Command("brew", args...)
	infoOutput, err := infoCmd.Output()
	if err != nil {
		*s.installedCasks = make([]models.Cask, 0)
		return nil
	}

	var response struct {
		Casks []models.Cask `json:"casks"`
	}
	if err := json.Unmarshal(infoOutput, &response); err != nil {
		return err
	}

	*s.installedCasks = response.Casks
	s.markCasksAsInstalled()
	writeCacheFile(cacheFile, infoOutput)
	return nil
}

// markCasksAsInstalled sets LocallyInstalled and IsCask for all installed casks.
func (s *BrewService) markCasksAsInstalled() {
	for i := range *s.installedCasks {
		(*s.installedCasks)[i].LocallyInstalled = true
		(*s.installedCasks)[i].IsCask = true
	}
}

// loadRemote retrieves the list of remote Homebrew formulae from the API and caches them locally.
func (s *BrewService) loadRemote(forceDownload bool) error {
	if err := ensureCacheDir(); err != nil {
		return err
	}

	const cacheFile = "formula.json"
	if !forceDownload {
		if data := readCacheFile(cacheFile, 1000); data != nil {
			*s.remote = make([]models.Formula, 0)
			if err := json.Unmarshal(data, &s.remote); err == nil && len(*s.remote) > 0 {
				return nil
			}
		}
	}

	body, err := fetchFromAPI(FormulaeAPIURL)
	if err != nil {
		return err
	}

	*s.remote = make([]models.Formula, 0)
	if err := json.Unmarshal(body, s.remote); err != nil {
		return err
	}

	writeCacheFile(cacheFile, body)
	return nil
}

// loadRemoteCasks retrieves the list of remote Homebrew casks from the API and caches them locally.
func (s *BrewService) loadRemoteCasks(forceDownload bool) error {
	if err := ensureCacheDir(); err != nil {
		return err
	}

	const cacheFile = "cask.json"
	if !forceDownload {
		if data := readCacheFile(cacheFile, 1000); data != nil {
			*s.remoteCasks = make([]models.Cask, 0)
			if err := json.Unmarshal(data, &s.remoteCasks); err == nil && len(*s.remoteCasks) > 0 {
				return nil
			}
		}
	}

	body, err := fetchFromAPI(CaskAPIURL)
	if err != nil {
		return err
	}

	*s.remoteCasks = make([]models.Cask, 0)
	if err := json.Unmarshal(body, s.remoteCasks); err != nil {
		return err
	}

	writeCacheFile(cacheFile, body)
	return nil
}

// loadAnalytics retrieves the analytics data for Homebrew formulae from the API and caches them locally.
func (s *BrewService) loadAnalytics(forceDownload bool) error {
	if err := ensureCacheDir(); err != nil {
		return err
	}

	const cacheFile = "analytics.json"
	if !forceDownload {
		if data := readCacheFile(cacheFile, 100); data != nil {
			analytics := models.Analytics{}
			if err := json.Unmarshal(data, &analytics); err == nil && len(analytics.Items) > 0 {
				s.analytics = make(map[string]models.AnalyticsItem)
				for _, f := range analytics.Items {
					s.analytics[f.Formula] = f
				}
				return nil
			}
		}
	}

	body, err := fetchFromAPI(AnalyticsAPIURL)
	if err != nil {
		return err
	}

	analytics := models.Analytics{}
	if err := json.Unmarshal(body, &analytics); err != nil {
		return err
	}

	s.analytics = make(map[string]models.AnalyticsItem)
	for _, f := range analytics.Items {
		s.analytics[f.Formula] = f
	}

	writeCacheFile(cacheFile, body)
	return nil
}

// loadCaskAnalytics retrieves the analytics data for Homebrew casks from the API and caches them locally.
func (s *BrewService) loadCaskAnalytics(forceDownload bool) error {
	if err := ensureCacheDir(); err != nil {
		return err
	}

	const cacheFile = "cask-analytics.json"
	if !forceDownload {
		if data := readCacheFile(cacheFile, 100); data != nil {
			analytics := models.Analytics{}
			if err := json.Unmarshal(data, &analytics); err == nil && len(analytics.Items) > 0 {
				s.caskAnalytics = make(map[string]models.AnalyticsItem)
				for _, c := range analytics.Items {
					if c.Cask != "" {
						s.caskAnalytics[c.Cask] = c
					}
				}
				return nil
			}
		}
	}

	body, err := fetchFromAPI(CaskAnalyticsAPIURL)
	if err != nil {
		return err
	}

	analytics := models.Analytics{}
	if err := json.Unmarshal(body, &analytics); err != nil {
		return err
	}

	s.caskAnalytics = make(map[string]models.AnalyticsItem)
	for _, c := range analytics.Items {
		if c.Cask != "" {
			s.caskAnalytics[c.Cask] = c
		}
	}

	writeCacheFile(cacheFile, body)
	return nil
}

// LoadTapPackagesCache loads cached tap packages from disk.
func (s *BrewService) LoadTapPackagesCache() map[string]models.Package {
	result := make(map[string]models.Package)

	const cacheFile = "tap_packages.json"
	if data := readCacheFile(cacheFile, 10); data != nil {
		var packages []models.Package
		if err := json.Unmarshal(data, &packages); err == nil {
			for _, pkg := range packages {
				result[pkg.Name] = pkg
			}
		}
	}

	return result
}

// SaveTapPackagesToCache saves tap packages to disk cache.
func (s *BrewService) SaveTapPackagesToCache(packages []models.Package) error {
	if err := ensureCacheDir(); err != nil {
		return err
	}

	data, err := json.Marshal(packages)
	if err != nil {
		return err
	}

	writeCacheFile("tap_packages.json", data)
	return nil
}
