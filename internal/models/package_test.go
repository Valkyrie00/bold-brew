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

func TestNewPackageFromCask_OutdatedByVersionMismatch(t *testing.T) {
	installedVersion := "1.0.0"
	c := &Cask{
		Token:            "vorta",
		Name:             []string{"Vorta"},
		Version:          "1.1.0",
		Installed:        &installedVersion,
		Outdated:         false, // Homebrew didn't flag it
		LocallyInstalled: true,
	}

	pkg := NewPackageFromCask(c)
	if !pkg.Outdated {
		t.Error("Outdated = false, want true (version mismatch: 1.0.0 vs 1.1.0)")
	}
}

func TestNewPackageFromCask_NotOutdatedWhenAutoUpdates(t *testing.T) {
	installedVersion := "4.8.3" // stale receipt: the app self-updated past this
	c := &Cask{
		Token:            "macfuse",
		Name:             []string{"macFUSE"},
		Version:          "5.2.0",
		Installed:        &installedVersion,
		Outdated:         false, // Homebrew skips auto_updates casks by default
		AutoUpdates:      true,
		LocallyInstalled: true,
	}

	pkg := NewPackageFromCask(c)
	if pkg.Outdated {
		t.Error("Outdated = true, want false (auto_updates cask: receipt mismatch is not a real update)")
	}
}

func TestNewPackageFromCask_OutdatedWhenAutoUpdatesFlaggedByHomebrew(t *testing.T) {
	installedVersion := "1.0.0"
	c := &Cask{
		Token:            "self-updating-app",
		Name:             []string{"Self Updating App"},
		Version:          "2.0.0",
		Installed:        &installedVersion,
		Outdated:         true, // e.g. brew outdated --greedy would report this
		AutoUpdates:      true,
		LocallyInstalled: true,
	}

	pkg := NewPackageFromCask(c)
	if !pkg.Outdated {
		t.Error("Outdated = false, want true (Homebrew's own flag must still be respected)")
	}
}

func TestNewPackageFromCask_NotOutdatedWhenVersionLatest(t *testing.T) {
	installedVersion := "latest"
	c := &Cask{
		Token:            "some-app",
		Name:             []string{"Some App"},
		Version:          "latest",
		Installed:        &installedVersion,
		Outdated:         false,
		LocallyInstalled: true,
	}

	pkg := NewPackageFromCask(c)
	if pkg.Outdated {
		t.Error("Outdated = true, want false (version is 'latest')")
	}
}

func TestNewPackageFromCask_NotOutdatedWhenVersionsMatch(t *testing.T) {
	installedVersion := "2.0.0"
	c := &Cask{
		Token:            "up-to-date",
		Name:             []string{"Up To Date"},
		Version:          "2.0.0",
		Installed:        &installedVersion,
		Outdated:         false,
		LocallyInstalled: true,
	}

	pkg := NewPackageFromCask(c)
	if pkg.Outdated {
		t.Error("Outdated = true, want false (versions match)")
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
