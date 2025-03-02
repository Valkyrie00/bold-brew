package services

import (
	"bbrew/internal/models"
	"encoding/json"
	"net/http"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var prefixPathCache = make(map[string]string)

type BrewServiceInterface interface {
	GetPrefixPath(packageName string) (path string, err error)
	GetAllFormulae() (formulae *[]models.Formula)
	LoadAllFormulae() (err error)
	GetCurrentBrewVersion() (version string, err error)
}

type BrewService struct {
	cache     sync.Mutex
	all       *[]models.Formula
	installed *[]models.Formula
	remote    *[]models.Formula
	analytics map[string]models.AnalyticsItem
}

var NewBrewService = func() BrewServiceInterface {
	return &BrewService{
		cache:     sync.Mutex{},
		all:       new([]models.Formula),
		installed: new([]models.Formula),
		remote:    new([]models.Formula),
	}
}

func (s *BrewService) GetPrefixPath(packageName string) (path string, err error) {
	s.cache.Lock()
	defer s.cache.Unlock()

	var found bool
	if path, found = prefixPathCache[packageName]; found {
		return path, nil
	}

	cmd := exec.Command("brew", "--prefix", packageName)
	output, err := cmd.Output()
	if err != nil {
		return "Unknown", err
	}

	path = strings.TrimSpace(string(output))
	prefixPathCache[packageName] = path
	return path, nil
}

func (s *BrewService) GetAllFormulae() (formulae *[]models.Formula) {
	return s.all
}

func (s *BrewService) LoadAllFormulae() (err error) {
	_ = s.loadInstalled()
	_ = s.loadRemote()
	_ = s.loadAnalytics()

	packageMap := make(map[string]models.Formula)

	// Add installed packages to the map
	for _, formula := range *s.installed {
		packageMap[formula.Name] = formula
	}

	// Add remote packages to the map if they don't already exist
	for _, formula := range *s.remote {
		if _, exists := packageMap[formula.Name]; !exists {
			packageMap[formula.Name] = formula
		}
	}

	*s.all = make([]models.Formula, 0, len(packageMap))
	for _, formula := range packageMap {
		// patch analytics info
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

	return nil
}

func (s *BrewService) loadRemote() (err error) {
	resp, err := http.Get("https://formulae.brew.sh/api/formula.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	*s.remote = make([]models.Formula, 0)
	err = json.NewDecoder(resp.Body).Decode(&s.remote)
	if err != nil {
		return err
	}

	return nil
}

func (s *BrewService) loadAnalytics() (err error) {
	resp, err := http.Get("https://formulae.brew.sh/api/analytics/install-on-request/90d.json")
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

func (s *BrewService) GetCurrentBrewVersion() (version string, err error) {
	cmd := exec.Command("brew", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}
