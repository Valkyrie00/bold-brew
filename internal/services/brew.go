package services

import (
	"bbrew/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/adrg/xdg"
	"github.com/rivo/tview"
)

const FormulaeAPIURL = "https://formulae.brew.sh/api/formula.json"
const CaskAPIURL = "https://formulae.brew.sh/api/cask.json"
const AnalyticsAPIURL = "https://formulae.brew.sh/api/analytics/install-on-request/90d.json"
const CaskAnalyticsAPIURL = "https://formulae.brew.sh/api/analytics/cask-install/90d.json"

// getCacheDir - returns the cache directory following XDG Base Directory Specification.
func getCacheDir() string {
	return filepath.Join(xdg.CacheHome, "bbrew")
}

type BrewServiceInterface interface {
	GetPrefixPath() (path string)
	GetFormulae() (formulae *[]models.Formula)
	GetPackages() (packages *[]models.Package)
	SetupData(forceDownload bool) (err error)
	GetBrewVersion() (version string, err error)

	UpdateHomebrew() error
	UpdateAllPackages(app *tview.Application, outputView *tview.TextView) error
	UpdatePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error
	RemovePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error
	InstallPackage(info models.Package, app *tview.Application, outputView *tview.TextView) error
	InstallAllPackages(packages []models.Package, app *tview.Application, outputView *tview.TextView) error
	RemoveAllPackages(packages []models.Package, app *tview.Application, outputView *tview.TextView) error
	ParseBrewfile(filepath string) ([]models.BrewfileEntry, error)
	ParseBrewfileWithTaps(filepath string) (*models.BrewfileResult, error)
	InstallTap(tapName string, app *tview.Application, outputView *tview.TextView) error
	IsTapInstalled(tapName string) bool
	IsPackageInstalled(name string, isCask bool) bool
	GetInstalledCaskNames() map[string]bool
	GetInstalledFormulaNames() map[string]bool
	GetPackagesInfo(names []string, isCask bool) map[string]models.Package
	LoadTapPackagesCache() map[string]models.Package
	SaveTapPackagesToCache(packages []models.Package) error
}

// BrewService provides methods to interact with Homebrew, including
// retrieving formulae, casks, and handling analytics.
type BrewService struct {
	// Formula lists
	all       *[]models.Formula
	installed *[]models.Formula
	remote    *[]models.Formula
	analytics map[string]models.AnalyticsItem

	// Cask lists
	allCasks       *[]models.Cask
	installedCasks *[]models.Cask
	remoteCasks    *[]models.Cask
	caskAnalytics  map[string]models.AnalyticsItem

	// Unified package list
	allPackages *[]models.Package

	brewVersion string
	prefixPath  string
}

// NewBrewService creates a new instance of BrewService with initialized package lists.
var NewBrewService = func() BrewServiceInterface {
	return &BrewService{
		all:            new([]models.Formula),
		installed:      new([]models.Formula),
		remote:         new([]models.Formula),
		allCasks:       new([]models.Cask),
		installedCasks: new([]models.Cask),
		remoteCasks:    new([]models.Cask),
		allPackages:    new([]models.Package),
	}
}

// GetPrefixPath retrieves the Homebrew prefix path, caching it for future calls.
func (s *BrewService) GetPrefixPath() (path string) {
	if s.prefixPath != "" {
		return s.prefixPath
	}

	cmd := exec.Command("brew", "--prefix")
	output, err := cmd.Output()
	if err != nil {
		s.prefixPath = "Unknown"
		return
	}

	s.prefixPath = strings.TrimSpace(string(output))
	return s.prefixPath
}

// GetFormulae retrieves all formulae, merging remote and installed packages,
func (s *BrewService) GetFormulae() (formulae *[]models.Formula) {
	packageMap := make(map[string]models.Formula)

	// Add REMOTE packages to the map if they don't already exist
	for _, formula := range *s.remote {
		if _, exists := packageMap[formula.Name]; !exists {
			packageMap[formula.Name] = formula
		}
	}

	// Add INSTALLED packages to the map
	for _, formula := range *s.installed {
		packageMap[formula.Name] = formula
	}

	*s.all = make([]models.Formula, 0, len(packageMap))
	for _, formula := range packageMap {
		// Merge analytics data if available
		if a, exists := s.analytics[formula.Name]; exists && a.Number > 0 {
			downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
			formula.Analytics90dRank = a.Number
			formula.Analytics90dDownloads = downloads
		}

		*s.all = append(*s.all, formula)
	}

	// Sort the list by name
	sort.Slice(*s.all, func(i, j int) bool {
		return (*s.all)[i].Name < (*s.all)[j].Name
	})

	return s.all
}

