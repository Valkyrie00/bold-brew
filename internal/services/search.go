package services

import (
	"bbrew/internal/models"
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// search filters the packages based on the search text and the current filter state.
func (s *AppService) search(searchText string, scrollToTop bool) {
	var filteredList []models.Package
	uniquePackages := make(map[string]bool)

	// Determine the source list based on the current filter state
	// If Brewfile mode is active, use brewfilePackages as the base source
	sourceList := s.packages
	if s.IsBrewfileMode() {
		sourceList = s.brewfilePackages
	}

	// Apply active filter on the source list
	sourceList = s.applyFilter(sourceList)

	if searchText == "" {
		// Reset to the appropriate list when the search string is empty
		filteredList = *sourceList
	} else {
		// Apply the search filter
		searchTextLower := strings.ToLower(searchText)
		for _, info := range *sourceList {
			if strings.Contains(strings.ToLower(info.Name), searchTextLower) ||
				strings.Contains(strings.ToLower(info.Description), searchTextLower) {
				if !uniquePackages[info.Name] {
					filteredList = append(filteredList, info)
					uniquePackages[info.Name] = true
				}
			}
		}

		// sort by analytics rank
		sort.Slice(filteredList, func(i, j int) bool {
			if filteredList[i].Analytics90dRank == 0 {
				return false
			}
			if filteredList[j].Analytics90dRank == 0 {
				return true
			}
			return filteredList[i].Analytics90dRank < filteredList[j].Analytics90dRank
		})
	}

	*s.filteredPackages = filteredList
	s.setResults(s.filteredPackages, scrollToTop)
}

// applyFilter filters packages based on the active filter type.
func (s *AppService) applyFilter(sourceList *[]models.Package) *[]models.Package {
	if s.activeFilter == FilterNone {
		return sourceList
	}

	filteredSource := &[]models.Package{}
	for _, info := range *sourceList {
		include := false
		switch s.activeFilter {
		case FilterInstalled:
			include = info.LocallyInstalled
		case FilterOutdated:
			include = info.LocallyInstalled && info.Outdated
		case FilterLeaves:
			include = info.LocallyInstalled && info.InstalledOnRequest
		case FilterCasks:
			include = info.Type == models.PackageTypeCask
		}
		if include {
			*filteredSource = append(*filteredSource, info)
		}
	}
	return filteredSource
}

// forceRefreshResults forces a refresh of the Homebrew formulae and cask data and updates the results in the UI.
func (s *AppService) forceRefreshResults() {
	// Use cached API data (fast) - only installed status needs refresh
	_ = s.dataProvider.SetupData(false)
	s.packages = s.dataProvider.GetPackages()

	// If in Brewfile mode, load tap packages and verify installed status
	if s.IsBrewfileMode() {
		s.fetchTapPackages()
		_ = s.loadBrewfilePackages() // Gets fresh installed status via FetchInstalledCaskNames/FormulaNames
		*s.filteredPackages = *s.brewfilePackages
	} else {
		// For non-Brewfile mode, get fresh installed status
		installedCasks := s.dataProvider.FetchInstalledCaskNames()
		installedFormulae := s.dataProvider.FetchInstalledFormulaNames()
		for i := range *s.packages {
			pkg := &(*s.packages)[i]
			if pkg.Type == models.PackageTypeCask {
				pkg.LocallyInstalled = installedCasks[pkg.Name]
			} else {
				pkg.LocallyInstalled = installedFormulae[pkg.Name]
			}
		}
		*s.filteredPackages = *s.packages
	}

	s.app.QueueUpdateDraw(func() {
		s.search(s.layout.GetSearch().Field().GetText(), false)
	})
}

// setResults updates the results table with the provided data and optionally scrolls to the top.
func (s *AppService) setResults(data *[]models.Package, scrollToTop bool) {
	s.layout.GetTable().Clear()
	s.layout.GetTable().SetTableHeaders("Type", "Name", "Version", "Description", "Downloads")

	for i, info := range *data {
		// Type cell with escaped brackets
		typeTag := tview.Escape("[F]") // Formula
		if info.Type == models.PackageTypeCask {
			typeTag = tview.Escape("[C]") // Cask
		}
		typeCell := tview.NewTableCell(typeTag).SetSelectable(true).SetAlign(tview.AlignLeft)

		// Version handling
		version := info.Version

		// Name cell
		nameCell := tview.NewTableCell(info.Name).SetSelectable(true)
		if info.LocallyInstalled {
			nameCell.SetTextColor(tcell.ColorGreen)
		}

		// Version cell
		versionCell := tview.NewTableCell(version).SetSelectable(true)
		if info.LocallyInstalled && info.Outdated {
			versionCell.SetTextColor(tcell.ColorOrange)
		}

		// Downloads cell
		downloadsCell := tview.NewTableCell(fmt.Sprintf("%d", info.Analytics90dDownloads)).SetSelectable(true).SetAlign(tview.AlignRight)

		// Set cells with new column order: Type, Name, Version, Description, Downloads
		s.layout.GetTable().View().SetCell(i+1, 0, typeCell.SetExpansion(0))
		s.layout.GetTable().View().SetCell(i+1, 1, nameCell.SetExpansion(0))
		s.layout.GetTable().View().SetCell(i+1, 2, versionCell.SetExpansion(0))
		s.layout.GetTable().View().SetCell(i+1, 3, tview.NewTableCell(info.Description).SetSelectable(true).SetExpansion(1))
		s.layout.GetTable().View().SetCell(i+1, 4, downloadsCell.SetExpansion(0))
	}

	// Update the details view with the first item in the list
	if len(*data) > 0 && scrollToTop {
		s.layout.GetTable().View().Select(1, 0)
		s.layout.GetTable().View().ScrollToBeginning()
		s.layout.GetDetails().SetContent(&(*data)[0])
	} else if len(*data) == 0 {
		s.layout.GetDetails().SetContent(nil) // Clear details if no results
	}

	// Update the filter counter
	// In Brewfile mode, show total Brewfile packages instead of all packages
	totalCount := len(*s.packages)
	if s.IsBrewfileMode() {
		totalCount = len(*s.brewfilePackages)
	}
	s.layout.GetSearch().UpdateCounter(totalCount, len(*s.filteredPackages))
}
