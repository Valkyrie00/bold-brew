package services

import (
	"context"
	"encoding/json"
	"fmt"
)

type SelfUpdateServiceInterface interface {
	CheckForUpdates(ctx context.Context) (string, error)
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

// CheckForUpdates checks for the latest version of the Bold Brew package using Homebrew.
// It queries "bbrew" without a tap prefix so it resolves to whichever source the user
// installed from (homebrew-core or the tap), ensuring the notification matches what
// the user can actually upgrade to.
func (s *SelfUpdateService) CheckForUpdates(ctx context.Context) (string, error) {
	cmd := brewCommandContext(ctx, "info", "--json=v1", "bbrew")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("context cancelled")
		}
		return "", fmt.Errorf("failed to fetch latest version: %v", err)
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