// GetPackages retrieves all packages (formulae + casks), merging remote and installed.
func (s *BrewService) GetPackages() (packages *[]models.Package) {
	packageMap := make(map[string]models.Package)

	// Add REMOTE formulae
	for _, formula := range *s.remote {
		if _, exists := packageMap[formula.Name]; !exists {
			f := formula // Create a copy to avoid implicit memory aliasing
			pkg := models.NewPackageFromFormula(&f)
			// Merge analytics data
			if a, exists := s.analytics[formula.Name]; exists && a.Number > 0 {
				downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
				pkg.Analytics90dRank = a.Number
				pkg.Analytics90dDownloads = downloads
			}
			packageMap[formula.Name] = pkg
		}
	}

	// Add INSTALLED formulae (override remote data)
	for _, formula := range *s.installed {
		f := formula // Create a copy to avoid implicit memory aliasing
		pkg := models.NewPackageFromFormula(&f)
		// Merge analytics data
		if a, exists := s.analytics[formula.Name]; exists && a.Number > 0 {
			downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
			pkg.Analytics90dRank = a.Number
			pkg.Analytics90dDownloads = downloads
		}
		packageMap[formula.Name] = pkg
	}

	// Add REMOTE casks
	for _, cask := range *s.remoteCasks {
		if _, exists := packageMap[cask.Token]; !exists {
			c := cask // Create a copy to avoid implicit memory aliasing
			pkg := models.NewPackageFromCask(&c)
			// Merge analytics data
			if a, exists := s.caskAnalytics[cask.Token]; exists && a.Number > 0 {
				downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
				pkg.Analytics90dRank = a.Number
				pkg.Analytics90dDownloads = downloads
			}
			packageMap[cask.Token] = pkg
		}
	}

	// Add INSTALLED casks (override remote data)
	for _, cask := range *s.installedCasks {
		c := cask // Create a copy to avoid implicit memory aliasing
		pkg := models.NewPackageFromCask(&c)
		// Merge analytics data
		if a, exists := s.caskAnalytics[cask.Token]; exists && a.Number > 0 {
			downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
			pkg.Analytics90dRank = a.Number
			pkg.Analytics90dDownloads = downloads
		}
		packageMap[cask.Token] = pkg
	}

	// Convert map to slice
	*s.allPackages = make([]models.Package, 0, len(packageMap))
	for _, pkg := range packageMap {
		*s.allPackages = append(*s.allPackages, pkg)
	}

	// Sort by name
	sort.Slice(*s.allPackages, func(i, j int) bool {
		return (*s.allPackages)[i].Name < (*s.allPackages)[j].Name
	})

	return s.allPackages
}

// SetupData initializes the BrewService by loading installed packages, remote formulae, casks, and analytics data.
func (s *BrewService) SetupData(forceDownload bool) (err error) {
	// Load formulae
	if err = s.loadInstalled(forceDownload); err != nil {
		return fmt.Errorf("failed to load installed formulae: %w", err)
	}

	if err = s.loadRemote(forceDownload); err != nil {
		return fmt.Errorf("failed to load remote formulae: %w", err)
	}

	if err = s.loadAnalytics(forceDownload); err != nil {
		return fmt.Errorf("failed to load formulae analytics: %w", err)
	}

	// Load casks
	if err = s.loadInstalledCasks(forceDownload); err != nil {
		return fmt.Errorf("failed to load installed casks: %w", err)
	}

	if err = s.loadRemoteCasks(forceDownload); err != nil {
		return fmt.Errorf("failed to load remote casks: %w", err)
	}

	if err = s.loadCaskAnalytics(forceDownload); err != nil {
		return fmt.Errorf("failed to load cask analytics: %w", err)
	}

	return nil
}

