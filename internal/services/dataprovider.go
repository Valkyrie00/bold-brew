package services

import (
	"bbrew/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// API URLs for Homebrew data
const (
	formulaeAPIURL      = "https://formulae.brew.sh/api/formula.json"
	caskAPIURL          = "https://formulae.brew.sh/api/cask.json"
	analyticsAPIURL     = "https://formulae.brew.sh/api/analytics/install-on-request/90d.json"
	caskAnalyticsAPIURL = "https://formulae.brew.sh/api/analytics/cask-install/90d.json"
)

// Cache file names
const (
	cacheFileInstalled      = "installed.json"
	cacheFileInstalledCasks = "installed-casks.json"
	cacheFileFormulae       = "formula.json"
	cacheFileCasks          = "cask.json"
	cacheFileAnalytics      = "analytics.json"
	cacheFileCaskAnalytics  = "cask-analytics.json"
	cacheFileTapPackages    = "tap-packages.json"
)

// DataProviderInterface defines the contract for data operations.
// DataProvider is the central repository for all Homebrew package data.
type DataProviderInterface interface {
	// Setup and retrieval
	SetupData(forceRefresh bool) error
	GetPackages() *[]models.Package

	// Installation status checks (runs brew list command)
	FetchInstalledCaskNames() map[string]bool
	FetchInstalledFormulaNames() map[string]bool

	// Tap packages - gets from cache or fetches via brew info
	GetTapPackages(entries []models.BrewfileEntry, existingPackages map[string]models.Package, forceRefresh bool) ([]models.Package, error)
}

// DataProvider implements DataProviderInterface.
// It is the central repository for all Homebrew package data.
type DataProvider struct {
	// Formula lists
	installedFormulae *[]models.Formula
	remoteFormulae    *[]models.Formula
	formulaeAnalytics map[string]models.AnalyticsItem

	// Cask lists
	installedCasks *[]models.Cask
	remoteCasks    *[]models.Cask
	caskAnalytics  map[string]models.AnalyticsItem

	// Unified package list
	allPackages *[]models.Package

	prefixPath string
}

// NewDataProvider creates a new DataProvider instance with initialized data structures.
func NewDataProvider() *DataProvider {
	return &DataProvider{
		installedFormulae: new([]models.Formula),
		remoteFormulae:    new([]models.Formula),
		installedCasks:    new([]models.Cask),
		remoteCasks:       new([]models.Cask),
		allPackages:       new([]models.Package),
	}
}

// fetchFromAPI downloads data from a URL.
func fetchFromAPI(url string) ([]byte, error) {
	resp, err := http.Get(url) // #nosec G107 - URLs are internal constants
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
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

// GetInstalledFormulae retrieves installed formulae, optionally using cache.
func (d *DataProvider) GetInstalledFormulae(forceRefresh bool) ([]models.Formula, error) {
	if err := ensureCacheDir(); err != nil {
		return nil, err
	}

	if !forceRefresh {
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

// GetInstalledCasks retrieves installed casks, optionally using cache.
func (d *DataProvider) GetInstalledCasks(forceRefresh bool) ([]models.Cask, error) {
	if err := ensureCacheDir(); err != nil {
		return nil, err
	}

	if !forceRefresh {
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

// GetRemoteFormulae retrieves remote formulae from API, optionally using cache.
func (d *DataProvider) GetRemoteFormulae(forceRefresh bool) ([]models.Formula, error) {
	if err := ensureCacheDir(); err != nil {
		return nil, err
	}

	if !forceRefresh {
		if data := readCacheFile(cacheFileFormulae, 1000); data != nil {
			var formulae []models.Formula
			if err := json.Unmarshal(data, &formulae); err == nil && len(formulae) > 0 {
				return formulae, nil
			}
		}
	}

	body, err := fetchFromAPI(formulaeAPIURL)
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

// GetRemoteCasks retrieves remote casks from API, optionally using cache.
func (d *DataProvider) GetRemoteCasks(forceRefresh bool) ([]models.Cask, error) {
	if err := ensureCacheDir(); err != nil {
		return nil, err
	}

	if !forceRefresh {
		if data := readCacheFile(cacheFileCasks, 1000); data != nil {
			var casks []models.Cask
			if err := json.Unmarshal(data, &casks); err == nil && len(casks) > 0 {
				return casks, nil
			}
		}
	}

	body, err := fetchFromAPI(caskAPIURL)
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

// GetFormulaeAnalytics retrieves formulae analytics from API, optionally using cache.
func (d *DataProvider) GetFormulaeAnalytics(forceRefresh bool) (map[string]models.AnalyticsItem, error) {
	if err := ensureCacheDir(); err != nil {
		return nil, err
	}

	if !forceRefresh {
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

	body, err := fetchFromAPI(analyticsAPIURL)
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

// GetCaskAnalytics retrieves cask analytics from API, optionally using cache.
func (d *DataProvider) GetCaskAnalytics(forceRefresh bool) (map[string]models.AnalyticsItem, error) {
	if err := ensureCacheDir(); err != nil {
		return nil, err
	}

	if !forceRefresh {
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

	body, err := fetchFromAPI(caskAnalyticsAPIURL)
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

// GetTapPackages retrieves package info for third-party tap entries.
// It checks cache first, then fetches missing packages via `brew info`.
// Results are cached for faster subsequent lookups.
func (d *DataProvider) GetTapPackages(entries []models.BrewfileEntry, existingPackages map[string]models.Package, forceRefresh bool) ([]models.Package, error) {
	if len(entries) == 0 {
		return nil, nil
	}

	result := make([]models.Package, 0)
	foundPackages := make(map[string]bool)

	// 1. Get from cache (if not forceRefresh)
	cachedPackages := make(map[string]models.Package)
	if !forceRefresh {
		if data := readCacheFile(cacheFileTapPackages, 10); data != nil {
			var packages []models.Package
			if err := json.Unmarshal(data, &packages); err == nil {
				for _, pkg := range packages {
					cachedPackages[pkg.Name] = pkg
				}
			}
		}
	}

	// 2. Collect packages from existingPackages (already loaded from APIs)
	// and packages from cache, tracking what we still need to fetch
	var missingCasks []string
	var missingFormulae []string

	for _, entry := range entries {
		// Check if already in existingPackages (from API)
		if pkg, exists := existingPackages[entry.Name]; exists {
			result = append(result, pkg)
			foundPackages[entry.Name] = true
			continue
		}

		// Check if in cache
		if pkg, exists := cachedPackages[entry.Name]; exists {
			result = append(result, pkg)
			foundPackages[entry.Name] = true
			continue
		}

		// Need to fetch this package
		if entry.IsCask {
			missingCasks = append(missingCasks, entry.Name)
		} else {
			missingFormulae = append(missingFormulae, entry.Name)
		}
	}

	// 3. Fetch missing packages via brew info
	if len(missingCasks) > 0 {
		fetched := d.fetchPackagesInfo(missingCasks, true)
		for _, name := range missingCasks {
			if pkg, exists := fetched[name]; exists {
				result = append(result, pkg)
			} else {
				// Fallback for packages that couldn't be fetched
				result = append(result, models.Package{
					Name:        name,
					DisplayName: name,
					Description: "(unable to load package info)",
					Type:        models.PackageTypeCask,
				})
			}
		}
	}

	if len(missingFormulae) > 0 {
		fetched := d.fetchPackagesInfo(missingFormulae, false)
		for _, name := range missingFormulae {
			if pkg, exists := fetched[name]; exists {
				result = append(result, pkg)
			} else {
				// Fallback for packages that couldn't be fetched
				result = append(result, models.Package{
					Name:        name,
					DisplayName: name,
					Description: "(unable to load package info)",
					Type:        models.PackageTypeFormula,
				})
			}
		}
	}

	// 4. Save all tap packages to cache
	if len(result) > 0 {
		if err := ensureCacheDir(); err == nil {
			if data, err := json.Marshal(result); err == nil {
				writeCacheFile(cacheFileTapPackages, data)
			}
		}
	}

	return result, nil
}

// fetchPackagesInfo retrieves package info via brew info command.
func (d *DataProvider) fetchPackagesInfo(names []string, isCask bool) map[string]models.Package {
	result := make(map[string]models.Package)
	if len(names) == 0 {
		return result
	}

	var cmd *exec.Cmd
	if isCask {
		args := append([]string{"info", "--json=v2", "--cask"}, names...)
		cmd = exec.Command("brew", args...)
	} else {
		args := append([]string{"info", "--json=v1"}, names...)
		cmd = exec.Command("brew", args...)
	}

	output, err := cmd.Output()
	if err != nil {
		// Try individual fetches as fallback
		for _, name := range names {
			if pkg := d.fetchSinglePackageInfo(name, isCask); pkg != nil {
				result[name] = *pkg
			}
		}
		return result
	}

	if isCask {
		var response struct {
			Casks []models.Cask `json:"casks"`
		}
		if err := json.Unmarshal(output, &response); err == nil {
			for _, cask := range response.Casks {
				c := cask
				pkg := models.NewPackageFromCask(&c)
				result[c.Token] = pkg
				// Also map FullToken if available (e.g. user/repo/token)
				if c.FullToken != "" && c.FullToken != c.Token {
					result[c.FullToken] = pkg
				}
			}
		}
	} else {
		var formulae []models.Formula
		if err := json.Unmarshal(output, &formulae); err == nil {
			for _, formula := range formulae {
				f := formula
				pkg := models.NewPackageFromFormula(&f)
				result[f.Name] = pkg
				// Also map FullName if available (e.g. user/repo/name)
				if f.FullName != "" && f.FullName != f.Name {
					result[f.FullName] = pkg
				}
			}
		}
	}

	return result
}

// fetchSinglePackageInfo fetches info for a single package.
func (d *DataProvider) fetchSinglePackageInfo(name string, isCask bool) *models.Package {
	var cmd *exec.Cmd
	if isCask {
		cmd = exec.Command("brew", "info", "--json=v2", "--cask", name)
	} else {
		cmd = exec.Command("brew", "info", "--json=v1", name)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	if isCask {
		var response struct {
			Casks []models.Cask `json:"casks"`
		}
		if err := json.Unmarshal(output, &response); err != nil || len(response.Casks) == 0 {
			return nil
		}
		pkg := models.NewPackageFromCask(&response.Casks[0])
		return &pkg
	}

	var formulae []models.Formula
	if err := json.Unmarshal(output, &formulae); err != nil || len(formulae) == 0 {
		return nil
	}
	pkg := models.NewPackageFromFormula(&formulae[0])
	return &pkg
}

// SetupData initializes the DataProvider by loading all package data.
func (d *DataProvider) SetupData(forceRefresh bool) error {
	// Get installed formulae
	installed, err := d.GetInstalledFormulae(forceRefresh)
	if err != nil {
		return fmt.Errorf("failed to get installed formulae: %w", err)
	}
	*d.installedFormulae = installed

	// Get remote formulae
	remote, err := d.GetRemoteFormulae(forceRefresh)
	if err != nil {
		return fmt.Errorf("failed to get remote formulae: %w", err)
	}
	*d.remoteFormulae = remote

	// Get formulae analytics
	analytics, err := d.GetFormulaeAnalytics(forceRefresh)
	if err != nil {
		return fmt.Errorf("failed to get formulae analytics: %w", err)
	}
	d.formulaeAnalytics = analytics

	// Get installed casks
	installedCasks, err := d.GetInstalledCasks(forceRefresh)
	if err != nil {
		return fmt.Errorf("failed to get installed casks: %w", err)
	}
	*d.installedCasks = installedCasks

	// Get remote casks
	remoteCasks, err := d.GetRemoteCasks(forceRefresh)
	if err != nil {
		return fmt.Errorf("failed to get remote casks: %w", err)
	}
	*d.remoteCasks = remoteCasks

	// Get cask analytics
	caskAnalytics, err := d.GetCaskAnalytics(forceRefresh)
	if err != nil {
		return fmt.Errorf("failed to get cask analytics: %w", err)
	}
	d.caskAnalytics = caskAnalytics

	return nil
}

// GetPackages retrieves all packages (formulae + casks), merging remote and installed.
func (d *DataProvider) GetPackages() *[]models.Package {
	packageMap := make(map[string]models.Package)

	for _, formula := range *d.remoteFormulae {
		if _, exists := packageMap[formula.Name]; !exists {
			f := formula
			pkg := models.NewPackageFromFormula(&f)
			if a, exists := d.formulaeAnalytics[formula.Name]; exists && a.Number > 0 {
				downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
				pkg.Analytics90dRank = a.Number
				pkg.Analytics90dDownloads = downloads
			}
			packageMap[formula.Name] = pkg
		}
	}

	for _, formula := range *d.installedFormulae {
		f := formula
		pkg := models.NewPackageFromFormula(&f)
		if a, exists := d.formulaeAnalytics[formula.Name]; exists && a.Number > 0 {
			downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
			pkg.Analytics90dRank = a.Number
			pkg.Analytics90dDownloads = downloads
		}
		packageMap[formula.Name] = pkg
	}

	for _, cask := range *d.remoteCasks {
		if _, exists := packageMap[cask.Token]; !exists {
			c := cask
			pkg := models.NewPackageFromCask(&c)
			if a, exists := d.caskAnalytics[cask.Token]; exists && a.Number > 0 {
				downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
				pkg.Analytics90dRank = a.Number
				pkg.Analytics90dDownloads = downloads
			}
			packageMap[cask.Token] = pkg
		}
	}

	for _, cask := range *d.installedCasks {
		c := cask
		pkg := models.NewPackageFromCask(&c)
		if a, exists := d.caskAnalytics[cask.Token]; exists && a.Number > 0 {
			downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
			pkg.Analytics90dRank = a.Number
			pkg.Analytics90dDownloads = downloads
		}
		packageMap[cask.Token] = pkg
	}

	*d.allPackages = make([]models.Package, 0, len(packageMap))
	for _, pkg := range packageMap {
		*d.allPackages = append(*d.allPackages, pkg)
	}

	sort.Slice(*d.allPackages, func(i, j int) bool {
		return (*d.allPackages)[i].Name < (*d.allPackages)[j].Name
	})

	return d.allPackages
}

// fetchInstalledNames returns a map of installed package names for the given type.
func (d *DataProvider) fetchInstalledNames(packageType string) map[string]bool {
	result := make(map[string]bool)
	cmd := exec.Command("brew", "list", packageType)
	output, err := cmd.Output()
	if err != nil {
		return result
	}
	for _, name := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if name != "" {
			result[name] = true
		}
	}
	return result
}

// FetchInstalledCaskNames returns a map of installed cask names for quick lookup.
// Note: This runs `brew list --cask` each time it's called.
func (d *DataProvider) FetchInstalledCaskNames() map[string]bool {
	return d.fetchInstalledNames("--cask")
}

// FetchInstalledFormulaNames returns a map of installed formula names for quick lookup.
// Note: This runs `brew list --formula` each time it's called.
func (d *DataProvider) FetchInstalledFormulaNames() map[string]bool {
	return d.fetchInstalledNames("--formula")
}
