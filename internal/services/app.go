package services

import (
	"bbrew/internal/models"
	"bbrew/internal/ui"
	"bbrew/internal/ui/theme"
	"context"
	"fmt"
	"os"
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
	inputService      InputServiceInterface
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
	s.inputService = NewInputService(s, s.brewService)
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

// BuildApp builds the application layout, sets up event handlers, and initializes the UI components.
func (s *AppService) BuildApp() {
	// Build the layout
	s.layout.Setup()

	// Update header and enable Brewfile mode features if needed
	headerName := AppName
	if s.IsBrewfileMode() {
		headerName = fmt.Sprintf("%s [Brewfile Mode]", AppName)
		s.layout.GetSearch().Field().SetLabel("Search (Brewfile): ")
		s.inputService.EnableBrewfileMode() // Add Install All action
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
	s.app.SetInputCapture(s.inputService.HandleKeyEventInput)

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
