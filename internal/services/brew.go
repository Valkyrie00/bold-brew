package services

import (
	"bbrew/internal/models"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/rivo/tview"
)

// API URLs for Homebrew data
const (
	FormulaeAPIURL      = "https://formulae.brew.sh/api/formula.json"
	CaskAPIURL          = "https://formulae.brew.sh/api/cask.json"
	AnalyticsAPIURL     = "https://formulae.brew.sh/api/analytics/install-on-request/90d.json"
	CaskAnalyticsAPIURL = "https://formulae.brew.sh/api/analytics/cask-install/90d.json"
)

// getCacheDir returns the cache directory following XDG Base Directory Specification.
func getCacheDir() string {
	return filepath.Join(xdg.CacheHome, "bbrew")
}

// BrewServiceInterface defines the contract for interacting with Homebrew.
type BrewServiceInterface interface {
	// Core info
	GetPrefixPath() string
	GetBrewVersion() (string, error)

	// Package retrieval
	GetFormulae() *[]models.Formula
	GetPackages() *[]models.Package
	SetupData(forceDownload bool) error

	// Package operations
	UpdateHomebrew() error
	UpdateAllPackages(app *tview.Application, outputView *tview.TextView) error
	UpdatePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error
	RemovePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error
	InstallPackage(info models.Package, app *tview.Application, outputView *tview.TextView) error
	InstallAllPackages(packages []models.Package, app *tview.Application, outputView *tview.TextView) error
	RemoveAllPackages(packages []models.Package, app *tview.Application, outputView *tview.TextView) error

	// Brewfile support
	ParseBrewfile(filepath string) ([]models.BrewfileEntry, error)
	ParseBrewfileWithTaps(filepath string) (*models.BrewfileResult, error)

	// Tap support
	InstallTap(tapName string, app *tview.Application, outputView *tview.TextView) error
	IsTapInstalled(tapName string) bool

	// Package info
	IsPackageInstalled(name string, isCask bool) bool
	GetInstalledCaskNames() map[string]bool
	GetInstalledFormulaNames() map[string]bool
	GetPackagesInfo(names []string, isCask bool) map[string]models.Package

	// Cache
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
func (s *BrewService) GetPrefixPath() string {
	if s.prefixPath != "" {
		return s.prefixPath
	}

	cmd := exec.Command("brew", "--prefix")
	output, err := cmd.Output()
	if err != nil {
		s.prefixPath = "Unknown"
		return s.prefixPath
	}

	s.prefixPath = strings.TrimSpace(string(output))
	return s.prefixPath
}

// GetBrewVersion retrieves the version of Homebrew installed on the system, caching it for future calls.
func (s *BrewService) GetBrewVersion() (string, error) {
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
	return cmd.Run()
}

