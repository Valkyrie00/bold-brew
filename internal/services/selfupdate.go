package services

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type SelfUpdateServiceInterface interface {
	CheckForUpdates() (string, error)
}

type SelfUpdateService struct{}

type boldBrewStatusInfo struct {
	Versions struct {
		Stable string `json:"stable"`
	} `json:"versions"`
}

var NewSelfUpdateService = func() SelfUpdateServiceInterface {
	return &SelfUpdateService{}
}

func (s *SelfUpdateService) CheckForUpdates() (string, error) {
	latestVersion, err := s.getLatestVersionFromTap()
	if err != nil {
		return "", err
	}
	return latestVersion, nil
}

func (s *SelfUpdateService) getLatestVersionFromTap() (string, error) {
	cmd := exec.Command("brew", "info", "--json=v1", "valkyrie00/bbrew/bbrew")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest version from tap: %v", err)
	}

	var info []boldBrewStatusInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return "", fmt.Errorf("failed to parse version info: %v", err)
	}

	if len(info) == 0 {
		return "", fmt.Errorf("no version information found")
	}

	return info[0].Versions.Stable, nil
}
