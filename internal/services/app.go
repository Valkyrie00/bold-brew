package services

import (
	"bbrew/internal/models"
	"fmt"
	"github.com/gdamore/tcell/v2"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rivo/tview"
	"strings"
)

var (
	AppName    = "Bold Brew"
	AppVersion = "0.0.1"
)

type AppServiceInterface interface {
	GetApp() *tview.Application
	Boot() (err error)
	BuildApp()
}

type AppService struct {
	app *tview.Application

	// Data
	packages          *[]models.Formula
	filteredPackages  *[]models.Formula
	showOnlyInstalled bool
	brewVersion       string

	// Services
	BrewService       BrewServiceInterface
	CommandService    CommandServiceInterface
	SelfUpdateService SelfUpdateServiceInterface
	LayoutService     LayoutServiceInterface
}

var NewAppService = func() AppServiceInterface {
	return &AppService{
		app:               tview.NewApplication(), // Initialize the application
		packages:          new([]models.Formula),
		filteredPackages:  new([]models.Formula),
		showOnlyInstalled: false, // Default to show all packages
		brewVersion:       "-",

		// Services
		BrewService:       NewBrewService(),
		CommandService:    NewCommandService(),
		SelfUpdateService: NewSelfUpdateService(),
		LayoutService:     NewLayoutService(),
	}
}

func (s *AppService) GetApp() *tview.Application {
	return s.app
}

func (s *AppService) Boot() (err error) {
	if err = s.BrewService.LoadAllFormulae(); err != nil {
		return fmt.Errorf("failed to load Homebrew formulae: %v", err)
	}

	s.packages = s.BrewService.GetAllFormulae()
	*s.filteredPackages = *s.packages

	if s.brewVersion, err = s.BrewService.GetCurrentBrewVersion(); err != nil {
		return fmt.Errorf("failed to get Homebrew version: %v", err)
	}

	return nil
}

func (s *AppService) search(searchText string) {
	var filteredList []models.Formula
	uniquePackages := make(map[string]bool)

	// Determine the source list based on the current filter state
	sourceList := s.packages
	if s.showOnlyInstalled {
		sourceList = &[]models.Formula{}
		for _, info := range *s.packages {
			if len(info.Installed) > 0 {
				*sourceList = append(*sourceList, info)
			}
		}
	}

	if searchText == "" {
		// Reset to the appropriate list when the search string is empty
		filteredList = *sourceList
	} else {
		// Apply the search filter
		for _, info := range *sourceList {
			if strings.Contains(strings.ToLower(info.Name), strings.ToLower(searchText)) ||
				strings.Contains(strings.ToLower(info.Description), strings.ToLower(searchText)) {
				if !uniquePackages[info.Name] {
					filteredList = append(filteredList, info)
					uniquePackages[info.Name] = true
				}
			}
		}
	}

	*s.filteredPackages = filteredList
	s.setResults(s.filteredPackages)
}

func (s *AppService) setDetails(info *models.Formula) {
	if info != nil {
		installedVersion := "Not installed"
		packagePrefix := "-"
		installedOnRequest := false
		if len(info.Installed) > 0 {
			if info.Installed[0].Version == info.Versions.Stable {
				installedVersion = info.Installed[0].Version
			} else {
				installedVersion = fmt.Sprintf("[orange]%s[-]", info.Installed[0].Version)
			}
			packagePrefix, _ = s.BrewService.GetPrefixPath(info.Name)
			installedOnRequest = info.Installed[0].InstalledOnRequest
		}

		dependencies := strings.Join(info.Dependencies, ", ")
		if dependencies == "" {
			dependencies = "None"
		}

		generalInfo := fmt.Sprintf(
			"[blue]Name:[-] %s\n[blue]Description:[-] %s\n[blue]Homepage:[-] %s\n[blue]License:[-] %s\n[blue]Tap:[-] %s",
			info.FullName, info.Description, info.Homepage, info.License, info.Tap,
		)

		installInfo := fmt.Sprintf(
			"[blue]Installed:[-] %s\n[blue]Available Version:[-] %s\n[blue]Install Path:[-] %s\n[blue]Dependencies:[-] %s\n[blue]Installed On Request:[-] %t\n[blue]Outdated:[-] %t\n",
			installedVersion, info.Versions.Stable, packagePrefix, dependencies, installedOnRequest, info.Outdated,
		)

		s.LayoutService.GetDetailsView().SetText(
			fmt.Sprintf("%s\n\n%s", generalInfo, installInfo),
		)
		return
	}

	s.LayoutService.GetDetailsView().SetText("")
}

