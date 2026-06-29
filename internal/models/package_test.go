package models

import "testing"

func TestNewPackageFromFormula(t *testing.T) {
	f := &Formula{
		Name:        "wget",
		FullName:    "wget",
		Description: "Internet file retriever",
		Homepage:    "https://www.gnu.org/software/wget/",
		Versions:    Versions{Stable: "1.21.4"},
		Installed: []Installed{
			{InstalledOnRequest: true},
		},
		LocallyInstalled:      true,
		Outdated:              false,
		Analytics90dRank:      5,
		Analytics90dDownloads: 100000,
	}

	pkg := NewPackageFromFormula(f)

	if pkg.Name != "wget" {
		t.Errorf("Name = %q, want %q", pkg.Name, "wget")
	}
	if pkg.Type != PackageTypeFormula {
		t.Errorf("Type = %q, want %q", pkg.Type, PackageTypeFormula)
	}
	if pkg.Version != "1.21.4" {
		t.Errorf("Version = %q, want %q", pkg.Version, "1.21.4")
	}
	if !pkg.LocallyInstalled {
		t.Error("LocallyInstalled = false, want true")
	}
	if !pkg.InstalledOnRequest {
		t.Error("InstalledOnRequest = false, want true")
	}
	if pkg.Formula != f {
		t.Error("Formula pointer not preserved")
	}
	if pkg.Cask != nil {
		t.Error("Cask should be nil for formula package")
	}
}

func TestNewPackageFromFormula_NotInstalledOnRequest(t *testing.T) {
	f := &Formula{
		Name:     "openssl",
		FullName: "openssl",
		Versions: Versions{Stable: "3.1.0"},
		Installed: []Installed{
			{InstalledOnRequest: false},
		},
	}

	pkg := NewPackageFromFormula(f)
	if pkg.InstalledOnRequest {
		t.Error("InstalledOnRequest = true, want false (dependency)")
	}
}

func TestNewPackageFromFormula_NoInstalled(t *testing.T) {
	f := &Formula{
		Name:      "not-installed",
		FullName:  "not-installed",
		Versions:  Versions{Stable: "1.0"},
		Installed: nil,
	}

	pkg := NewPackageFromFormula(f)
	if pkg.InstalledOnRequest {
		t.Error("InstalledOnRequest = true, want false (not installed)")
	}
}

func TestNewPackageFromCask(t *testing.T) {
	c := &Cask{
		Token:                 "firefox",
		Name:                  []string{"Mozilla Firefox"},
		Description:           "Web browser",
		Homepage:              "https://www.mozilla.org/firefox/",
		Version:               "120.0",
		LocallyInstalled:      true,
		Outdated:              true,
		Analytics90dRank:      1,
		Analytics90dDownloads: 500000,
	}

	pkg := NewPackageFromCask(c)

	if pkg.Name != "firefox" {
		t.Errorf("Name = %q, want %q", pkg.Name, "firefox")
	}
	if pkg.DisplayName != "Mozilla Firefox" {
		t.Errorf("DisplayName = %q, want %q", pkg.DisplayName, "Mozilla Firefox")
	}
	if pkg.Type != PackageTypeCask {
		t.Errorf("Type = %q, want %q", pkg.Type, PackageTypeCask)
	}
	if !pkg.Outdated {
		t.Error("Outdated = false, want true")
	}
	if !pkg.InstalledOnRequest {
		t.Error("InstalledOnRequest = false, want true (casks always explicit)")
	}
	if pkg.Cask != c {
		t.Error("Cask pointer not preserved")
	}
	if pkg.Formula != nil {
		t.Error("Formula should be nil for cask package")
	}
}

func TestNewPackageFromCask_EmptyName(t *testing.T) {
	c := &Cask{
		Token:   "my-app",
		Name:    []string{},
		Version: "1.0",
	}

	pkg := NewPackageFromCask(c)
	if pkg.DisplayName != "my-app" {
		t.Errorf("DisplayName = %q, want %q (fallback to Token)", pkg.DisplayName, "my-app")
	}
}
