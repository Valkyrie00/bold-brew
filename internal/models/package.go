package models

// PackageType distinguishes between formulae and casks.
type PackageType string

const (
	PackageTypeFormula PackageType = "formula"
	PackageTypeCask    PackageType = "cask"
	PackageTypeFlatpak PackageType = "flatpak"
)

// Package represents a unified view of both Formula and Cask for UI display.
type Package struct {
	// Common fields
	Name                  string      // Formula.Name or Cask.Token
	DisplayName           string      // Formula.FullName or Cask.Name[0]
	Description           string      // desc
	Homepage              string      // homepage
	Version               string      // versions.stable or version
	LocallyInstalled      bool        // Is installed locally
	Outdated              bool        // Needs update
	Type                  PackageType // formula or cask
	Analytics90dRank      int
	Analytics90dDownloads int

	// Original data (for operations)
	Formula *Formula `json:"-"` // nil if Type == cask
	Cask    *Cask    `json:"-"` // nil if Type == formula

	// For leaves filter (only meaningful for formulae)
	InstalledOnRequest bool
}

// NewPackageFromFormula creates a Package from a Formula.
func NewPackageFromFormula(f *Formula) Package {
	installedOnRequest := false
	if len(f.Installed) > 0 {
		installedOnRequest = f.Installed[0].InstalledOnRequest
	}

	return Package{
		Name:                  f.Name,
		DisplayName:           f.FullName,
		Description:           f.Description,
		Homepage:              f.Homepage,
		Version:               f.Versions.Stable,
		LocallyInstalled:      f.LocallyInstalled,
		Outdated:              f.Outdated,
		Type:                  PackageTypeFormula,
		Analytics90dRank:      f.Analytics90dRank,
		Analytics90dDownloads: f.Analytics90dDownloads,
		Formula:               f,
		Cask:                  nil,
		InstalledOnRequest:    installedOnRequest,
	}
}

// NewPackageFromCask creates a Package from a Cask.
func NewPackageFromCask(c *Cask) Package {
	displayName := c.Token
	if len(c.Name) > 0 {
		displayName = c.Name[0]
	}

	return Package{
		Name:                  c.Token,
		DisplayName:           displayName,
		Description:           c.Description,
		Homepage:              c.Homepage,
		Version:               c.Version,
		LocallyInstalled:      c.LocallyInstalled,
		Outdated:              c.Outdated,
		Type:                  PackageTypeCask,
		Analytics90dRank:      c.Analytics90dRank,
		Analytics90dDownloads: c.Analytics90dDownloads,
		Formula:               nil,
		Cask:                  c,
		InstalledOnRequest:    true, // Casks are always explicitly installed
	}
}