// loadInstalled retrieves installed formulae, optionally using cache.
func (s *BrewService) loadInstalled(forceDownload bool) (err error) {
	cacheDir := getCacheDir()
	installedFile := filepath.Join(cacheDir, "installed.json")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0750); err != nil {
			return err
		}
	}

	// Check if we should use the cached file
	if !forceDownload {
		if fileInfo, err := os.Stat(installedFile); err == nil {
			// Only use cache if file is not empty (size > 10 bytes for "[]" or "{}")
			if fileInfo.Size() > 10 {
				// #nosec G304 -- installedFile path is safely constructed from UserHomeDir and sanitized with filepath.Join
				data, err := os.ReadFile(installedFile)
				if err == nil && len(data) > 0 {
					*s.installed = make([]models.Formula, 0)
					if err := json.Unmarshal(data, &s.installed); err == nil {
						// Mark all installed Packages as locally installed and set LocalPath
						prefix := s.GetPrefixPath()
						for i := range *s.installed {
							(*s.installed)[i].LocallyInstalled = true
							(*s.installed)[i].LocalPath = filepath.Join(prefix, "Cellar", (*s.installed)[i].Name)
						}
						return nil
					}
				}
			}
		}
	}

	cmd := exec.Command("brew", "info", "--json=v1", "--installed")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	*s.installed = make([]models.Formula, 0)
	err = json.Unmarshal(output, &s.installed)
	if err != nil {
		return err
	}

	// Mark all installed Packages as locally installed and set LocalPath
	prefix := s.GetPrefixPath()
	for i := range *s.installed {
		(*s.installed)[i].LocallyInstalled = true
		(*s.installed)[i].LocalPath = filepath.Join(prefix, "Cellar", (*s.installed)[i].Name)
	}

	// Cache the installed formulae data
	_ = os.WriteFile(installedFile, output, 0600)
	return nil
}

// loadInstalledCasks retrieves installed casks, optionally using cache.
func (s *BrewService) loadInstalledCasks(forceDownload bool) (err error) {
	cacheDir := getCacheDir()
	installedCasksFile := filepath.Join(cacheDir, "installed-casks.json")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0750); err != nil {
			return err
		}
	}

	// Check if we should use the cached file
	if !forceDownload {
		if fileInfo, err := os.Stat(installedCasksFile); err == nil {
			// Only use cache if file is not empty (size > 10 bytes for minimal JSON)
			if fileInfo.Size() > 10 {
				// #nosec G304 -- installedCasksFile path is safely constructed from UserHomeDir and sanitized with filepath.Join
				data, err := os.ReadFile(installedCasksFile)
				if err == nil && len(data) > 0 {
					var response struct {
						Casks []models.Cask `json:"casks"`
					}
					if err := json.Unmarshal(data, &response); err == nil {
						*s.installedCasks = response.Casks
						// Mark all installed casks as locally installed
						for i := range *s.installedCasks {
							(*s.installedCasks)[i].LocallyInstalled = true
							(*s.installedCasks)[i].IsCask = true
						}
						return nil
					}
				}
			}
		}
	}

	// Get list of installed cask names
	listCmd := exec.Command("brew", "list", "--cask")
	listOutput, err := listCmd.Output()
	if err != nil {
		// If no casks are installed, brew returns error - ignore it
		*s.installedCasks = make([]models.Cask, 0)
		return nil
	}

	// Parse cask names (one per line)
	caskNames := strings.Split(strings.TrimSpace(string(listOutput)), "\n")
	if len(caskNames) == 0 || (len(caskNames) == 1 && caskNames[0] == "") {
		*s.installedCasks = make([]models.Cask, 0)
		return nil
	}

	// Get info for each installed cask using --json=v2 (v2 required for casks)
	args := append([]string{"info", "--json=v2", "--cask"}, caskNames...)
	infoCmd := exec.Command("brew", args...)
	infoOutput, err := infoCmd.Output()
	if err != nil {
		*s.installedCasks = make([]models.Cask, 0)
		return nil
	}

	// Parse JSON response (v2 returns object with "formulae" and "casks" keys)
	// We only need the "casks" array since we specified --cask flag
	var response struct {
		Casks []models.Cask `json:"casks"`
	}
	err = json.Unmarshal(infoOutput, &response)
	if err != nil {
		return err
	}

	*s.installedCasks = response.Casks

	// Mark all installed casks as locally installed
	for i := range *s.installedCasks {
		(*s.installedCasks)[i].LocallyInstalled = true
		(*s.installedCasks)[i].IsCask = true
	}

	// Cache the installed casks data
	_ = os.WriteFile(installedCasksFile, infoOutput, 0600)
	return nil
}

