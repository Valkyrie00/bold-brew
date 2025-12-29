package services

import (
	"bbrew/internal/models"
	"bbrew/internal/ui"
	"bbrew/internal/ui/theme"
	"context"
	"fmt"
	"os"
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
	SetBrewfilePath(path string)
	IsBrewfileMode() bool
	GetBrewfilePackages() *[]models.Package
}

// AppService manages the application state, Homebrew integration, and UI components.
type AppService struct {
	app    *tview.Application
	theme  *theme.Theme
	layout ui.LayoutInterface

	packages          *[]models.Package
	filteredPackages  *[]models.Package
	showOnlyInstalled bool
	showOnlyOutdated  bool
	showOnlyLeaves    bool
	showOnlyCasks     bool
	brewVersion       string

	// Brewfile support
	brewfilePath     string
	brewfilePackages *[]models.Package
	brewfileTaps     []string // Taps required by the Brewfile

	brewService       BrewServiceInterface
	dataProvider      DataProviderInterface // Direct access for Brewfile operations
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

		packages:          new([]models.Package),
		filteredPackages:  new([]models.Package),
		showOnlyInstalled: false,
		showOnlyOutdated:  false,
		showOnlyLeaves:    false,
		showOnlyCasks:     false,
		brewVersion:       "-",

		brewfilePath:     "",
		brewfilePackages: new([]models.Package),
	}

	// Initialize services
	s.dataProvider = NewDataProvider()
	s.brewService = NewBrewService()
	s.ioService = NewIOService(s, s.brewService)
	s.selfUpdateService = NewSelfUpdateService()

	return s
}

func (s *AppService) GetApp() *tview.Application             { return s.app }
func (s *AppService) GetLayout() ui.LayoutInterface          { return s.layout }
func (s *AppService) SetBrewfilePath(path string)            { s.brewfilePath = path }
func (s *AppService) IsBrewfileMode() bool                   { return s.brewfilePath != "" }
func (s *AppService) GetBrewfilePackages() *[]models.Package { return s.brewfilePackages }

// Boot initializes the application by setting up Homebrew and loading formulae data.
func (s *AppService) Boot() (err error) {
	if s.brewVersion, err = s.brewService.GetBrewVersion(); err != nil {
		// This error is critical, as we need Homebrew to function
		return fmt.Errorf("failed to get Homebrew version: %v", err)
	}

	// Load Homebrew data from cache for fast startup
	// Installation status might be stale but will be refreshed in background by updateHomeBrew()
	if err = s.dataProvider.SetupData(false); err != nil {
		// Log error but don't fail - app can work with empty/partial data
		fmt.Fprintf(os.Stderr, "Warning: failed to load Homebrew data (will retry in background): %v\n", err)
	}

	// Initialize packages and filteredPackages
	s.packages = s.dataProvider.GetPackages()
	*s.filteredPackages = *s.packages

	// If Brewfile is specified, parse it and filter packages
	if s.IsBrewfileMode() {
		if err = s.loadBrewfilePackages(); err != nil {
			return fmt.Errorf("failed to load Brewfile: %v", err)
		}
	}

	return nil
}

// updateHomeBrew updates the Homebrew formulae and refreshes the results in the UI.
func (s *AppService) updateHomeBrew() {
	s.app.QueueUpdateDraw(func() {
		s.layout.GetNotifier().ShowWarning("Updating Homebrew formulae...")
	})
	if err := s.brewService.UpdateHomebrew(); err != nil {
		s.app.QueueUpdateDraw(func() {
			s.layout.GetNotifier().ShowError("Could not update Homebrew formulae")
		})
		return
	}
	// Clear loading message and update results
	s.app.QueueUpdateDraw(func() {
		s.layout.GetNotifier().ShowSuccess("Homebrew formulae updated successfully")
	})
	s.forceRefreshResults()
}

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

	// Apply filters on the base source list (either all packages or Brewfile packages)
	if s.showOnlyInstalled && !s.showOnlyOutdated {
		filteredSource := &[]models.Package{}
		for _, info := range *sourceList {
			if info.LocallyInstalled {
				*filteredSource = append(*filteredSource, info)
			}
		}
		sourceList = filteredSource
	}

	if s.showOnlyOutdated {
		filteredSource := &[]models.Package{}
		for _, info := range *sourceList {
			if info.LocallyInstalled && info.Outdated {
				*filteredSource = append(*filteredSource, info)
			}
		}
		sourceList = filteredSource
	}

	if s.showOnlyLeaves {
		filteredSource := &[]models.Package{}
		for _, info := range *sourceList {
			if info.LocallyInstalled && info.InstalledOnRequest {
				*filteredSource = append(*filteredSource, info)
			}
		}
		sourceList = filteredSource
	}

	if s.showOnlyCasks {
		filteredSource := &[]models.Package{}
		for _, info := range *sourceList {
			if info.Type == models.PackageTypeCask {
				*filteredSource = append(*filteredSource, info)
			}
		}
		sourceList = filteredSource
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

// forceRefreshResults forces a refresh of the Homebrew formulae and cask data and updates the results in the UI.
func (s *AppService) forceRefreshResults() {
	// Use cached API data (fast) - only installed status needs refresh
	_ = s.dataProvider.SetupData(false)
	s.packages = s.dataProvider.GetPackages()

	// If in Brewfile mode, load tap packages and verify installed status
	if s.IsBrewfileMode() {
		s.fetchTapPackages()
		_ = s.loadBrewfilePackages() // Gets fresh installed status via GetInstalledCaskNames/FormulaNames
		*s.filteredPackages = *s.brewfilePackages
	} else {
		// For non-Brewfile mode, get fresh installed status
		installedCasks := s.dataProvider.GetInstalledCaskNames()
		installedFormulae := s.dataProvider.GetInstalledFormulaNames()
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
	s.layout.GetTable().SetTableHeaders("Type", "Name", "Version", "Description", "↓ (90d)")

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

// BuildApp builds the application layout, sets up event handlers, and initializes the UI components.
func (s *AppService) BuildApp() {
	// Build the layout
	s.layout.Setup()

	// Update header and enable Brewfile mode features if needed
	headerName := AppName
	if s.IsBrewfileMode() {
		headerName = fmt.Sprintf("%s [Brewfile Mode]", AppName)
		s.layout.GetSearch().Field().SetLabel("Search (Brewfile): ")
		s.ioService.EnableBrewfileMode() // Add Install All action
	}
	s.layout.GetHeader().Update(headerName, AppVersion, s.brewVersion)

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
				headerName := AppName
				if s.IsBrewfileMode() {
					headerName = fmt.Sprintf("%s [Brewfile Mode]", AppName)
				}
				s.layout.GetHeader().Update(headerName, AppVersion, s.brewVersion)
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

	// Start background tasks: install taps first (if Brewfile mode), then update Homebrew
	go func() {
		// In Brewfile mode, install missing taps first
		if s.IsBrewfileMode() && len(s.brewfileTaps) > 0 {
			s.installBrewfileTapsAtStartup()
		}
		// Then update Homebrew (which will reload all data including new taps)
		s.updateHomeBrew()
	}()

	// Set initial results based on mode
	if s.IsBrewfileMode() {
		*s.filteredPackages = *s.brewfilePackages // Sync filteredPackages
		s.setResults(s.brewfilePackages, true)    // Show only Brewfile packages
	} else {
		s.setResults(s.packages, true) // Show all packages
	}
}
