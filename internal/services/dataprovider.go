package services

import (
	"bbrew/internal/models"
	"encoding/json"
	"fmt"
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
	SetupData(forceDownload bool) error
	GetPackages() *[]models.Package
	GetFormulae() *[]models.Formula

	// Installation status checks
	IsPackageInstalled(name string, isCask bool) bool
	GetInstalledCaskNames() map[string]bool
	GetInstalledFormulaNames() map[string]bool

	// Tap packages - unified method that loads from cache or fetches via brew info
	LoadTapPackages(entries []models.BrewfileEntry, existingPackages map[string]models.Package, forceDownload bool) ([]models.Package, error)
}

// DataProvider implements DataProviderInterface.
// It is the central repository for all Homebrew package data.
type DataProvider struct {
	// Formula lists
	allFormulae       *[]models.Formula
	installedFormulae *[]models.Formula
	remoteFormulae    *[]models.Formula
	formulaeAnalytics map[string]models.AnalyticsItem

	// Cask lists
	allCasks       *[]models.Cask
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
		allFormulae:       new([]models.Formula),
		installedFormulae: new([]models.Formula),
		remoteFormulae:    new([]models.Formula),
		allCasks:          new([]models.Cask),
		installedCasks:    new([]models.Cask),
		remoteCasks:       new([]models.Cask),
		allPackages:       new([]models.Package),
	}
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

// LoadTapPackages loads tap packages using a unified pattern:
// 1. Load from cache (if not forceDownload)
// 2. For packages not in cache or existingPackages, fetch via brew info
// 3. Save all to cache
// 4. Return all tap packages
func (d *DataProvider) LoadTapPackages(entries []models.BrewfileEntry, existingPackages map[string]models.Package, forceDownload bool) ([]models.Package, error) {
	if len(entries) == 0 {
		return nil, nil
	}

	result := make([]models.Package, 0)
	foundPackages := make(map[string]bool)

	// 1. Load from cache (if not forceDownload)
	cachedPackages := make(map[string]models.Package)
	if !forceDownload {
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
			}
		}
	} else {
		var formulae []models.Formula
		if err := json.Unmarshal(output, &formulae); err == nil {
			for _, formula := range formulae {
				f := formula
				pkg := models.NewPackageFromFormula(&f)
				result[f.Name] = pkg
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

// =============================================================================
// Data Setup and Retrieval Methods
// =============================================================================

// SetupData initializes the DataProvider by loading all package data.
func (d *DataProvider) SetupData(forceDownload bool) error {
	// Load installed formulae
	installed, err := d.LoadInstalledFormulae(forceDownload)
	if err != nil {
		return fmt.Errorf("failed to load installed formulae: %w", err)
	}
	*d.installedFormulae = installed

	// Load remote formulae
	remote, err := d.LoadRemoteFormulae(forceDownload)
	if err != nil {
		return fmt.Errorf("failed to load remote formulae: %w", err)
	}
	*d.remoteFormulae = remote

	// Load formulae analytics
	analytics, err := d.LoadFormulaeAnalytics(forceDownload)
	if err != nil {
		return fmt.Errorf("failed to load formulae analytics: %w", err)
	}
	d.formulaeAnalytics = analytics

	// Load installed casks
	installedCasks, err := d.LoadInstalledCasks(forceDownload)
	if err != nil {
		return fmt.Errorf("failed to load installed casks: %w", err)
	}
	*d.installedCasks = installedCasks

	// Load remote casks
	remoteCasks, err := d.LoadRemoteCasks(forceDownload)
	if err != nil {
		return fmt.Errorf("failed to load remote casks: %w", err)
	}
	*d.remoteCasks = remoteCasks

	// Load cask analytics
	caskAnalytics, err := d.LoadCaskAnalytics(forceDownload)
	if err != nil {
		return fmt.Errorf("failed to load cask analytics: %w", err)
	}
	d.caskAnalytics = caskAnalytics

	return nil
}

// GetFormulae retrieves all formulae, merging remote and installed packages.
func (d *DataProvider) GetFormulae() *[]models.Formula {
	packageMap := make(map[string]models.Formula)

	for _, formula := range *d.remoteFormulae {
		if _, exists := packageMap[formula.Name]; !exists {
			packageMap[formula.Name] = formula
		}
	}

	for _, formula := range *d.installedFormulae {
		packageMap[formula.Name] = formula
	}

	*d.allFormulae = make([]models.Formula, 0, len(packageMap))
	for _, formula := range packageMap {
		if a, exists := d.formulaeAnalytics[formula.Name]; exists && a.Number > 0 {
			downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
			formula.Analytics90dRank = a.Number
			formula.Analytics90dDownloads = downloads
		}
		*d.allFormulae = append(*d.allFormulae, formula)
	}

	sort.Slice(*d.allFormulae, func(i, j int) bool {
		return (*d.allFormulae)[i].Name < (*d.allFormulae)[j].Name
	})

	return d.allFormulae
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

// IsPackageInstalled checks if a package (formula or cask) is installed by name.
func (d *DataProvider) IsPackageInstalled(name string, isCask bool) bool {
	var cmd *exec.Cmd
	if isCask {
		cmd = exec.Command("brew", "list", "--cask", name)
	} else {
		cmd = exec.Command("brew", "list", "--formula", name)
	}
	err := cmd.Run()
	return err == nil
}

// GetInstalledCaskNames returns a map of installed cask names for quick lookup.
func (d *DataProvider) GetInstalledCaskNames() map[string]bool {
	result := make(map[string]bool)
	cmd := exec.Command("brew", "list", "--cask")
	output, err := cmd.Output()
	if err != nil {
		return result
	}
	names := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, name := range names {
		if name != "" {
			result[name] = true
		}
	}
	return result
}

// GetInstalledFormulaNames returns a map of installed formula names for quick lookup.
func (d *DataProvider) GetInstalledFormulaNames() map[string]bool {
	result := make(map[string]bool)
	cmd := exec.Command("brew", "list", "--formula")
	output, err := cmd.Output()
	if err != nil {
		return result
	}
	names := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, name := range names {
		if name != "" {
			result[name] = true
		}
	}
	return result
}
