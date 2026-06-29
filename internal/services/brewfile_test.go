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

func TestParseBrewfileWithTaps_FileNotFound(t *testing.T) {
	_, err := parseBrewfileWithTaps("/nonexistent/path/Brewfile")
	if err == nil {
		t.Error("expected error for nonexistent file")
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
