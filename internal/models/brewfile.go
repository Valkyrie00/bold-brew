package models

// BrewfileEntry represents a single entry from a Brewfile
type BrewfileEntry struct {
	Name   string
	IsCask bool
}
