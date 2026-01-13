package services

import (
	"bbrew/internal/models"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

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
}

// FlatpakService implements FlatpakServiceInterface.
type FlatpakService struct{}

// NewFlatpakService creates a new instance of FlatpakService.
var NewFlatpakService = func() FlatpakServiceInterface {
	return &FlatpakService{}
}

// IsFlatpakInstalled checks if the flatpak binary exists in the PATH.
func (s *FlatpakService) IsFlatpakInstalled() bool {
	_, err := exec.LookPath("flatpak")
	return err == nil
}

// EnsureFlathubRemote checks if the flathub remote exists, and adds it if missing.
func (s *FlatpakService) EnsureFlathubRemote(app *tview.Application, outputView *tview.TextView) error {
	// Check if flathub exists
	checkCmd := exec.Command("flatpak", "remote-list")
	output, err := checkCmd.Output()
	if err == nil && strings.Contains(string(output), "flathub") {
		return nil // Already exists
	}

	// Add flathub
	addCmd := exec.Command("flatpak", "remote-add", "--if-not-exists", "flathub", "https://dl.flathub.org/repo/flathub.flatpakrepo")
	return s.executeCommand(app, addCmd, outputView)
}

// GetInstalledPackages returns a map of installed Flatpak application IDs.
func (s *FlatpakService) GetInstalledPackages() (map[string]bool, error) {
	cmd := exec.Command("flatpak", "list", "--app", "--columns=application")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	installed := make(map[string]bool)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if id := strings.TrimSpace(line); id != "" {
			installed[id] = true
		}
	}
	return installed, nil
}

// GetRemoteMetadata fetches metadata (name, version, description) for all applications in Flathub.
// This is an expensive operation so it should be used sparingly or cached at the app level.
func (s *FlatpakService) GetRemoteMetadata() (map[string]models.Package, error) {
	// Fetch columns: application ID, name, version, description
	cmd := exec.Command("flatpak", "remote-ls", "flathub", "--app", "--columns=application,name,version,description")
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
			
			// Some rows might have missing fields that flatpak leaves as empty strings or skips?
			// Checking actual output suggests it tabs empty fields correctly.

			metadata[id] = models.Package{
				Name:        id,
				DisplayName: name,
				Version:     version,
				Description: desc,
			}
		}
	}
	return metadata, nil
}


// InstallPackage installs a Flatpak from Flathub.
func (s *FlatpakService) InstallPackage(info models.Package, app *tview.Application, outputView *tview.TextView) error {
	// flatpak install -y flathub <app-id>
	cmd := exec.Command("flatpak", "install", "-y", "flathub", info.Name)
	return s.executeCommand(app, cmd, outputView)
}

// RemovePackage uninstalls a Flatpak.
func (s *FlatpakService) RemovePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error {
	// flatpak uninstall -y <app-id>
	cmd := exec.Command("flatpak", "uninstall", "-y", info.Name)
	return s.executeCommand(app, cmd, outputView)
}

// UpdatePackage updates a specific Flatpak.
func (s *FlatpakService) UpdatePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error {
	// flatpak update -y <app-id>
	cmd := exec.Command("flatpak", "update", "-y", info.Name)
	return s.executeCommand(app, cmd, outputView)
}

// executeCommand runs a command and captures its output, updating the provided TextView.
// Duplicated from BrewService for modularity as requested (no shared base yet).
func (s *FlatpakService) executeCommand(
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

	var wg sync.WaitGroup
	wg.Add(3)

	cmdErrCh := make(chan error, 1)

	go func() {
		defer wg.Done()
		defer stdoutWriter.Close()
		defer stderrWriter.Close()
		cmdErrCh <- cmd.Wait()
	}()

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
					_, _ = outputView.Write(output)
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
					_, _ = outputView.Write(output)
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

	return <-cmdErrCh
}
