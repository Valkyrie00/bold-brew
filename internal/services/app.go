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

	packages          *[]models.Package
	filteredPackages  *[]models.Package
	showOnlyInstalled bool
	showOnlyOutdated  bool
	showOnlyLeaves    bool
	showOnlyCasks     bool
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

		packages:          new([]models.Package),
		filteredPackages:  new([]models.Package),
		showOnlyInstalled: false,
		showOnlyOutdated:  false,
		showOnlyLeaves:    false,
		showOnlyCasks:     false,
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

func (s *AppService) Boot() (err error) {
	if s.brewVersion, err = s.brewService.GetBrewVersion(); err != nil {
		return fmt.Errorf("failed to get Homebrew version: %v", err)
	}
	return nil
}

// loadInitialData streams Homebrew packages asynchronously.
func (s *AppService) loadInitialData() {
	s.layout.GetNotifier().ShowWarning("Loading Homebrew formulae...")

	pkgChan, errChan := s.brewService.StreamPackages(false)

	loader := &packageLoader{
		app:            s.app,
		appService:     s,
		batchSize:      200,
		updateInterval: 150 * time.Millisecond,
	}

	loader.streamPackages(pkgChan, errChan)

	go s.updateHomeBrew()
}

type packageLoader struct {
	app            *tview.Application
	appService     *AppService
	batchSize      int
	updateInterval time.Duration
}

func (l *packageLoader) streamPackages(pkgChan <-chan models.Package, errChan <-chan error) {
	var packages []models.Package
	batch := make([]models.Package, 0, l.batchSize)

	ticker := time.NewTicker(l.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case pkg, ok := <-pkgChan:
			if !ok {
				if len(batch) > 0 {
					packages = append(packages, batch...)
					l.updateUI(packages)
				}
				l.showSuccess(len(packages))
				return
			}
			batch = append(batch, pkg)

		case <-ticker.C:
			if len(batch) > 0 {
				packages = append(packages, batch...)
				l.updateUI(packages)
				batch = batch[:0]
			}

		case err, ok := <-errChan:
			if ok && err != nil {
				l.showError(err)
			}
		}
	}
}

func (l *packageLoader) updateUI(packages []models.Package) {
	pkgCopy := make([]models.Package, len(packages))
	copy(pkgCopy, packages)

	l.app.QueueUpdateDraw(func() {
		l.appService.packages = &pkgCopy
		*l.appService.filteredPackages = pkgCopy
		l.appService.setResults(l.appService.packages, false)
		l.appService.layout.GetNotifier().ShowWarning(fmt.Sprintf("Loading... %d packages", len(pkgCopy)))
	})
}

func (l *packageLoader) showSuccess(count int) {
	l.app.QueueUpdateDraw(func() {
		l.appService.layout.GetNotifier().ShowSuccess(fmt.Sprintf("Loaded %d packages", count))
	})
}

func (l *packageLoader) showError(err error) {
	l.app.QueueUpdateDraw(func() {
		l.appService.layout.GetNotifier().ShowError(fmt.Sprintf("Error loading: %v", err))
	})
}

// updateHomeBrew updates the Homebrew formulae and refreshes the results in the UI.
func (s *AppService) updateHomeBrew() {
	s.layout.GetNotifier().ShowWarning("Updating Homebrew formulae...")
	if err := s.brewService.UpdateHomebrew(); err != nil {
		s.layout.GetNotifier().ShowError("Could not update Homebrew formulae")
		return
	}
	s.layout.GetNotifier().ShowSuccess("Homebrew formulae updated successfully")
	s.forceRefreshResults()
}

// search filters the packages based on the search text and the current filter state.
func (s *AppService) search(searchText string, scrollToTop bool) {
	var filteredList []models.Package
	uniquePackages := make(map[string]bool)

	// Determine the source list based on the current filter state
	sourceList := s.packages
	if s.showOnlyInstalled && !s.showOnlyOutdated {
		sourceList = &[]models.Package{}
		for _, info := range *s.packages {
			if info.LocallyInstalled {
				*sourceList = append(*sourceList, info)
			}
		}
	}

	if s.showOnlyOutdated {
		sourceList = &[]models.Package{}
		for _, info := range *s.packages {
			if info.LocallyInstalled && info.Outdated {
				*sourceList = append(*sourceList, info)
			}
		}
	}

	if s.showOnlyLeaves {
		sourceList = &[]models.Package{}
		for _, info := range *s.packages {
			if info.LocallyInstalled && info.InstalledOnRequest {
				*sourceList = append(*sourceList, info)
			}
		}
	}

	if s.showOnlyCasks {
		sourceList = &[]models.Package{}
		for _, info := range *s.packages {
			if info.Type == models.PackageTypeCask {
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

// forceRefreshResults forces a refresh of the Homebrew formulae and cask data and updates the results in the UI.
func (s *AppService) forceRefreshResults() {
	_ = s.brewService.SetupData(true)
	s.packages = s.brewService.GetPackages()
	*s.filteredPackages = *s.packages

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

	go s.loadInitialData()
}