// loadRemote retrieves the list of remote Homebrew formulae from the API and caches them locally.
func (s *BrewService) loadRemote(forceDownload bool) (err error) {
	cacheDir := getCacheDir()
	formulaFile := filepath.Join(cacheDir, "formula.json")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0750); err != nil {
			return err
		}
	}

	// Check if we should use the cached file
	if !forceDownload {
		if fileInfo, err := os.Stat(formulaFile); err == nil {
			// Only use cache if file has reasonable size (formulae list should be several MB)
			if fileInfo.Size() > 1000 {
				// #nosec G304 -- formulaFile path is safely constructed from UserHomeDir and sanitized with filepath.Join
				data, err := os.ReadFile(formulaFile)
				if err == nil && len(data) > 0 {
					*s.remote = make([]models.Formula, 0)
					if err := json.Unmarshal(data, &s.remote); err == nil && len(*s.remote) > 0 {
						return nil
					}
				}
			}
		}
	}

	resp, err := http.Get(FormulaeAPIURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	*s.remote = make([]models.Formula, 0)
	err = json.Unmarshal(body, s.remote)
	if err != nil {
		return err
	}

	// Cache the remote formulae data
	_ = os.WriteFile(formulaFile, body, 0600)
	return nil
}

// loadRemoteCasks retrieves the list of remote Homebrew casks from the API and caches them locally.
func (s *BrewService) loadRemoteCasks(forceDownload bool) (err error) {
	cacheDir := getCacheDir()
	caskFile := filepath.Join(cacheDir, "cask.json")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0750); err != nil {
			return err
		}
	}

	// Check if we should use the cached file
	if !forceDownload {
		if fileInfo, err := os.Stat(caskFile); err == nil {
			// Only use cache if file has reasonable size (cask list should be several MB)
			if fileInfo.Size() > 1000 {
				// #nosec G304 -- caskFile path is safely constructed from UserHomeDir and sanitized with filepath.Join
				data, err := os.ReadFile(caskFile)
				if err == nil && len(data) > 0 {
					*s.remoteCasks = make([]models.Cask, 0)
					if err := json.Unmarshal(data, &s.remoteCasks); err == nil && len(*s.remoteCasks) > 0 {
						return nil
					}
				}
			}
		}
	}

	resp, err := http.Get(CaskAPIURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	*s.remoteCasks = make([]models.Cask, 0)
	err = json.Unmarshal(body, s.remoteCasks)
	if err != nil {
		return err
	}

	// Cache the remote cask data
	_ = os.WriteFile(caskFile, body, 0600)
	return nil
}

// loadAnalytics retrieves the analytics data for Homebrew formulae from the API and caches them locally.
func (s *BrewService) loadAnalytics(forceDownload bool) (err error) {
	cacheDir := getCacheDir()
	analyticsFile := filepath.Join(cacheDir, "analytics.json")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0750); err != nil {
			return err
		}
	}

	// Check if we should use the cached file
	if !forceDownload {
		if fileInfo, err := os.Stat(analyticsFile); err == nil {
			// Only use cache if file has reasonable size (analytics should be > 1KB)
			if fileInfo.Size() > 100 {
				// #nosec G304 -- analyticsFile path is safely constructed from UserHomeDir and sanitized with filepath.Join
				data, err := os.ReadFile(analyticsFile)
				if err == nil && len(data) > 0 {
					analytics := models.Analytics{}
					if err := json.Unmarshal(data, &analytics); err == nil && len(analytics.Items) > 0 {
						analyticsByFormula := map[string]models.AnalyticsItem{}
						for _, f := range analytics.Items {
							analyticsByFormula[f.Formula] = f
						}
						s.analytics = analyticsByFormula
						return nil
					}
				}
			}
		}
	}

	resp, err := http.Get(AnalyticsAPIURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	analytics := models.Analytics{}
	err = json.Unmarshal(body, &analytics)
	if err != nil {
		return err
	}

	analyticsByFormula := map[string]models.AnalyticsItem{}
	for _, f := range analytics.Items {
		analyticsByFormula[f.Formula] = f
	}

	s.analytics = analyticsByFormula

	// Cache the analytics data
	_ = os.WriteFile(analyticsFile, body, 0600)
	return nil
}

