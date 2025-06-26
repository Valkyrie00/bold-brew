package services

import (
	"bbrew/internal/models"
	"bbrew/internal/ui"
	"bbrew/internal/ui/theme"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	AppName    = "Bold Brew"
	AppVersion = "0.0.1"
)

type AppServiceInterface interface {
	GetApp() *tview.Application
	GetLayout() ui.LayoutInterface
	Boot() (err error)
	BuildApp()
}

// AppService manages the application state, Homebrew integration, and UI components.
type AppService struct {
	app    *tview.Application
	theme  *theme.Theme
	layout ui.LayoutInterface

	packages          *[]models.Formula
	filteredPackages  *[]models.Formula
	showOnlyInstalled bool
	showOnlyOutdated  bool
	brewVersion       string

	brewService       BrewServiceInterface
	selfUpdateService SelfUpdateServiceInterface
	ioService         IOServiceInterface
}

// NewAppService creates a new instance of AppService with initialized components.
var NewAppService = func() AppServiceInterface {
	app := tview.NewApplication()
	themeService := theme.NewTheme()
	layout := ui.NewLayout(themeService)

	s := &AppService{
		app:    app,
		theme:  themeService,
		layout: layout,

		packages:          new([]models.Formula),
		filteredPackages:  new([]models.Formula),
		showOnlyInstalled: false,
		showOnlyOutdated:  false,
		brewVersion:       "-",
	}

	// Initialize services
	s.ioService = NewIOService(s)
	s.brewService = NewBrewService()
	s.selfUpdateService = NewSelfUpdateService()

	return s
}

func (s *AppService) GetApp() *tview.Application    { return s.app }
func (s *AppService) GetLayout() ui.LayoutInterface { return s.layout }

// Boot initializes the application by setting up Homebrew and loading formulae data.
func (s *AppService) Boot() (err error) {
	if s.brewVersion, err = s.brewService.GetBrewVersion(); err != nil {
		// This error is critical, as we need Homebrew to function
		return fmt.Errorf("failed to get Homebrew version: %v", err)
	}

	// Download and parse Homebrew formulae data
	if err = s.brewService.SetupData(false); err != nil {
		return fmt.Errorf("failed to load Homebrew formulae: %v", err)
	}

	// Initialize packages and filteredPackages
	s.packages = s.brewService.GetFormulae()
	*s.filteredPackages = *s.packages
	return nil
}

// updateHomeBrew updates the Homebrew formulae and refreshes the results in the UI.
func (s *AppService) updateHomeBrew() {
	s.layout.GetNotifier().ShowWarning("Updating Homebrew formulae...")
	if err := s.brewService.UpdateHomebrew(); err != nil {
		s.layout.GetNotifier().ShowError("Could not update Homebrew formulae")
		return
	}
	// Clear loading message and update results
	s.layout.GetNotifier().ShowSuccess("Homebrew formulae updated successfully")
	s.forceRefreshResults()
}

