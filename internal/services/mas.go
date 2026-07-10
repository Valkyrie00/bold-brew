package services

import (
	"io"
	"os/exec"
	"strings"

	"bbrew/internal/models"
)

// MasAppInfo holds metadata retrieved from `mas info`.
type MasAppInfo struct {
	Version  string
	Homepage string
}

// MasServiceInterface defines the contract for Mac App Store operations.
type MasServiceInterface interface {
	IsMasInstalled() bool
	GetInstalledApps() (map[string]bool, error)
	GetAppInfo(appID string) (*MasAppInfo, error)
	InstallApp(info models.Package, output io.Writer) error
	RemoveApp(info models.Package, output io.Writer) error
}

// MasService implements MasServiceInterface.
type MasService struct{}

// NewMasService creates a new instance of MasService.
var NewMasService = func() MasServiceInterface {
	return &MasService{}
}

// IsMasInstalled checks if the mas binary exists in the PATH.
func (s *MasService) IsMasInstalled() bool {
	_, err := exec.LookPath("mas")
	return err == nil
}

// GetInstalledApps returns a map of installed Mac App Store app IDs.
// Uses `mas list` which outputs lines like: "1234567890 App Name (1.0)"
func (s *MasService) GetInstalledApps() (map[string]bool, error) {
	installed := make(map[string]bool)

	cmd := exec.Command("mas", "list")
	output, err := cmd.Output()
	if err != nil {
		return installed, nil
	}

	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			installed[parts[0]] = true
		}
	}

	return installed, nil
}

// GetAppInfo retrieves metadata for a Mac App Store app via `mas info`.
// Output is a table with "▁" separators, e.g.:
//
//	Version ▁▁▁▁ 2.2.6
//	From ▁▁▁▁▁▁▁ https://apps.apple.com/...
func (s *MasService) GetAppInfo(appID string) (*MasAppInfo, error) {
	cmd := exec.Command("mas", "info", appID) // #nosec G204
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	info := &MasAppInfo{}
	for _, line := range strings.Split(string(output), "\n") {
		// Each line: "Label ▁▁▁ Value"
		parts := strings.SplitN(line, "▁", 2)
		if len(parts) != 2 {
			continue
		}
		label := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(strings.TrimLeft(parts[1], "▁ "))

		switch strings.ToLower(label) {
		case "version":
			info.Version = value
		case "from":
			info.Homepage = value
		}
	}

	return info, nil
}

// InstallApp installs a Mac App Store app by its ID.
func (s *MasService) InstallApp(info models.Package, output io.Writer) error {
	masID := ""
	if info.Cask != nil {
		return nil
	}
	// The MasID is stored in the package's Name field when Type is PackageTypeMas
	// and also available via the DisplayName for UI purposes.
	// We use a convention: Name = masID for mas packages
	masID = info.Name
	if masID == "" {
		return nil
	}
	cmd := exec.Command("mas", "install", masID) // #nosec G204
	return ExecuteCommand(cmd, output)
}

// RemoveApp uninstalls a Mac App Store app by its ID.
func (s *MasService) RemoveApp(info models.Package, output io.Writer) error {
	masID := info.Name
	if masID == "" {
		return nil
	}
	cmd := exec.Command("mas", "uninstall", masID) // #nosec G204
	return ExecuteCommand(cmd, output)
}
