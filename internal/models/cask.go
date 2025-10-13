package models

// Cask represents a Homebrew cask (GUI application).
type Cask struct {
	Token                 string             `json:"token"`
	FullToken             string             `json:"full_token"`
	OldTokens             []string           `json:"old_tokens"`
	Tap                   string             `json:"tap"`
	Name                  []string           `json:"name"`
	Description           string             `json:"desc"`
	Homepage              string             `json:"homepage"`
	URL                   string             `json:"url"`
	Version               string             `json:"version"`
	Installed             *string            `json:"installed"`      // Null if not installed, version string if installed
	InstalledTime         *int64             `json:"installed_time"` // Unix timestamp
	Outdated              bool               `json:"outdated"`
	SHA256                string             `json:"sha256"`
	Deprecated            bool               `json:"deprecated"`
	DeprecationDate       interface{}        `json:"deprecation_date"`
	DeprecationReason     interface{}        `json:"deprecation_reason"`
	Disabled              bool               `json:"disabled"`
	DisableDate           interface{}        `json:"disable_date"`
	DisableReason         interface{}        `json:"disable_reason"`
	TapGitHead            string             `json:"tap_git_head"`
	RubySourcePath        string             `json:"ruby_source_path"`
	RubySourceChecksum    RubySourceChecksum `json:"ruby_source_checksum"`
	Analytics90dRank      int                // Internal: Populated from analytics
	Analytics90dDownloads int                // Internal: Populated from analytics
	LocallyInstalled      bool               `json:"-"` // Internal flag
	IsCask                bool               `json:"-"` // Internal flag to distinguish from formulae
}
