package services

import (
	"bbrew/internal/models"
	"os/exec"
	"strings"

	"github.com/rivo/tview"
)

// BrewServiceInterface defines the contract for Homebrew operations.
// BrewService is a pure executor of brew commands - it does NOT hold data.
// For data retrieval, use DataProviderInterface.
type BrewServiceInterface interface {
	// Core info
	GetBrewVersion() (string, error)

	// Package operations
	UpdateHomebrew() error
	UpdateAllPackages(app *tview.Application, outputView *tview.TextView) error
	UpdatePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error
	RemovePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error
	InstallPackage(info models.Package, app *tview.Application, outputView *tview.TextView) error

	// Tap support
	InstallTap(tapName string, app *tview.Application, outputView *tview.TextView) error
	IsTapInstalled(tapName string) bool
}

// BrewService provides methods to execute Homebrew commands.
// It is a pure executor - no data storage. Use DataProvider for data.
type BrewService struct {
	brewVersion string
}

// NewBrewService creates a new instance of BrewService.
var NewBrewService = func() BrewServiceInterface {
	return &BrewService{}
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