func (s *AppService) forceRefreshResults() {
	s.app.QueueUpdateDraw(func() {
		_ = s.BrewService.LoadAllFormulae()
		s.search(s.LayoutService.GetSearchField().GetText())
		s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetResultTable()) //TODO: da capire se rimuovere
	})
}

func (s *AppService) setResults(data *[]models.Formula) {
	headers := []string{"Name", "Description", "Version"}
	s.LayoutService.GetResultTable().Clear()

	for i, header := range headers {
		s.LayoutService.GetResultTable().SetCell(0, i, tview.NewTableCell(header).
			SetTextColor(tcell.ColorBlue).
			SetAlign(tview.AlignLeft).
			SetSelectable(false).
			SetExpansion(1))
	}

	for i, info := range *data {
		version := info.Versions.Stable
		if len(info.Installed) > 0 && info.Installed[0].Version != info.Versions.Stable {
			version = fmt.Sprintf("(%s) < %s", info.Installed[0].Version, info.Versions.Stable)
		}

		nameCell := tview.NewTableCell(info.Name).SetSelectable(true)
		if len(info.Installed) > 0 {
			nameCell.SetTextColor(tcell.ColorGreen)
		}

		versionCell := tview.NewTableCell(version).SetSelectable(true)
		if version != "" && len(info.Installed) > 0 && info.Installed[0].Version < info.Versions.Stable {
			versionCell.SetTextColor(tcell.ColorOrange)
		}

		s.LayoutService.GetResultTable().SetCell(i+1, 0, nameCell)
		s.LayoutService.GetResultTable().SetCell(i+1, 1, tview.NewTableCell(info.Description).SetSelectable(true))
		s.LayoutService.GetResultTable().SetCell(i+1, 2, versionCell)
	}

	// Update the details view with the first item in the list
	if len(*data) > 0 {
		s.LayoutService.GetResultTable().Select(1, 0)
		s.LayoutService.GetResultTable().ScrollToBeginning()
		s.setDetails(&(*data)[0])

		// Update the filter counter
		s.LayoutService.UpdateFilterCounterView(len(*s.packages), len(*s.filteredPackages))
		return
	}

	s.setDetails(nil)
}

func (s *AppService) BuildApp() {
	// Evaluate if there is a new version available
	latestVersion, err := s.SelfUpdateService.CheckForUpdates()
	if err == nil && latestVersion != AppVersion {
		AppVersion = fmt.Sprintf("%s ([orange]Update available: %s[-])", AppVersion, latestVersion)
	}

	// Build the layout
	s.LayoutService.SetHeaderView(AppName, AppVersion, s.brewVersion)
	s.LayoutService.SetLegendView()
	s.LayoutService.SetDetailsView()
	s.LayoutService.SetBuildOutputView()
	s.LayoutService.SetFilterCounterView()

	// Result table section
	tableSelectionChangedFunc := func(row, column int) {
		if row > 0 && row-1 < len(*s.filteredPackages) {
			s.setDetails(&(*s.filteredPackages)[row-1])
		}
	}
	s.LayoutService.SetResultTable(tableSelectionChangedFunc)

	// Search field section
	inputDoneFunc := func(key tcell.Key) {
		if key == tcell.KeyEnter || key == tcell.KeyEscape {
			s.app.SetFocus(s.LayoutService.GetResultTable())
		}
	}
	changedFunc := func(text string) {
		s.search(s.LayoutService.GetSearchField().GetText())
	}
	s.LayoutService.SetSearchField(inputDoneFunc, changedFunc)

	// Set the grid layout (final step)
	s.LayoutService.SetGrid()

	// Add key event handler and set the root view
	s.app.SetInputCapture(s.handleKeyEventInput)
	s.app.SetRoot(s.LayoutService.GetGrid(), true)
	s.app.SetFocus(s.LayoutService.GetResultTable())

	// Fill the table with the initial data
	s.setResults(s.packages)
}
