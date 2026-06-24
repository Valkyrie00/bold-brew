package models

// BrewfileEntry represents a single entry from a Brewfile
type BrewfileEntry struct {
	Name      string
	IsCask    bool
	IsFlatpak bool
}

// BrewfileResult contains all parsed entries from a Brewfile
type BrewfileResult struct {
	Taps     []string        // List of taps to install
	Packages []BrewfileEntry // List of packages (formulae and casks)
}