// search filters the packages based on the search text and the current filter state.
func (s *AppService) search(searchText string, scrollToTop bool) {
	var filteredList []models.Formula
	uniquePackages := make(map[string]bool)

	// Determine the source list based on the current filter state
	sourceList := s.packages
	if s.showOnlyInstalled && !s.showOnlyOutdated {
		sourceList = &[]models.Formula{}
		for _, info := range *s.packages {
			if info.LocallyInstalled {
				*sourceList = append(*sourceList, info)
			}
		}
	}

	if s.showOnlyOutdated {
		sourceList = &[]models.Formula{}
		for _, info := range *s.packages {
			if info.LocallyInstalled && info.Outdated {
				*sourceList = append(*sourceList, info)
			}
		}
	}

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

// forceRefreshResults forces a refresh of the Homebrew formulae data and updates the results in the UI.
func (s *AppService) forceRefreshResults() {
	_ = s.brewService.SetupData(true)
	s.packages = s.brewService.GetFormulae()
	*s.filteredPackages = *s.packages

	s.app.QueueUpdateDraw(func() {
		s.search(s.layout.GetSearch().Field().GetText(), false)
	})
}

// setResults updates the results table with the provided data and optionally scrolls to the top.
func (s *AppService) setResults(data *[]models.Formula, scrollToTop bool) {
	s.layout.GetTable().Clear()
	s.layout.GetTable().SetTableHeaders("Name", "Version", "Description", "↓ (90d)")

	for i, info := range *data {
		version := info.Versions.Stable
		if len(info.Installed) > 0 {
			// Check if the installed version is the same as the stable version (handle revisions)
			if strings.HasPrefix(info.Installed[0].Version, info.Versions.Stable) {
				version = info.Installed[0].Version
			} else if info.Installed[0].Version != info.Versions.Stable {
				version = fmt.Sprintf("%s → %s", info.Installed[0].Version, info.Versions.Stable)
			}
		}

		nameCell := tview.NewTableCell(info.Name).SetSelectable(true)
		if info.LocallyInstalled {
			nameCell.SetTextColor(tcell.ColorGreen)
		}

		versionCell := tview.NewTableCell(version).SetSelectable(true)
		if info.LocallyInstalled && info.Outdated {
			versionCell.SetTextColor(tcell.ColorOrange)
		}

		downloadsCell := tview.NewTableCell(fmt.Sprintf("%d", info.Analytics90dDownloads)).SetSelectable(true).SetAlign(tview.AlignRight)

		s.layout.GetTable().View().SetCell(i+1, 0, nameCell.SetExpansion(0))
		s.layout.GetTable().View().SetCell(i+1, 1, versionCell.SetExpansion(0))
		s.layout.GetTable().View().SetCell(i+1, 2, tview.NewTableCell(info.Description).SetSelectable(true).SetExpansion(1))
		s.layout.GetTable().View().SetCell(i+1, 3, downloadsCell.SetExpansion(0))
	}

	// Update the details view with the first item in the list
	if len(*data) > 0 {
		if scrollToTop {
			s.layout.GetTable().View().Select(1, 0)
			s.layout.GetTable().View().ScrollToBeginning()
			s.layout.GetDetails().SetContent(&(*data)[0])
		}

		// Update the filter counter
		s.layout.GetSearch().UpdateCounter(len(*s.packages), len(*s.filteredPackages))
		return
	}

	s.layout.GetDetails().SetContent(nil) // Clear details if no results
}

// BuildApp builds the application layout, sets up event handlers, and initializes the UI components.
func (s *AppService) BuildApp() {
	// Build the layout
	s.layout.Setup()
	s.layout.GetHeader().Update(AppName, AppVersion, s.brewVersion)

	// Evaluate if there is a new version available
	// This is done in a goroutine to avoid blocking the UI during startup
	// In the future, this could be replaced with a more sophisticated update check, and update
	// the user if a new version is available instantly instead of waiting for the next app start
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		if latestVersion, err := s.selfUpdateService.CheckForUpdates(ctx); err == nil && latestVersion != AppVersion {
			s.app.QueueUpdateDraw(func() {
				AppVersion = fmt.Sprintf("%s ([orange]New Version Available: %s[-])", AppVersion, latestVersion)
				s.layout.GetHeader().Update(AppName, AppVersion, s.brewVersion)
			})
		}
	}()

	// Table handler to update the details view when a table row is selected
	tableSelectionChangedFunc := func(row, _ int) {
		if row > 0 && row-1 < len(*s.filteredPackages) {
			s.layout.GetDetails().SetContent(&(*s.filteredPackages)[row-1])
		}
	}
	s.layout.GetTable().View().SetSelectionChangedFunc(tableSelectionChangedFunc)

	// Search input handlers
	inputDoneFunc := func(key tcell.Key) {
		if key == tcell.KeyEnter || key == tcell.KeyEscape {
			s.app.SetFocus(s.layout.GetTable().View()) // Set focus back to the table on Enter or Escape
		}
	}
	changedFunc := func(text string) { // Each time the search input changes
		s.search(text, true) // Perform search and scroll to top
	}
	s.layout.GetSearch().SetHandlers(inputDoneFunc, changedFunc)

	// Add key event handler
	s.app.SetInputCapture(s.ioService.HandleKeyEventInput)

	// Set the root of the application to the layout's root and focus on the table view
	s.app.SetRoot(s.layout.Root(), true)
	s.app.SetFocus(s.layout.GetTable().View())

	go s.updateHomeBrew()          // Update Async the Homebrew formulae
	s.setResults(s.packages, true) // Set the results
}
