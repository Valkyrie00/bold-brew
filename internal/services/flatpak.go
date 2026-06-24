package services

import (
	"bbrew/internal/models"
	"os/exec"
	"strings"

	"github.com/rivo/tview"
)

// FlatpakServiceInterface defines the contract for Flatpak operations.
type FlatpakServiceInterface interface {
	IsFlatpakInstalled() bool
	EnsureFlathubRemote(app *tview.Application, outputView *tview.TextView) error
	GetInstalledPackages() (map[string]bool, error)
	GetRemoteMetadata() (map[string]models.Package, error)
	InstallPackage(info models.Package, app *tview.Application, outputView *tview.TextView) error
	RemovePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error
	UpdatePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error
	UpdateAllPackages(app *tview.Application, outputView *tview.TextView) error
}

// FlatpakService implements FlatpakServiceInterface.
type FlatpakService struct {
	cachedMetadata map[string]models.Package
}

// NewFlatpakService creates a new instance of FlatpakService.
var NewFlatpakService = func() FlatpakServiceInterface {
	return &FlatpakService{}
}

// IsFlatpakInstalled checks if the flatpak binary exists in the PATH.
func (s *FlatpakService) IsFlatpakInstalled() bool {
	_, err := exec.LookPath("flatpak")
	return err == nil
}

// EnsureFlathubRemote ensures flathub is available as a user-level remote.
// Even if flathub exists at system level, --user installs need a user-level remote.
func (s *FlatpakService) EnsureFlathubRemote(app *tview.Application, outputView *tview.TextView) error {
	checkCmd := exec.Command("flatpak", "remote-list", "--user")
	output, err := checkCmd.Output()
	if err == nil && strings.Contains(string(output), "flathub") {
		return nil
	}

	addCmd := exec.Command("flatpak", "remote-add", "--user", "--if-not-exists", "flathub", "https://dl.flathub.org/repo/flathub.flatpakrepo")
	return s.executeCommand(app, addCmd, outputView)
}

// GetInstalledPackages returns a map of installed Flatpak application IDs (both user and system).
func (s *FlatpakService) GetInstalledPackages() (map[string]bool, error) {
	installed := make(map[string]bool)

	for _, scope := range []string{"--user", "--system"} {
		cmd := exec.Command("flatpak", "list", scope, "--app", "--columns=application") // #nosec G204 - scope is a hardcoded constant
		output, err := cmd.Output()
		if err != nil {
			continue
		}
		for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
			if id := strings.TrimSpace(line); id != "" {
				installed[id] = true
			}
		}
	}

	return installed, nil
}

// GetRemoteMetadata fetches metadata (name, version, description) for all applications in Flathub.
// Results are cached in memory to avoid repeated expensive `flatpak remote-ls` calls.
func (s *FlatpakService) GetRemoteMetadata() (map[string]models.Package, error) {
	if s.cachedMetadata != nil {
		return s.cachedMetadata, nil
	}

	cmd := exec.Command("flatpak", "remote-ls", "--user", "flathub", "--app", "--columns=application,name,version,description")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	metadata := make(map[string]models.Package)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) >= 1 {
			id := strings.TrimSpace(parts[0])
			name := id
			version := ""
			desc := ""

			if len(parts) >= 2 {
				name = strings.TrimSpace(parts[1])
			}
			if len(parts) >= 3 {
				version = strings.TrimSpace(parts[2])
			}
			if len(parts) >= 4 {
				desc = strings.TrimSpace(parts[3])
			}

			metadata[id] = models.Package{
				Name:        id,
				DisplayName: name,
				Version:     version,
				Description: desc,
				Type:        models.PackageTypeFlatpak,
			}
		}
	}

	s.cachedMetadata = metadata
	return metadata, nil
}

// InstallPackage installs a Flatpak from Flathub.
func (s *FlatpakService) InstallPackage(info models.Package, app *tview.Application, outputView *tview.TextView) error {
	cmd := exec.Command("flatpak", "install", "--user", "-y", "flathub", info.Name) // #nosec G204
	return s.executeCommand(app, cmd, outputView)
}

// RemovePackage uninstalls a Flatpak.
func (s *FlatpakService) RemovePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error {
	cmd := exec.Command("flatpak", "uninstall", "--user", "-y", info.Name) // #nosec G204
	return s.executeCommand(app, cmd, outputView)
}

// UpdatePackage updates a specific Flatpak.
func (s *FlatpakService) UpdatePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error {
	cmd := exec.Command("flatpak", "update", "--user", "-y", info.Name) // #nosec G204
	return s.executeCommand(app, cmd, outputView)
}

// UpdateAllPackages updates all installed user-level Flatpak applications.
func (s *FlatpakService) UpdateAllPackages(app *tview.Application, outputView *tview.TextView) error {
	cmd := exec.Command("flatpak", "update", "--user", "-y") // #nosec G204
	return s.executeCommand(app, cmd, outputView)
}

// executeCommand runs a command and captures its output, updating the provided TextView.
func (s *FlatpakService) executeCommand(app *tview.Application, cmd *exec.Cmd, outputView *tview.TextView) error {
	return ExecuteCommand(app, cmd, outputView)
}