// loadCaskAnalytics retrieves the analytics data for Homebrew casks from the API and caches them locally.
func (s *BrewService) loadCaskAnalytics(forceDownload bool) (err error) {
	cacheDir := getCacheDir()
	caskAnalyticsFile := filepath.Join(cacheDir, "cask-analytics.json")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0750); err != nil {
			return err
		}
	}

	// Check if we should use the cached file
	if !forceDownload {
		if fileInfo, err := os.Stat(caskAnalyticsFile); err == nil {
			// Only use cache if file has reasonable size (analytics should be > 1KB)
			if fileInfo.Size() > 100 {
				// #nosec G304 -- caskAnalyticsFile path is safely constructed from UserHomeDir and sanitized with filepath.Join
				data, err := os.ReadFile(caskAnalyticsFile)
				if err == nil && len(data) > 0 {
					analytics := models.Analytics{}
					if err := json.Unmarshal(data, &analytics); err == nil && len(analytics.Items) > 0 {
						analyticsByCask := map[string]models.AnalyticsItem{}
						for _, c := range analytics.Items {
							// Cask analytics use the "cask" field instead of "formula"
							caskName := c.Cask
							if caskName != "" {
								analyticsByCask[caskName] = c
							}
						}
						s.caskAnalytics = analyticsByCask
						return nil
					}
				}
			}
		}
	}

	resp, err := http.Get(CaskAnalyticsAPIURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	analytics := models.Analytics{}
	err = json.Unmarshal(body, &analytics)
	if err != nil {
		return err
	}

	analyticsByCask := map[string]models.AnalyticsItem{}
	for _, c := range analytics.Items {
		// Cask analytics use the "cask" field instead of "formula"
		caskName := c.Cask
		if caskName != "" {
			analyticsByCask[caskName] = c
		}
	}

	s.caskAnalytics = analyticsByCask

	// Cache the cask analytics data
	_ = os.WriteFile(caskAnalyticsFile, body, 0600)
	return nil
}

// GetBrewVersion retrieves the version of Homebrew installed on the system, caching it for future calls.
func (s *BrewService) GetBrewVersion() (version string, err error) {
	if s.brewVersion != "" {
		return s.brewVersion, nil
	}

	cmd := exec.Command("brew", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	s.brewVersion = strings.TrimSpace(string(output))
	return s.brewVersion, nil
}

// UpdateHomebrew updates the Homebrew package manager by running the `brew update` command.
func (s *BrewService) UpdateHomebrew() error {
	cmd := exec.Command("brew", "update")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (s *BrewService) UpdateAllPackages(app *tview.Application, outputView *tview.TextView) error {
	cmd := exec.Command("brew", "upgrade") // #nosec G204
	return s.executeCommand(app, cmd, outputView)
}

func (s *BrewService) UpdatePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error {
	var cmd *exec.Cmd
	if info.Type == models.PackageTypeCask {
		cmd = exec.Command("brew", "upgrade", "--cask", info.Name) // #nosec G204
	} else {
		cmd = exec.Command("brew", "upgrade", info.Name) // #nosec G204
	}
	return s.executeCommand(app, cmd, outputView)
}

func (s *BrewService) RemovePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error {
	var cmd *exec.Cmd
	if info.Type == models.PackageTypeCask {
		cmd = exec.Command("brew", "uninstall", "--cask", info.Name) // #nosec G204
	} else {
		cmd = exec.Command("brew", "uninstall", info.Name) // #nosec G204
	}
	return s.executeCommand(app, cmd, outputView)
}

func (s *BrewService) InstallPackage(info models.Package, app *tview.Application, outputView *tview.TextView) error {
	var cmd *exec.Cmd
	if info.Type == models.PackageTypeCask {
		cmd = exec.Command("brew", "install", "--cask", info.Name) // #nosec G204
	} else {
		cmd = exec.Command("brew", "install", info.Name) // #nosec G204
	}
	return s.executeCommand(app, cmd, outputView)
}

// executeCommand runs a command and captures its output, updating the provided TextView in the application.
func (s *BrewService) executeCommand(
	app *tview.Application,
	cmd *exec.Cmd,
	outputView *tview.TextView,
) error {
	stdoutPipe, stdoutWriter := io.Pipe()
	stderrPipe, stderrWriter := io.Pipe()
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	if err := cmd.Start(); err != nil {
		return err
	}

	// Add a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(3)

	// Channel to capture the command exit error
	cmdErrCh := make(chan error, 1)

	// Goroutine to wait for the command to finish
	go func() {
		defer wg.Done()
		defer stdoutWriter.Close()
		defer stderrWriter.Close()
		cmdErrCh <- cmd.Wait()
	}()

	// Stdout handler
	go func() {
		defer wg.Done()
		defer stdoutPipe.Close()
		buf := make([]byte, 1024)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				output := make([]byte, n)
				copy(output, buf[:n])
				app.QueueUpdateDraw(func() {
					_, _ = outputView.Write(output) // #nosec G104
					outputView.ScrollToEnd()
				})
			}
			if err != nil {
				if err != io.EOF {
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(outputView, "\nError: %v\n", err)
					})
				}
				break
			}
		}
	}()

	// Stderr handler
	go func() {
		defer wg.Done()
		defer stderrPipe.Close()
		buf := make([]byte, 1024)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				output := make([]byte, n)
				copy(output, buf[:n])
				app.QueueUpdateDraw(func() {
					_, _ = outputView.Write(output) // #nosec G104
					outputView.ScrollToEnd()
				})
			}
			if err != nil {
				if err != io.EOF {
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(outputView, "\nError: %v\n", err)
					})
				}
				break
			}
		}
	}()

	wg.Wait()

	// Return the command exit error (nil if successful)
	return <-cmdErrCh
}

