package services

import (
	"testing"

	"bbrew/internal/models"
)

func TestGetTapPackages_SkipsMasEntries(t *testing.T) {
	d := &DataProvider{}

	entries := []models.BrewfileEntry{
		{Name: "AmorphousDiskMark", IsMas: true, MasID: "1168254295"},
		{Name: "Display Menu", IsMas: true, MasID: "549083868"},
		{Name: "org.gnome.Calculator", IsFlatpak: true},
	}

	existingPackages := make(map[string]models.Package)

	result, err := d.GetTapPackages(entries, existingPackages, false)
	if err != nil {
		t.Fatalf("GetTapPackages() error: %v", err)
	}

	// All entries are MAS or Flatpak, so nothing should be returned
	if len(result) != 0 {
		t.Errorf("GetTapPackages() returned %d packages, want 0 (MAS and Flatpak should be skipped)", len(result))
		for _, pkg := range result {
			t.Errorf("  unexpected package: %s (type: %s)", pkg.Name, pkg.Type)
		}
	}
}

func TestGetTapPackages_MixedEntries(t *testing.T) {
	d := &DataProvider{}

	entries := []models.BrewfileEntry{
		{Name: "AmorphousDiskMark", IsMas: true, MasID: "1168254295"},
		{Name: "org.gnome.Calculator", IsFlatpak: true},
		{Name: "wget"},
	}

	// Provide wget as an existing package so it gets returned without brew info fetch
	existingPackages := map[string]models.Package{
		"wget": {Name: "wget", Type: models.PackageTypeFormula, Description: "Internet file retriever"},
	}

	result, err := d.GetTapPackages(entries, existingPackages, false)
	if err != nil {
		t.Fatalf("GetTapPackages() error: %v", err)
	}

	// Only wget should be returned (MAS and Flatpak are skipped)
	if len(result) != 1 {
		t.Fatalf("GetTapPackages() returned %d packages, want 1", len(result))
	}

	if result[0].Name != "wget" {
		t.Errorf("GetTapPackages()[0].Name = %q, want %q", result[0].Name, "wget")
	}
}

func TestGetFlatpakPackages_Basic(t *testing.T) {
	d := &DataProvider{}

	entries := []models.BrewfileEntry{
		{Name: "com.spotify.Client", IsFlatpak: true},
		{Name: "org.mozilla.firefox", IsFlatpak: true},
		{Name: "wget"},
		{Name: "AmorphousDiskMark", IsMas: true, MasID: "1168254295"},
	}

	installedIDs := map[string]bool{
		"com.spotify.Client": true,
	}

	metadata := map[string]models.Package{
		"com.spotify.Client": {
			Name:        "com.spotify.Client",
			DisplayName: "Spotify",
			Description: "Music streaming service",
			Version:     "1.2.3",
		},
	}

	result, err := d.GetFlatpakPackages(entries, installedIDs, metadata)
	if err != nil {
		t.Fatalf("GetFlatpakPackages() error: %v", err)
	}

	// Should only return flatpak entries (2), not formula or MAS
	if len(result) != 2 {
		t.Fatalf("GetFlatpakPackages() returned %d packages, want 2", len(result))
	}

	// Verify Spotify is marked as installed with metadata
	if result[0].Name != "com.spotify.Client" {
		t.Errorf("result[0].Name = %q, want %q", result[0].Name, "com.spotify.Client")
	}
	if !result[0].LocallyInstalled {
		t.Error("Spotify should be marked as locally installed")
	}
	if result[0].DisplayName != "Spotify" {
		t.Errorf("result[0].DisplayName = %q, want %q", result[0].DisplayName, "Spotify")
	}
}
