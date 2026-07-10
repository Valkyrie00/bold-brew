package services

import (
	"os"
	"testing"
)

func TestExtractQuotedValue(t *testing.T) {
	tests := []struct {
		line   string
		want   string
		wantOK bool
	}{
		{`brew "wget"`, "wget", true},
		{`cask "firefox"`, "firefox", true},
		{`tap "homebrew/cask"`, "homebrew/cask", true},
		{`flatpak "org.gnome.Calculator"`, "org.gnome.Calculator", true},
		{`brew "wget", args: ["HEAD"]`, "wget", true},
		{`no quotes here`, "", false},
		{`missing end "quote`, "", false},
		{``, "", false},
	}

	for _, tt := range tests {
		got, ok := extractQuotedValue(tt.line)
		if ok != tt.wantOK {
			t.Errorf("extractQuotedValue(%q) ok = %v, want %v", tt.line, ok, tt.wantOK)
		}
		if got != tt.want {
			t.Errorf("extractQuotedValue(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestParseBrewfileWithTaps(t *testing.T) {
	content := `# My Brewfile
tap "homebrew/cask"
tap "homebrew/core"

brew "wget"
brew "curl"
brew "jq"

cask "firefox"
cask "visual-studio-code"

flatpak "org.gnome.Calculator"
`
	tmpFile := createTempBrewfile(t, content)

	result, err := parseBrewfileWithTaps(tmpFile)
	if err != nil {
		t.Fatalf("parseBrewfileWithTaps() error: %v", err)
	}

	if len(result.Taps) != 2 {
		t.Errorf("Taps count = %d, want 2", len(result.Taps))
	}
	if result.Taps[0] != "homebrew/cask" {
		t.Errorf("Taps[0] = %q, want %q", result.Taps[0], "homebrew/cask")
	}

	if len(result.Packages) != 6 {
		t.Errorf("Packages count = %d, want 6", len(result.Packages))
	}

	// Check formula entries
	if result.Packages[0].Name != "wget" || result.Packages[0].IsCask || result.Packages[0].IsFlatpak {
		t.Errorf("Packages[0] = %+v, want wget formula", result.Packages[0])
	}

	// Check cask entries
	if result.Packages[3].Name != "firefox" || !result.Packages[3].IsCask {
		t.Errorf("Packages[3] = %+v, want firefox cask", result.Packages[3])
	}

	// Check flatpak entries
	if result.Packages[5].Name != "org.gnome.Calculator" || !result.Packages[5].IsFlatpak {
		t.Errorf("Packages[5] = %+v, want org.gnome.Calculator flatpak", result.Packages[5])
	}
}

func TestParseBrewfileWithTaps_EmptyFile(t *testing.T) {
	tmpFile := createTempBrewfile(t, "")

	result, err := parseBrewfileWithTaps(tmpFile)
	if err != nil {
		t.Fatalf("parseBrewfileWithTaps() error: %v", err)
	}

	if len(result.Taps) != 0 {
		t.Errorf("Taps count = %d, want 0", len(result.Taps))
	}
	if len(result.Packages) != 0 {
		t.Errorf("Packages count = %d, want 0", len(result.Packages))
	}
}

func TestParseBrewfileWithTaps_CommentsOnly(t *testing.T) {
	content := `# This is a comment
# Another comment
`
	tmpFile := createTempBrewfile(t, content)

	result, err := parseBrewfileWithTaps(tmpFile)
	if err != nil {
		t.Fatalf("parseBrewfileWithTaps() error: %v", err)
	}

	if len(result.Packages) != 0 {
		t.Errorf("Packages count = %d, want 0", len(result.Packages))
	}
}

func TestParseBrewfileWithTaps_MasEntries(t *testing.T) {
	content := `brew "mas"
mas "AmorphousDiskMark", id: 1168254295
mas "Display Menu", id: 549083868
`
	tmpFile := createTempBrewfile(t, content)

	result, err := parseBrewfileWithTaps(tmpFile)
	if err != nil {
		t.Fatalf("parseBrewfileWithTaps() error: %v", err)
	}

	// Should have 3 packages: mas formula + 2 mas apps
	if len(result.Packages) != 3 {
		t.Fatalf("Packages count = %d, want 3", len(result.Packages))
	}

	// First entry is the mas formula itself
	if result.Packages[0].Name != "mas" || result.Packages[0].IsMas {
		t.Errorf("Packages[0] = %+v, want mas formula", result.Packages[0])
	}

	// Second entry is a mas app
	if result.Packages[1].Name != "AmorphousDiskMark" || !result.Packages[1].IsMas {
		t.Errorf("Packages[1] = %+v, want AmorphousDiskMark mas app", result.Packages[1])
	}
	if result.Packages[1].MasID != "1168254295" {
		t.Errorf("Packages[1].MasID = %q, want %q", result.Packages[1].MasID, "1168254295")
	}

	// Third entry
	if result.Packages[2].MasID != "549083868" {
		t.Errorf("Packages[2].MasID = %q, want %q", result.Packages[2].MasID, "549083868")
	}
}

func TestExtractMasID(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{`mas "App", id: 1234567`, "1234567"},
		{`mas "App", id:1234567`, "1234567"},
		{`mas "App", id: 549083868`, "549083868"},
		{`mas "App"`, ""},
		{`mas "App", id: `, ""},
	}

	for _, tt := range tests {
		got := extractMasID(tt.line)
		if got != tt.want {
			t.Errorf("extractMasID(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestParseBrewfileWithTaps_FileNotFound(t *testing.T) {
	_, err := parseBrewfileWithTaps("/nonexistent/path/Brewfile")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseBrewfileWithTaps_MasEntriesNotInPackageMap(t *testing.T) {
	content := `brew "wget"
cask "firefox"
mas "AmorphousDiskMark", id: 1168254295
mas "Display Menu", id: 549083868
`
	tmpFile := createTempBrewfile(t, content)

	result, err := parseBrewfileWithTaps(tmpFile)
	if err != nil {
		t.Fatalf("parseBrewfileWithTaps() error: %v", err)
	}

	// Build the packageMap the same way loadBrewfilePackages does
	packageMap := make(map[string]string)
	for _, entry := range result.Packages {
		if entry.IsMas {
			continue
		}
		if entry.IsCask {
			packageMap[entry.Name] = "cask"
		} else if entry.IsFlatpak {
			packageMap[entry.Name] = "flatpak"
		} else {
			packageMap[entry.Name] = "formula"
		}
	}

	// MAS app names must NOT appear in the package map
	if _, exists := packageMap["AmorphousDiskMark"]; exists {
		t.Error("MAS entry 'AmorphousDiskMark' should not be in packageMap")
	}
	if _, exists := packageMap["Display Menu"]; exists {
		t.Error("MAS entry 'Display Menu' should not be in packageMap")
	}

	// Regular entries should be present
	if _, exists := packageMap["wget"]; !exists {
		t.Error("formula 'wget' should be in packageMap")
	}
	if _, exists := packageMap["firefox"]; !exists {
		t.Error("cask 'firefox' should be in packageMap")
	}
}

func TestParseBrewfileWithTaps_MasEntriesExcludedFromTapEntries(t *testing.T) {
	content := `brew "wget"
mas "AmorphousDiskMark", id: 1168254295
mas "Hand Mirror", id: 1502839586
`
	tmpFile := createTempBrewfile(t, content)

	result, err := parseBrewfileWithTaps(tmpFile)
	if err != nil {
		t.Fatalf("parseBrewfileWithTaps() error: %v", err)
	}

	// Simulate the tap entry collection logic from loadBrewfilePackages:
	// entries not found in main list, excluding flatpak and mas
	foundPackages := make(map[string]bool)
	foundPackages["wget"] = true // pretend wget was found in Homebrew catalog

	var tapEntries []string
	for _, entry := range result.Packages {
		if !foundPackages[entry.Name] && !entry.IsFlatpak && !entry.IsMas {
			tapEntries = append(tapEntries, entry.Name)
		}
	}

	// MAS entries must not end up as tap entries
	for _, name := range tapEntries {
		if name == "AmorphousDiskMark" || name == "Hand Mirror" {
			t.Errorf("MAS entry %q should not appear in tapEntries", name)
		}
	}
}

func createTempBrewfile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "Brewfile")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}
