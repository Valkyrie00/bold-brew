package services

import (
	"bbrew/internal/models"
	"encoding/json"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// GetFormulae retrieves all formulae, merging remote and installed packages.
func (s *BrewService) GetFormulae() *[]models.Formula {
	packageMap := make(map[string]models.Formula)

	for _, formula := range *s.remote {
		if _, exists := packageMap[formula.Name]; !exists {
			packageMap[formula.Name] = formula
		}
	}

	for _, formula := range *s.installed {
		packageMap[formula.Name] = formula
	}

	*s.all = make([]models.Formula, 0, len(packageMap))
	for _, formula := range packageMap {
		if a, exists := s.analytics[formula.Name]; exists && a.Number > 0 {
			downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
			formula.Analytics90dRank = a.Number
			formula.Analytics90dDownloads = downloads
		}
		*s.all = append(*s.all, formula)
	}

	sort.Slice(*s.all, func(i, j int) bool {
		return (*s.all)[i].Name < (*s.all)[j].Name
	})

	return s.all
}

// GetPackages retrieves all packages (formulae + casks), merging remote and installed.
func (s *BrewService) GetPackages() *[]models.Package {
	packageMap := make(map[string]models.Package)

	for _, formula := range *s.remote {
		if _, exists := packageMap[formula.Name]; !exists {
			f := formula
			pkg := models.NewPackageFromFormula(&f)
			if a, exists := s.analytics[formula.Name]; exists && a.Number > 0 {
				downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
				pkg.Analytics90dRank = a.Number
				pkg.Analytics90dDownloads = downloads
			}
			packageMap[formula.Name] = pkg
		}
	}

	for _, formula := range *s.installed {
		f := formula
		pkg := models.NewPackageFromFormula(&f)
		if a, exists := s.analytics[formula.Name]; exists && a.Number > 0 {
			downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
			pkg.Analytics90dRank = a.Number
			pkg.Analytics90dDownloads = downloads
		}
		packageMap[formula.Name] = pkg
	}

	for _, cask := range *s.remoteCasks {
		if _, exists := packageMap[cask.Token]; !exists {
			c := cask
			pkg := models.NewPackageFromCask(&c)
			if a, exists := s.caskAnalytics[cask.Token]; exists && a.Number > 0 {
				downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
				pkg.Analytics90dRank = a.Number
				pkg.Analytics90dDownloads = downloads
			}
			packageMap[cask.Token] = pkg
		}
	}

	for _, cask := range *s.installedCasks {
		c := cask
		pkg := models.NewPackageFromCask(&c)
		if a, exists := s.caskAnalytics[cask.Token]; exists && a.Number > 0 {
			downloads, _ := strconv.Atoi(strings.ReplaceAll(a.Count, ",", ""))
			pkg.Analytics90dRank = a.Number
			pkg.Analytics90dDownloads = downloads
		}
		packageMap[cask.Token] = pkg
	}

	*s.allPackages = make([]models.Package, 0, len(packageMap))
	for _, pkg := range packageMap {
		*s.allPackages = append(*s.allPackages, pkg)
	}

	sort.Slice(*s.allPackages, func(i, j int) bool {
		return (*s.allPackages)[i].Name < (*s.allPackages)[j].Name
	})

	return s.allPackages
}

// IsPackageInstalled checks if a package (formula or cask) is installed by name.
func (s *BrewService) IsPackageInstalled(name string, isCask bool) bool {
	var cmd *exec.Cmd
	if isCask {
		cmd = exec.Command("brew", "list", "--cask", name)
	} else {
		cmd = exec.Command("brew", "list", "--formula", name)
	}
	err := cmd.Run()
	return err == nil
}

// GetInstalledCaskNames returns a map of installed cask names for quick lookup.
func (s *BrewService) GetInstalledCaskNames() map[string]bool {
	result := make(map[string]bool)
	cmd := exec.Command("brew", "list", "--cask")
	output, err := cmd.Output()
	if err != nil {
		return result
	}
	names := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, name := range names {
		if name != "" {
			result[name] = true
		}
	}
	return result
}

// GetInstalledFormulaNames returns a map of installed formula names for quick lookup.
func (s *BrewService) GetInstalledFormulaNames() map[string]bool {
	result := make(map[string]bool)
	cmd := exec.Command("brew", "list", "--formula")
	output, err := cmd.Output()
	if err != nil {
		return result
	}
	names := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, name := range names {
		if name != "" {
			result[name] = true
		}
	}
	return result
}

// getPackageInfoSingle retrieves info for a single package directly.
func (s *BrewService) getPackageInfoSingle(name string, isCask bool) *models.Package {
	var cmd *exec.Cmd
	if isCask {
		cmd = exec.Command("brew", "info", "--json=v2", "--cask", name)
	} else {
		cmd = exec.Command("brew", "info", "--json=v1", name)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	if isCask {
		var response struct {
			Casks []models.Cask `json:"casks"`
		}
		if err := json.Unmarshal(output, &response); err != nil || len(response.Casks) == 0 {
			return nil
		}
		cask := response.Casks[0]
		cask.LocallyInstalled = s.IsPackageInstalled(name, true)
		pkg := models.NewPackageFromCask(&cask)
		return &pkg
	}

	var formulae []models.Formula
	if err := json.Unmarshal(output, &formulae); err != nil || len(formulae) == 0 {
		return nil
	}
	formula := formulae[0]
	formula.LocallyInstalled = s.IsPackageInstalled(name, false)
	pkg := models.NewPackageFromFormula(&formula)
	return &pkg
}

// GetPackagesInfo retrieves package information for multiple packages in a single brew call.
func (s *BrewService) GetPackagesInfo(names []string, isCask bool) map[string]models.Package {
	result := make(map[string]models.Package)
	if len(names) == 0 {
		return result
	}

	var cmd *exec.Cmd
	if isCask {
		args := append([]string{"info", "--json=v2", "--cask"}, names...)
		cmd = exec.Command("brew", args...)
	} else {
		args := append([]string{"info", "--json=v1"}, names...)
		cmd = exec.Command("brew", args...)
	}

	output, err := cmd.Output()
	if err != nil {
		for _, name := range names {
			if pkg := s.getPackageInfoSingle(name, isCask); pkg != nil {
				result[name] = *pkg
			}
		}
		return result
	}

	if isCask {
		var response struct {
			Casks []models.Cask `json:"casks"`
		}
		if err := json.Unmarshal(output, &response); err != nil {
			return result
		}
		for _, cask := range response.Casks {
			c := cask
			c.LocallyInstalled = s.IsPackageInstalled(c.Token, true)
			pkg := models.NewPackageFromCask(&c)
			result[c.Token] = pkg
		}
	} else {
		var formulae []models.Formula
		if err := json.Unmarshal(output, &formulae); err != nil {
			return result
		}
		for _, formula := range formulae {
			f := formula
			f.LocallyInstalled = s.IsPackageInstalled(f.Name, false)
			pkg := models.NewPackageFromFormula(&f)
			result[f.Name] = pkg
		}
	}

	return result
}