// ParseBrewfile parses a Brewfile and returns a list of packages to be installed.
// It handles 'tap', 'brew' and 'cask' entries in the Brewfile format.
func (s *BrewService) ParseBrewfile(filepath string) ([]models.BrewfileEntry, error) {
	result, err := s.ParseBrewfileWithTaps(filepath)
	if err != nil {
		return nil, err
	}
	return result.Packages, nil
}

// ParseBrewfileWithTaps parses a Brewfile and returns taps and packages separately.
// This allows installing taps before packages.
func (s *BrewService) ParseBrewfileWithTaps(filepath string) (*models.BrewfileResult, error) {
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
			// Extract package name from quotes
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
			// Extract package name from quotes
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

// InstallTap installs a Homebrew tap.
func (s *BrewService) InstallTap(tapName string, app *tview.Application, outputView *tview.TextView) error {
	cmd := exec.Command("brew", "tap", tapName) // #nosec G204
	return s.executeCommand(app, cmd, outputView)
}

// IsTapInstalled checks if a tap is already installed.
func (s *BrewService) IsTapInstalled(tapName string) bool {
	cmd := exec.Command("brew", "tap")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	taps := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, tap := range taps {
		if strings.TrimSpace(tap) == tapName {
			return true
		}
	}
	return false
}

// IsPackageInstalled checks if a package (formula or cask) is installed by name.
func (s *BrewService) IsPackageInstalled(name string, isCask bool) bool {
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
func (s *BrewService) GetInstalledCaskNames() map[string]bool {
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
func (s *BrewService) GetInstalledFormulaNames() map[string]bool {
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

// getPackageInfoSingle retrieves info for a single package directly.
// Used internally as fallback when batch call fails.
func (s *BrewService) getPackageInfoSingle(name string, isCask bool) *models.Package {
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
		cask := response.Casks[0]
		cask.LocallyInstalled = s.IsPackageInstalled(name, true)
		pkg := models.NewPackageFromCask(&cask)
		return &pkg
	}

	var formulae []models.Formula
	if err := json.Unmarshal(output, &formulae); err != nil || len(formulae) == 0 {
		return nil
	}
	formula := formulae[0]
	formula.LocallyInstalled = s.IsPackageInstalled(name, false)
	pkg := models.NewPackageFromFormula(&formula)
	return &pkg
}

// GetPackagesInfo retrieves package information for multiple packages in a single brew call.
// This is much faster than calling GetPackageInfo for each package individually.
// If the batch call fails, it falls back to individual calls for each package.
// Returns a map of package name to Package.
func (s *BrewService) GetPackagesInfo(names []string, isCask bool) map[string]models.Package {
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
		// Batch call failed - try each package individually
		for _, name := range names {
			if pkg := s.getPackageInfoSingle(name, isCask); pkg != nil {
				result[name] = *pkg
			}
		}
		return result
	}

	if isCask {
		// Parse cask JSON (v2 format)
		var response struct {
			Casks []models.Cask `json:"casks"`
		}
		if err := json.Unmarshal(output, &response); err != nil {
			return result
		}
		for _, cask := range response.Casks {
			c := cask
			c.LocallyInstalled = s.IsPackageInstalled(c.Token, true)
			pkg := models.NewPackageFromCask(&c)
			result[c.Token] = pkg
		}
	} else {
		// Parse formula JSON (v1 format)
		var formulae []models.Formula
		if err := json.Unmarshal(output, &formulae); err != nil {
			return result
		}
		for _, formula := range formulae {
			f := formula
			f.LocallyInstalled = s.IsPackageInstalled(f.Name, false)
			pkg := models.NewPackageFromFormula(&f)
			result[f.Name] = pkg
		}
	}

	return result
}

// InstallAllPackages installs a list of packages sequentially.
// This is designed for Brewfile mode where the package list is curated and small.
func (s *BrewService) InstallAllPackages(packages []models.Package, app *tview.Application, outputView *tview.TextView) error {
	for _, pkg := range packages {
		// Check if package is already installed
		if pkg.LocallyInstalled {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(outputView, "[SKIP] %s (already installed)\n", pkg.Name)
			})
			continue
		}

		app.QueueUpdateDraw(func() {
			fmt.Fprintf(outputView, "\n[INSTALL] Installing %s...\n", pkg.Name)
		})

		if err := s.InstallPackage(pkg, app, outputView); err != nil {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(outputView, "[ERROR] Failed to install %s: %v\n", pkg.Name, err)
			})
			// Continue with next package even if one fails
			continue
		}

		app.QueueUpdateDraw(func() {
			fmt.Fprintf(outputView, "[SUCCESS] %s installed successfully\n", pkg.Name)
		})
	}

	return nil
}

