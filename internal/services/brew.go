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

	"github.com/rivo/tview"
)

const FormulaeAPIURL = "https://formulae.brew.sh/api/formula.json"
const CaskAPIURL = "https://formulae.brew.sh/api/cask.json"
const AnalyticsAPIURL = "https://formulae.brew.sh/api/analytics/install-on-request/90d.json"
const CaskAnalyticsAPIURL = "https://formulae.brew.sh/api/analytics-cask/install/90d.json"

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
			pkg := models.NewPackageFromFormula(&formula)
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
		pkg := models.NewPackageFromFormula(&formula)
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
			pkg := models.NewPackageFromCask(&cask)
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
		pkg := models.NewPackageFromCask(&cask)
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
	if err = s.loadInstalled(); err != nil {
		return err
	}

	if err = s.loadRemote(forceDownload); err != nil {
		return err
	}

	if err = s.loadAnalytics(); err != nil {
		return err
	}

	// Load casks
	if err = s.loadInstalledCasks(); err != nil {
		return err
	}

	if err = s.loadRemoteCasks(forceDownload); err != nil {
		return err
	}

	if err = s.loadCaskAnalytics(); err != nil {
		return err
	}

	return nil
}

// loadInstalled retrieves the list of installed Homebrew formulae and updates their local paths.
func (s *BrewService) loadInstalled() (err error) {
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

	return nil
}

// loadInstalledCasks retrieves the list of installed Homebrew casks.
func (s *BrewService) loadInstalledCasks() (err error) {
	cmd := exec.Command("brew", "info", "--json=v1", "--cask", "--installed")
	output, err := cmd.Output()
	if err != nil {
		// If no casks are installed, brew returns error - ignore it
		*s.installedCasks = make([]models.Cask, 0)
		return nil
	}

	*s.installedCasks = make([]models.Cask, 0)
	err = json.Unmarshal(output, &s.installedCasks)
	if err != nil {
		return err
	}

	// Mark all installed casks as locally installed
	for i := range *s.installedCasks {
		(*s.installedCasks)[i].LocallyInstalled = true
		(*s.installedCasks)[i].IsCask = true
	}

	return nil
}

// loadRemote retrieves the list of remote Homebrew formulae from the API and caches them locally.
func (s *BrewService) loadRemote(forceDownload bool) (err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	bbrewDir := filepath.Join(homeDir, ".bbrew") // TODO: Move to config
	formulaFile := filepath.Join(bbrewDir, "formula.json")
	if _, err := os.Stat(bbrewDir); os.IsNotExist(err) {
		if err := os.MkdirAll(bbrewDir, 0750); err != nil {
			return err
		}
	}

	// Check if we should use the cached file
	if !forceDownload {
		if _, err := os.Stat(formulaFile); err == nil {
			// #nosec G304 -- formulaFile path is safely constructed from UserHomeDir and sanitized with filepath.Join
			data, err := os.ReadFile(formulaFile)
			if err == nil {
				*s.remote = make([]models.Formula, 0)
				if err := json.Unmarshal(data, &s.remote); err == nil {
					return nil
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	bbrewDir := filepath.Join(homeDir, ".bbrew")
	caskFile := filepath.Join(bbrewDir, "cask.json")
	if _, err := os.Stat(bbrewDir); os.IsNotExist(err) {
		if err := os.MkdirAll(bbrewDir, 0750); err != nil {
			return err
		}
	}

	// Check if we should use the cached file
	if !forceDownload {
		if _, err := os.Stat(caskFile); err == nil {
			// #nosec G304 -- caskFile path is safely constructed from UserHomeDir and sanitized with filepath.Join
			data, err := os.ReadFile(caskFile)
			if err == nil {
				*s.remoteCasks = make([]models.Cask, 0)
				if err := json.Unmarshal(data, &s.remoteCasks); err == nil {
					return nil
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

// loadAnalytics retrieves the analytics data for Homebrew formulae from the API.
func (s *BrewService) loadAnalytics() (err error) {
	resp, err := http.Get(AnalyticsAPIURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	analytics := models.Analytics{}
	err = json.NewDecoder(resp.Body).Decode(&analytics)
	if err != nil {
		return err
	}

	analyticsByFormula := map[string]models.AnalyticsItem{}
	for _, f := range analytics.Items {
		analyticsByFormula[f.Formula] = f
	}

	s.analytics = analyticsByFormula
	return nil
}

// loadCaskAnalytics retrieves the analytics data for Homebrew casks from the API.
func (s *BrewService) loadCaskAnalytics() (err error) {
	resp, err := http.Get(CaskAnalyticsAPIURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	analytics := models.Analytics{}
	err = json.NewDecoder(resp.Body).Decode(&analytics)
	if err != nil {
		return err
	}

	analyticsByCask := map[string]models.AnalyticsItem{}
	for _, c := range analytics.Items {
		analyticsByCask[c.Formula] = c // Formula field is used for both formulae and casks
	}

	s.caskAnalytics = analyticsByCask
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

	// Goroutine to wait for the command to finish
	go func() {
		defer wg.Done()
		defer stdoutWriter.Close()
		defer stderrWriter.Close()
		_ = cmd.Wait() // #nosec G104 -- Error is handled by pipe readers below
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
	return nil
}
