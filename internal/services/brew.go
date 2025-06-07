package services

import (
	"bbrew/internal/models"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const FormulaeAPIURL = "https://formulae.brew.sh/api/formula.json"
const AnalyticsAPIURL = "https://formulae.brew.sh/api/analytics/install-on-request/90d.json"

type BrewServiceInterface interface {
	GetPrefixPath() (path string)
	GetFormulae() (formulae *[]models.Formula)
	SetupData(forceDownload bool) (err error)
	GetBrewVersion() (version string, err error)
	UpdateHomebrew() error
}

type BrewService struct {
	// Package lists
	all       *[]models.Formula
	installed *[]models.Formula
	remote    *[]models.Formula
	analytics map[string]models.AnalyticsItem

	brewVersion string
	prefixPath  string
}

var NewBrewService = func() BrewServiceInterface {
	return &BrewService{
		all:       new([]models.Formula),
		installed: new([]models.Formula),
		remote:    new([]models.Formula),
	}
}

func (s *BrewService) GetPrefixPath() (path string) {
	if s.prefixPath != "" {
		return s.prefixPath
	}

	cmd := exec.Command("brew", "--prefix")
	output, err := cmd.Output()
	if err != nil {
		s.prefixPath = "Unknown"
		return
	}

	s.prefixPath = strings.TrimSpace(string(output))
	return s.prefixPath
}

func (s *BrewService) GetFormulae() (formulae *[]models.Formula) {
	packageMap := make(map[string]models.Formula)

	// Add remote packages to the map if they don't already exist
	for _, formula := range *s.remote {
		if _, exists := packageMap[formula.Name]; !exists {
			packageMap[formula.Name] = formula
		}
	}

	// Add installed packages to the map
	for _, formula := range *s.installed {
		packageMap[formula.Name] = formula
	}

	*s.all = make([]models.Formula, 0, len(packageMap))
	for _, formula := range packageMap {
		// Merge analytics data if available
		if a, exists := s.analytics[formula.Name]; exists && a.Number > 0 {
			downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
			formula.Analytics90dRank = a.Number
			formula.Analytics90dDownloads = downloads
		}

		*s.all = append(*s.all, formula)
	}

	// Sort the list by name
	sort.Slice(*s.all, func(i, j int) bool {
		return (*s.all)[i].Name < (*s.all)[j].Name
	})

	return s.all
}

func (s *BrewService) SetupData(forceDownload bool) (err error) {
	if err = s.loadInstalled(); err != nil {
		return err
	}

	if err = s.loadRemote(forceDownload); err != nil {
		return err
	}

	if err = s.loadAnalytics(); err != nil {
		return err
	}

	return nil
}

func (s *BrewService) loadInstalled() (err error) {
	cmd := exec.Command("brew", "info", "--json=v1", "--installed")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	*s.installed = make([]models.Formula, 0)
	err = json.Unmarshal(output, &s.installed)
	if err != nil {
		return err
	}

	// Mark all installed packages as locally installed
	for i := range *s.installed {
		(*s.installed)[i].LocallyInstalled = true
	}

	return nil
}

func (s *BrewService) loadRemote(forceDownload bool) (err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	bbrewDir := filepath.Join(homeDir, ".bbrew") // TODO: Move to config
	formulaFile := filepath.Join(bbrewDir, "formula.json")
	if _, err := os.Stat(bbrewDir); os.IsNotExist(err) {
		if err := os.MkdirAll(bbrewDir, 0755); err != nil {
			return err
		}
	}

	// Check if we should use the cached file
	if !forceDownload {
		if _, err := os.Stat(formulaFile); err == nil {
			data, err := os.ReadFile(formulaFile)
			if err == nil {
				*s.remote = make([]models.Formula, 0)
				if err := json.Unmarshal(data, &s.remote); err == nil {
					return nil
				}
			}
		}
	}

	resp, err := http.Get(FormulaeAPIURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	*s.remote = make([]models.Formula, 0)
	err = json.Unmarshal(body, s.remote)
	if err != nil {
		return err
	}

	// Cache the remote formulae data
	_ = os.WriteFile(formulaFile, body, 0600)
	return nil
}

func (s *BrewService) loadAnalytics() (err error) {
	resp, err := http.Get(AnalyticsAPIURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	analytics := models.Analytics{}
	err = json.NewDecoder(resp.Body).Decode(&analytics)
	if err != nil {
		return err
	}

	analyticsByFormula := map[string]models.AnalyticsItem{}
	for _, f := range analytics.Items {
		analyticsByFormula[f.Formula] = f
	}

	s.analytics = analyticsByFormula
	return nil
}

func (s *BrewService) GetBrewVersion() (version string, err error) {
	if s.brewVersion != "" {
		return s.brewVersion, nil
	}

	cmd := exec.Command("brew", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	s.brewVersion = strings.TrimSpace(string(output))
	return s.brewVersion, nil
}

func (s *BrewService) UpdateHomebrew() error {
	cmd := exec.Command("brew", "update")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
