package services

import (
	"bbrew/internal/models"
	"fmt"
	"os"
	"strings"
)

// ParseBrewfile parses a Brewfile and returns a list of packages to be installed.
func (s *BrewService) ParseBrewfile(filepath string) ([]models.BrewfileEntry, error) {
	result, err := s.ParseBrewfileWithTaps(filepath)
	if err != nil {
		return nil, err
	}
	return result.Packages, nil
}

// ParseBrewfileWithTaps parses a Brewfile and returns taps and packages separately.
func (s *BrewService) ParseBrewfileWithTaps(filepath string) (*models.BrewfileResult, error) {
	// #nosec G304 -- filepath is user-provided via CLI flag
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Brewfile: %w", err)
	}

	result := &models.BrewfileResult{
		Taps:     []string{},
		Packages: []models.BrewfileEntry{},
	}
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse tap entries: tap "user/repo"
		if strings.HasPrefix(line, "tap ") {
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start != -1 && end != -1 && start < end {
				tapName := line[start+1 : end]
				result.Taps = append(result.Taps, tapName)
			}
		}

		// Parse brew entries: brew "package-name"
		if strings.HasPrefix(line, "brew ") {
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start != -1 && end != -1 && start < end {
				packageName := line[start+1 : end]
				result.Packages = append(result.Packages, models.BrewfileEntry{
					Name:   packageName,
					IsCask: false,
				})
			}
		}

		// Parse cask entries: cask "package-name"
		if strings.HasPrefix(line, "cask ") {
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start != -1 && end != -1 && start < end {
				packageName := line[start+1 : end]
				result.Packages = append(result.Packages, models.BrewfileEntry{
					Name:   packageName,
					IsCask: true,
				})
			}
		}
	}

	return result, nil
}

