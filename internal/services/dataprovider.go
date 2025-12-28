package services

import (
	"bbrew/internal/models"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
)

// Cache file names
const (
	cacheFileInstalled      = "installed.json"
	cacheFileInstalledCasks = "installed-casks.json"
	cacheFileFormulae       = "formula.json"
	cacheFileCasks          = "cask.json"
	cacheFileAnalytics      = "analytics.json"
	cacheFileCaskAnalytics  = "cask-analytics.json"
	cacheFileTapPackages    = "tap_packages.json"
)

// DataProviderInterface defines the contract for data loading operations.
type DataProviderInterface interface {
	// Formulae
	LoadInstalledFormulae(forceDownload bool) ([]models.Formula, error)
	LoadRemoteFormulae(forceDownload bool) ([]models.Formula, error)
	LoadFormulaeAnalytics(forceDownload bool) (map[string]models.AnalyticsItem, error)

	// Casks
	LoadInstalledCasks(forceDownload bool) ([]models.Cask, error)
	LoadRemoteCasks(forceDownload bool) ([]models.Cask, error)
	LoadCaskAnalytics(forceDownload bool) (map[string]models.AnalyticsItem, error)

	// Tap packages cache
	LoadTapPackagesCache() map[string]models.Package
	SaveTapPackagesToCache(packages []models.Package) error
}

// DataProvider implements DataProviderInterface.
type DataProvider struct {
	prefixPath string
}

// NewDataProvider creates a new DataProvider instance.
func NewDataProvider() *DataProvider {
	return &DataProvider{}
}

// getPrefixPath returns the Homebrew prefix path, caching it.
func (d *DataProvider) getPrefixPath() string {
	if d.prefixPath != "" {
		return d.prefixPath
	}
	cmd := exec.Command("brew", "--prefix")
	output, err := cmd.Output()
	if err != nil {
		d.prefixPath = "Unknown"
		return d.prefixPath
	}
	d.prefixPath = strings.TrimSpace(string(output))
	return d.prefixPath
}

// LoadInstalledFormulae retrieves installed formulae, optionally using cache.
func (d *DataProvider) LoadInstalledFormulae(forceDownload bool) ([]models.Formula, error) {
	if err := ensureCacheDir(); err != nil {
		return nil, err
	}

	if !forceDownload {
		if data := readCacheFile(cacheFileInstalled, 10); data != nil {
			var formulae []models.Formula
			if err := json.Unmarshal(data, &formulae); err == nil {
				d.markFormulaeAsInstalled(&formulae)
				return formulae, nil
			}
		}
	}

	cmd := exec.Command("brew", "info", "--json=v1", "--installed")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var formulae []models.Formula
	if err := json.Unmarshal(output, &formulae); err != nil {
		return nil, err
	}

	d.markFormulaeAsInstalled(&formulae)
	writeCacheFile(cacheFileInstalled, output)
	return formulae, nil
}

// markFormulaeAsInstalled sets LocallyInstalled and LocalPath for formulae.
func (d *DataProvider) markFormulaeAsInstalled(formulae *[]models.Formula) {
	prefix := d.getPrefixPath()
	for i := range *formulae {
		(*formulae)[i].LocallyInstalled = true
		(*formulae)[i].LocalPath = filepath.Join(prefix, "Cellar", (*formulae)[i].Name)
	}
}

// LoadInstalledCasks retrieves installed casks, optionally using cache.
func (d *DataProvider) LoadInstalledCasks(forceDownload bool) ([]models.Cask, error) {
	if err := ensureCacheDir(); err != nil {
		return nil, err
	}

	if !forceDownload {
		if data := readCacheFile(cacheFileInstalledCasks, 10); data != nil {
			var response struct {
				Casks []models.Cask `json:"casks"`
			}
			if err := json.Unmarshal(data, &response); err == nil {
				d.markCasksAsInstalled(&response.Casks)
				return response.Casks, nil
			}
		}
	}

	// Get list of installed cask names
	listCmd := exec.Command("brew", "list", "--cask")
	listOutput, err := listCmd.Output()
	if err != nil {
		return []models.Cask{}, nil // No casks installed
	}

	caskNames := strings.Split(strings.TrimSpace(string(listOutput)), "\n")
	if len(caskNames) == 0 || (len(caskNames) == 1 && caskNames[0] == "") {
		return []models.Cask{}, nil
	}

	// Get info for each installed cask
	args := append([]string{"info", "--json=v2", "--cask"}, caskNames...)
	infoCmd := exec.Command("brew", args...)
	infoOutput, err := infoCmd.Output()
	if err != nil {
		return []models.Cask{}, nil
	}

	var response struct {
		Casks []models.Cask `json:"casks"`
	}
	if err := json.Unmarshal(infoOutput, &response); err != nil {
		return nil, err
	}

	d.markCasksAsInstalled(&response.Casks)
	writeCacheFile(cacheFileInstalledCasks, infoOutput)
	return response.Casks, nil
}

