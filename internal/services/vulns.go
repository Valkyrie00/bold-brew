package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"bbrew/internal/models"
)

// vulnResult holds the raw JSON scan result for a single formula.
type vulnResult struct {
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Vulnerabilities []models.Vulnerability `json:"vulnerabilities"`
}

// VulnsServiceInterface defines the contract for vulnerability scanning.
type VulnsServiceInterface interface {
	IsAvailable() bool
	ScanPackage(name string, output io.Writer) ([]models.Vulnerability, error)
	GetCachedVulns(name string) ([]models.Vulnerability, bool)
}

// VulnsService wraps the `brew vulns` command for vulnerability scanning.
type VulnsService struct {
	mu        sync.RWMutex
	available *bool // nil = not checked yet
	cache     map[string][]models.Vulnerability
}

var NewVulnsService = func() VulnsServiceInterface {
	return &VulnsService{
		cache: make(map[string][]models.Vulnerability),
	}
}

// resetAvailability clears the cached availability check so it's re-evaluated on next call.
func (s *VulnsService) resetAvailability() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.available = nil
}

// IsAvailable checks if the `brew vulns` command is installed.
func (s *VulnsService) IsAvailable() bool {
	s.mu.RLock()
	if s.available != nil {
		avail := *s.available
		s.mu.RUnlock()
		return avail
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.available != nil {
		return *s.available
	}

	cmd := brewCommand("vulns", "--help")
	err := cmd.Run()
	avail := err == nil
	s.available = &avail
	return avail
}

// ScanPackage runs `brew vulns <name>` streaming real-time output to the panel,
// then fetches JSON for structured caching. This avoids blocking the UI
// while the OSV database is queried.
func (s *VulnsService) ScanPackage(name string, output io.Writer) ([]models.Vulnerability, error) {
	if !s.IsAvailable() {
		return nil, fmt.Errorf("brew vulns is not installed. Install with: brew install homebrew/brew-vulns/brew-vulns")
	}

	// Phase 1: Stream human-readable output in real-time so the user sees progress
	cmd := brewCommand("vulns", name, "-m", "0") // #nosec G204
	scanErr := ExecuteCommand(cmd, output)
	// Exit code 1 = vulnerabilities found (not a failure for us)
	// Exit code 2 = scan error. We still continue to try JSON caching.
	if scanErr != nil {
		fmt.Fprintf(output, "\n")
	}

	// Phase 2: Fetch JSON for structured caching (fast — brew caches the OSV response)
	jsonCmd := brewCommand("vulns", name, "--json") // #nosec G204
	var jsonBuf bytes.Buffer
	jsonCmd.Stdout = &jsonBuf
	_ = jsonCmd.Run() // best-effort; if this fails we just skip caching

	vulns := make([]models.Vulnerability, 0)
	if jsonBuf.Len() > 0 {
		var results []vulnResult
		if err := json.Unmarshal(jsonBuf.Bytes(), &results); err == nil {
			for _, r := range results {
				vulns = append(vulns, r.Vulnerabilities...)
			}
		}
	}

	s.mu.Lock()
	s.cache[name] = vulns
	s.mu.Unlock()

	return vulns, nil
}

// GetCachedVulns returns cached vulnerability data for a package, if available.
func (s *VulnsService) GetCachedVulns(name string) ([]models.Vulnerability, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vulns, ok := s.cache[name]
	return vulns, ok
}

func pluralY(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}
