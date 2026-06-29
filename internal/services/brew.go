package services

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"

	"bbrew/internal/models"
)

// brewEnv returns the environment variables for non-interactive Homebrew execution.
// This ensures commands don't hang waiting for user input (Homebrew 6+ ask mode)
// and is backward-compatible with older Homebrew versions.
func brewEnv() []string {
	return append(os.Environ(),
		"NONINTERACTIVE=1",
		"HOMEBREW_NO_AUTO_UPDATE=1",
		"HOMEBREW_NO_ENV_HINTS=1",
	)
}

// brewCommand creates an exec.Cmd for brew with non-interactive environment settings.
func brewCommand(args ...string) *exec.Cmd {
	cmd := exec.Command("brew", args...)
	cmd.Stdin = nil
	cmd.Env = brewEnv()
	return cmd
}

// brewCommandContext creates a context-aware exec.Cmd for brew with non-interactive settings.
func brewCommandContext(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "brew", args...)
	cmd.Stdin = nil
	cmd.Env = brewEnv()
	return cmd
}

// BrewServiceInterface defines the contract for Homebrew operations.
// BrewService is a pure executor of brew commands - it does NOT hold data.
// For data retrieval, use DataProviderInterface.
type BrewServiceInterface interface {
	GetBrewVersion() (string, error)
	UpdateHomebrew() error
	UpdateAllPackages(output io.Writer) error
	UpdatePackage(info models.Package, output io.Writer) error
	RemovePackage(info models.Package, output io.Writer) error
	InstallPackage(info models.Package, output io.Writer) error
	InstallTap(tapName string, output io.Writer) error
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

	cmd := brewCommand("--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	s.brewVersion = strings.TrimSpace(string(output))
	return s.brewVersion, nil
}

// UpdateHomebrew updates the Homebrew package manager by running the `brew update` command.
func (s *BrewService) UpdateHomebrew() error {
	cmd := brewCommand("update")
	return cmd.Run()
}

// UpdateAllPackages upgrades all outdated packages.
func (s *BrewService) UpdateAllPackages(output io.Writer) error {
	cmd := brewCommand("upgrade") // #nosec G204
	return ExecuteCommand(cmd, output)
}

// UpdatePackage upgrades a specific package.
func (s *BrewService) UpdatePackage(info models.Package, output io.Writer) error {
	var cmd *exec.Cmd
	if info.Type == models.PackageTypeCask {
		cmd = brewCommand("upgrade", "--cask", info.Name) // #nosec G204
	} else {
		cmd = brewCommand("upgrade", info.Name) // #nosec G204
	}
	return ExecuteCommand(cmd, output)
}

// RemovePackage uninstalls a package.
func (s *BrewService) RemovePackage(info models.Package, output io.Writer) error {
	var cmd *exec.Cmd
	if info.Type == models.PackageTypeCask {
		cmd = brewCommand("uninstall", "--cask", info.Name) // #nosec G204
	} else {
		cmd = brewCommand("uninstall", info.Name) // #nosec G204
	}
	return ExecuteCommand(cmd, output)
}

// InstallPackage installs a package.
func (s *BrewService) InstallPackage(info models.Package, output io.Writer) error {
	var cmd *exec.Cmd
	if info.Type == models.PackageTypeCask {
		cmd = brewCommand("install", "--cask", info.Name) // #nosec G204
	} else {
		cmd = brewCommand("install", info.Name) // #nosec G204
	}
	return ExecuteCommand(cmd, output)
}

// InstallTap installs a Homebrew tap, trusting it for Homebrew 6+ tap trust enforcement.
// The --force flag marks the tap as trusted, which is required in Homebrew 6.0.0+
// and is safely ignored in older versions.
func (s *BrewService) InstallTap(tapName string, output io.Writer) error {
	cmd := brewCommand("tap", "--force", tapName) // #nosec G204
	return ExecuteCommand(cmd, output)
}

// IsTapInstalled checks if a tap is already installed.
func (s *BrewService) IsTapInstalled(tapName string) bool {
	cmd := brewCommand("tap")
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
