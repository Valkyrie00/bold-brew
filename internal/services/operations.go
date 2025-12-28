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

// UpdateAllPackages upgrades all outdated packages.
func (s *BrewService) UpdateAllPackages(app *tview.Application, outputView *tview.TextView) error {
	cmd := exec.Command("brew", "upgrade") // #nosec G204
	return s.executeCommand(app, cmd, outputView)
}

// UpdatePackage upgrades a specific package.
func (s *BrewService) UpdatePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error {
	var cmd *exec.Cmd
	if info.Type == models.PackageTypeCask {
		cmd = exec.Command("brew", "upgrade", "--cask", info.Name) // #nosec G204
	} else {
		cmd = exec.Command("brew", "upgrade", info.Name) // #nosec G204
	}
	return s.executeCommand(app, cmd, outputView)
}

// RemovePackage uninstalls a package.
func (s *BrewService) RemovePackage(info models.Package, app *tview.Application, outputView *tview.TextView) error {
	var cmd *exec.Cmd
	if info.Type == models.PackageTypeCask {
		cmd = exec.Command("brew", "uninstall", "--cask", info.Name) // #nosec G204
	} else {
		cmd = exec.Command("brew", "uninstall", info.Name) // #nosec G204
	}
	return s.executeCommand(app, cmd, outputView)
}

// InstallPackage installs a package.
func (s *BrewService) InstallPackage(info models.Package, app *tview.Application, outputView *tview.TextView) error {
	var cmd *exec.Cmd
	if info.Type == models.PackageTypeCask {
		cmd = exec.Command("brew", "install", "--cask", info.Name) // #nosec G204
	} else {
		cmd = exec.Command("brew", "install", info.Name) // #nosec G204
	}
	return s.executeCommand(app, cmd, outputView)
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

// InstallAllPackages installs a list of packages sequentially.
func (s *BrewService) InstallAllPackages(packages []models.Package, app *tview.Application, outputView *tview.TextView) error {
	for _, pkg := range packages {
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
			continue
		}

		app.QueueUpdateDraw(func() {
			fmt.Fprintf(outputView, "[SUCCESS] %s installed successfully\n", pkg.Name)
		})
	}

	return nil
}

// RemoveAllPackages removes a list of packages sequentially.
func (s *BrewService) RemoveAllPackages(packages []models.Package, app *tview.Application, outputView *tview.TextView) error {
	for _, pkg := range packages {
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
			continue
		}

		app.QueueUpdateDraw(func() {
			fmt.Fprintf(outputView, "[SUCCESS] %s removed successfully\n", pkg.Name)
		})
	}

	return nil
}

// executeCommand runs a command and captures its output, updating the provided TextView.
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

	return <-cmdErrCh
}