// markCasksAsInstalled sets LocallyInstalled and IsCask for casks.
func (d *DataProvider) markCasksAsInstalled(casks *[]models.Cask) {
	for i := range *casks {
		(*casks)[i].LocallyInstalled = true
		(*casks)[i].IsCask = true
	}
}

// LoadRemoteFormulae retrieves remote formulae from API, optionally using cache.
func (d *DataProvider) LoadRemoteFormulae(forceDownload bool) ([]models.Formula, error) {
	if err := ensureCacheDir(); err != nil {
		return nil, err
	}

	if !forceDownload {
		if data := readCacheFile(cacheFileFormulae, 1000); data != nil {
			var formulae []models.Formula
			if err := json.Unmarshal(data, &formulae); err == nil && len(formulae) > 0 {
				return formulae, nil
			}
		}
	}

	body, err := fetchFromAPI(FormulaeAPIURL)
	if err != nil {
		return nil, err
	}

	var formulae []models.Formula
	if err := json.Unmarshal(body, &formulae); err != nil {
		return nil, err
	}

	writeCacheFile(cacheFileFormulae, body)
	return formulae, nil
}

// LoadRemoteCasks retrieves remote casks from API, optionally using cache.
func (d *DataProvider) LoadRemoteCasks(forceDownload bool) ([]models.Cask, error) {
	if err := ensureCacheDir(); err != nil {
		return nil, err
	}

	if !forceDownload {
		if data := readCacheFile(cacheFileCasks, 1000); data != nil {
			var casks []models.Cask
			if err := json.Unmarshal(data, &casks); err == nil && len(casks) > 0 {
				return casks, nil
			}
		}
	}

	body, err := fetchFromAPI(CaskAPIURL)
	if err != nil {
		return nil, err
	}

	var casks []models.Cask
	if err := json.Unmarshal(body, &casks); err != nil {
		return nil, err
	}

	writeCacheFile(cacheFileCasks, body)
	return casks, nil
}

// LoadFormulaeAnalytics retrieves formulae analytics from API, optionally using cache.
func (d *DataProvider) LoadFormulaeAnalytics(forceDownload bool) (map[string]models.AnalyticsItem, error) {
	if err := ensureCacheDir(); err != nil {
		return nil, err
	}

	if !forceDownload {
		if data := readCacheFile(cacheFileAnalytics, 100); data != nil {
			analytics := models.Analytics{}
			if err := json.Unmarshal(data, &analytics); err == nil && len(analytics.Items) > 0 {
				result := make(map[string]models.AnalyticsItem)
				for _, f := range analytics.Items {
					result[f.Formula] = f
				}
				return result, nil
			}
		}
	}

	body, err := fetchFromAPI(AnalyticsAPIURL)
	if err != nil {
		return nil, err
	}

	analytics := models.Analytics{}
	if err := json.Unmarshal(body, &analytics); err != nil {
		return nil, err
	}

	result := make(map[string]models.AnalyticsItem)
	for _, f := range analytics.Items {
		result[f.Formula] = f
	}

	writeCacheFile(cacheFileAnalytics, body)
	return result, nil
}

// LoadCaskAnalytics retrieves cask analytics from API, optionally using cache.
func (d *DataProvider) LoadCaskAnalytics(forceDownload bool) (map[string]models.AnalyticsItem, error) {
	if err := ensureCacheDir(); err != nil {
		return nil, err
	}

	if !forceDownload {
		if data := readCacheFile(cacheFileCaskAnalytics, 100); data != nil {
			analytics := models.Analytics{}
			if err := json.Unmarshal(data, &analytics); err == nil && len(analytics.Items) > 0 {
				result := make(map[string]models.AnalyticsItem)
				for _, c := range analytics.Items {
					if c.Cask != "" {
						result[c.Cask] = c
					}
				}
				return result, nil
			}
		}
	}

	body, err := fetchFromAPI(CaskAnalyticsAPIURL)
	if err != nil {
		return nil, err
	}

	analytics := models.Analytics{}
	if err := json.Unmarshal(body, &analytics); err != nil {
		return nil, err
	}

	result := make(map[string]models.AnalyticsItem)
	for _, c := range analytics.Items {
		if c.Cask != "" {
			result[c.Cask] = c
		}
	}

	writeCacheFile(cacheFileCaskAnalytics, body)
	return result, nil
}

// LoadTapPackagesCache loads cached tap packages from disk.
func (d *DataProvider) LoadTapPackagesCache() map[string]models.Package {
	result := make(map[string]models.Package)

	if data := readCacheFile(cacheFileTapPackages, 10); data != nil {
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
func (d *DataProvider) SaveTapPackagesToCache(packages []models.Package) error {
	if err := ensureCacheDir(); err != nil {
		return err
	}

	data, err := json.Marshal(packages)
	if err != nil {
		return err
	}

	writeCacheFile(cacheFileTapPackages, data)
	return nil
}