// RemoveAllPackages removes a list of packages sequentially.
// This is designed for Brewfile mode where the package list is curated and small.
func (s *BrewService) RemoveAllPackages(packages []models.Package, app *tview.Application, outputView *tview.TextView) error {
	for _, pkg := range packages {
		// Check if package is not installed
		if !pkg.LocallyInstalled {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(outputView, "[SKIP] %s (not installed)\n", pkg.Name)
			})
			continue
		}

		app.QueueUpdateDraw(func() {
			fmt.Fprintf(outputView, "\n[REMOVE] Removing %s...\n", pkg.Name)
		})

		if err := s.RemovePackage(pkg, app, outputView); err != nil {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(outputView, "[ERROR] Failed to remove %s: %v\n", pkg.Name, err)
			})
			// Continue with next package even if one fails
			continue
		}

		app.QueueUpdateDraw(func() {
			fmt.Fprintf(outputView, "[SUCCESS] %s removed successfully\n", pkg.Name)
		})
	}

	return nil
}

// LoadTapPackagesCache loads cached tap packages from disk.
// Returns a map of package name to Package, or empty map if cache doesn't exist.
func (s *BrewService) LoadTapPackagesCache() map[string]models.Package {
	result := make(map[string]models.Package)

	cacheDir := getCacheDir()
	tapPackagesFile := filepath.Join(cacheDir, "tap_packages.json")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return result // Cache directory doesn't exist
	}

	// Check if cache file exists and has reasonable size
	if fileInfo, err := os.Stat(tapPackagesFile); err == nil {
		// Only use cache if file has reasonable size (> 10 bytes for minimal JSON "[]")
		if fileInfo.Size() > 10 {
			// #nosec G304 -- tapPackagesFile path is safely constructed from getCacheDir and sanitized with filepath.Join
			data, err := os.ReadFile(tapPackagesFile)
			if err == nil && len(data) > 0 {
				var packages []models.Package
				if err := json.Unmarshal(data, &packages); err == nil && len(packages) > 0 {
					// Convert to map for quick lookup
					for _, pkg := range packages {
						result[pkg.Name] = pkg
					}
					return result
				}
			}
		}
	}

	return result
}

// SaveTapPackagesToCache saves tap packages to disk cache.
func (s *BrewService) SaveTapPackagesToCache(packages []models.Package) error {
	cacheDir := getCacheDir()
	tapPackagesFile := filepath.Join(cacheDir, "tap_packages.json")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0750); err != nil {
			return err
		}
	}

	data, err := json.Marshal(packages)
	if err != nil {
		return err
	}

	// Cache the tap packages data
	return os.WriteFile(tapPackagesFile, data, 0600)
}
