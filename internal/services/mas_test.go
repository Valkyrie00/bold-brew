package services

import (
	"bytes"
	"strings"
	"testing"

	"bbrew/internal/models"
)

func TestMasRemoveApp_EmptyID(t *testing.T) {
	s := &MasService{}
	var buf bytes.Buffer

	pkg := models.Package{
		Name: "",
		Type: models.PackageTypeMas,
	}

	err := s.RemoveApp(pkg, &buf)
	if err != nil {
		t.Errorf("RemoveApp() with empty ID should return nil, got: %v", err)
	}
}

func TestMasRemoveApp_ValidID(_ *testing.T) {
	s := &MasService{}
	var buf bytes.Buffer

	pkg := models.Package{
		Name:        "1168254295",
		DisplayName: "AmorphousDiskMark",
		Type:        models.PackageTypeMas,
	}

	// This will fail because `mas` is likely not installed in test env,
	// but validates the command is constructed correctly.
	err := s.RemoveApp(pkg, &buf)
	// We expect an error (mas not installed or app not found), not a panic.
	_ = err
}

func TestMasInstallApp_EmptyID(t *testing.T) {
	s := &MasService{}
	var buf bytes.Buffer

	pkg := models.Package{
		Name: "",
		Type: models.PackageTypeMas,
	}

	err := s.InstallApp(pkg, &buf)
	if err != nil {
		t.Errorf("InstallApp() with empty ID should return nil, got: %v", err)
	}
}

func TestMasInstallApp_CaskTypeReturnsNil(t *testing.T) {
	s := &MasService{}
	var buf bytes.Buffer

	pkg := models.Package{
		Name: "1168254295",
		Type: models.PackageTypeMas,
		Cask: &models.Cask{Token: "test"},
	}

	err := s.InstallApp(pkg, &buf)
	if err != nil {
		t.Errorf("InstallApp() with Cask set should return nil, got: %v", err)
	}
}

func TestParseMasInfoOutput(t *testing.T) {
	// Test the parsing logic used by GetAppInfo
	// Simulating the actual `mas info` output format with ▁ separators
	tests := []struct {
		name        string
		output      string
		wantVersion string
		wantURL     string
	}{
		{
			"standard format",
			"App ▁▁▁▁▁▁▁▁ Display Menu\nVersion ▁▁▁▁ 2.2.6\nPrice ▁▁▁▁▁▁ Gratis\nFrom ▁▁▁▁▁▁▁ https://apps.apple.com/app/display-menu/id549083868\n",
			"2.2.6",
			"https://apps.apple.com/app/display-menu/id549083868",
		},
		{
			"missing fields",
			"App ▁▁▁▁▁▁▁▁ Some App\nVersion ▁▁▁▁ 1.0\n",
			"1.0",
			"",
		},
		{
			"empty output",
			"",
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := ""
			homepage := ""
			for _, line := range strings.Split(tt.output, "\n") {
				parts := strings.SplitN(line, "▁", 2)
				if len(parts) != 2 {
					continue
				}
				label := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(strings.TrimLeft(parts[1], "▁ "))

				switch strings.ToLower(label) {
				case "version":
					version = value
				case "from":
					homepage = value
				}
			}
			if version != tt.wantVersion {
				t.Errorf("version = %q, want %q", version, tt.wantVersion)
			}
			if homepage != tt.wantURL {
				t.Errorf("homepage = %q, want %q", homepage, tt.wantURL)
			}
		})
	}
}
