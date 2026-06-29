package services

import (
	"os"
	"strings"
	"testing"

	"bbrew/internal/models"
)

func TestExportBrewfile(t *testing.T) {
	packages := []models.Package{
		{Name: "wget", Type: models.PackageTypeFormula, LocallyInstalled: true},
		{Name: "curl", Type: models.PackageTypeFormula, LocallyInstalled: true},
		{Name: "firefox", Type: models.PackageTypeCask, LocallyInstalled: true},
		{Name: "not-installed", Type: models.PackageTypeFormula, LocallyInstalled: false},
		{Name: "visual-studio-code", Type: models.PackageTypeCask, LocallyInstalled: true},
	}

	s := &AppService{
		packages: &packages,
	}

	path, err := s.ExportBrewfile()
	if err != nil {
		t.Fatalf("ExportBrewfile() error: %v", err)
	}
	defer os.Remove(path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	content := string(data)

	// Should contain installed formulae
	if !strings.Contains(content, `brew "curl"`) {
		t.Error("missing brew curl")
	}
	if !strings.Contains(content, `brew "wget"`) {
		t.Error("missing brew wget")
	}

	// Should contain installed casks
	if !strings.Contains(content, `cask "firefox"`) {
		t.Error("missing cask firefox")
	}
	if !strings.Contains(content, `cask "visual-studio-code"`) {
		t.Error("missing cask visual-studio-code")
	}

	// Should NOT contain non-installed packages
	if strings.Contains(content, "not-installed") {
		t.Error("should not include non-installed packages")
	}

	// Should be sorted alphabetically
	curlIdx := strings.Index(content, `brew "curl"`)
	wgetIdx := strings.Index(content, `brew "wget"`)
	if curlIdx > wgetIdx {
		t.Error("formulae should be sorted alphabetically (curl before wget)")
	}

	firefoxIdx := strings.Index(content, `cask "firefox"`)
	vscodeIdx := strings.Index(content, `cask "visual-studio-code"`)
	if firefoxIdx > vscodeIdx {
		t.Error("casks should be sorted alphabetically (firefox before visual-studio-code)")
	}
}

func TestExportBrewfile_WithTaps(t *testing.T) {
	formula := &models.Formula{
		Name:     "font-fira-code",
		FullName: "homebrew/cask-fonts/font-fira-code",
	}

	packages := []models.Package{
		{
			Name:             "font-fira-code",
			Type:             models.PackageTypeFormula,
			LocallyInstalled: true,
			Formula:          formula,
		},
	}

	s := &AppService{
		packages: &packages,
	}

	path, err := s.ExportBrewfile()
	if err != nil {
		t.Fatalf("ExportBrewfile() error: %v", err)
	}
	defer os.Remove(path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, `tap "homebrew/cask-fonts"`) {
		t.Error("missing tap for third-party formula")
	}
}

func TestExportBrewfile_EmptyPackages(t *testing.T) {
	packages := []models.Package{}
	s := &AppService{
		packages: &packages,
	}

	_, err := s.ExportBrewfile()
	if err == nil {
		t.Error("expected error for empty packages")
	}
}

func TestExportBrewfile_NilPackages(t *testing.T) {
	s := &AppService{
		packages: nil,
	}

	_, err := s.ExportBrewfile()
	if err == nil {
		t.Error("expected error for nil packages")
	}
}

func TestExportBrewfile_OnlyNonInstalled(t *testing.T) {
	packages := []models.Package{
		{Name: "wget", Type: models.PackageTypeFormula, LocallyInstalled: false},
		{Name: "curl", Type: models.PackageTypeFormula, LocallyInstalled: false},
	}

	s := &AppService{
		packages: &packages,
	}

	path, err := s.ExportBrewfile()
	if err != nil {
		t.Fatalf("ExportBrewfile() error: %v", err)
	}
	defer os.Remove(path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "brew ") {
		t.Error("should not contain any brew entries for non-installed packages")
	}
}
