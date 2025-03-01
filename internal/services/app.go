package services

import (
	"bbrew/internal/models"
	"bbrew/internal/ui"
	"bbrew/internal/ui/theme"
	"context"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
	"time"
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
	app    *tview.Application
	layout ui.LayoutInterface
	theme  *theme.Theme

	packages          *[]models.Formula
	filteredPackages  *[]models.Formula
	showOnlyInstalled bool
	brewVersion       string

	BrewService       BrewServiceInterface
	CommandService    CommandServiceInterface
	SelfUpdateService SelfUpdateServiceInterface
}

func NewAppService() AppServiceInterface {
	app := tview.NewApplication()
	themeService := theme.NewTheme()

	appService := &AppService{
		app:    app,
		theme:  themeService,
		layout: ui.NewLayout(themeService),

		packages:          new([]models.Formula),
		filteredPackages:  new([]models.Formula),
		showOnlyInstalled: false,
		brewVersion:       "-",

		BrewService:       NewBrewService(),
		CommandService:    NewCommandService(),
		SelfUpdateService: NewSelfUpdateService(),
	}

	return appService
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

func (s *AppService) updateHomeBrew() {
	s.layout.GetNotifier().ShowWarning("Updating Homebrew formulae...")
	if err := s.CommandService.UpdateHomebrew(); err != nil {
		s.layout.GetNotifier().ShowError("Could not update Homebrew formulae")
		return
	}

	// Clear loading message and update results
	s.layout.GetNotifier().ShowSuccess("Homebrew formulae updated successfully")
	s.forceRefreshResults()
}

func (s *AppService) search(searchText string, scrollToTop bool) {
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
	s.setResults(s.filteredPackages, scrollToTop)
}

func (s *AppService) setDetails(info *models.Formula) {
	if info == nil {
		s.layout.GetDetails().SetContent("")
		return
	}

	// Installation status with colors
	installedStatus := "[red]Not installed[-]"
	installedIcon := "✗"
	if len(info.Installed) > 0 {
		installedStatus = "[green]Installed[-]"
		installedIcon = "✓"

		if info.Outdated {
			installedStatus = "[orange]Update available[-]"
			installedIcon = "⟳"
		}
	}

	// Basic information with icons
	basicInfo := fmt.Sprintf(
		"[yellow::b]%s %s[-]\n\n"+
			"[blue]• Name:[-] %s\n"+
			"[blue]• Version:[-] %s\n"+
			"[blue]• Status:[-] %s %s\n"+
			"[blue]• Tap:[-] %s\n"+
			"[blue]• License:[-] %s\n\n"+
			"[yellow::b]Description[-]\n%s\n\n"+
			"[blue]• Homepage:[-] %s",
		info.Name, installedIcon,
		info.FullName,
		info.Versions.Stable,
		installedStatus, s.getPackageVersionInfo(info),
		info.Tap,
		info.License,
		info.Description,
		info.Homepage,
	)

	// Installation details
	installDetails := s.getPackageInstallationDetails(info)

	// Dependencies with improved formatting
	dependenciesInfo := s.getDependenciesInfo(info)

	s.layout.GetDetails().SetContent(fmt.Sprintf("%s\n\n%s\n\n%s",
		basicInfo, installDetails, dependenciesInfo))
}

func (s *AppService) getPackageVersionInfo(info *models.Formula) string {
	if len(info.Installed) == 0 {
		return ""
	}

	if info.Installed[0].Version < info.Versions.Stable {
		return fmt.Sprintf("([orange]%s[-] → [green]%s[-])",
			info.Installed[0].Version, info.Versions.Stable)
	}

	return fmt.Sprintf("([green]%s[-])", info.Installed[0].Version)
}

func (s *AppService) getPackageInstallationDetails(info *models.Formula) string {
	if len(info.Installed) == 0 {
		return "[yellow::b]Installation[-]\nNot installed"
	}

	packagePrefix, _ := s.BrewService.GetPrefixPath(info.Name)
	installedOnRequest := "No"
	if info.Installed[0].InstalledOnRequest {
		installedOnRequest = "Yes"
	}

	return fmt.Sprintf(
		"[yellow::b]Installation Details[-]\n"+
			"[blue]• Path:[-] %s\n"+
			"[blue]• Installed on request:[-] %s\n"+
			"[blue]• Installed version:[-] %s",
		packagePrefix,
		installedOnRequest,
		info.Installed[0].Version,
	)
}

func (s *AppService) getDependenciesInfo(info *models.Formula) string {
	title := "[yellow::b]Dependencies[-]\n"

	if len(info.Dependencies) == 0 {
		return title + "No dependencies"
	}

	// Format dependencies in multiple columns or with separators
	deps := ""
	for i, dep := range info.Dependencies {
		deps += dep
		if i < len(info.Dependencies)-1 {
			if (i+1)%3 == 0 {
				deps += "\n"
			} else {
				deps += ", "
			}
		}
	}

	return title + deps
}

func (s *AppService) forceRefreshResults() {
	s.app.QueueUpdateDraw(func() {
		_ = s.BrewService.LoadAllFormulae()
		s.search(s.layout.GetSearch().Field().GetText(), false)
	})
}

func (s *AppService) setResults(data *[]models.Formula, scrollToTop bool) {
	s.layout.GetTable().Clear()
	s.layout.GetTable().SetTableHeaders("Name", "Version", "Description")

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

		s.layout.GetTable().View().SetCell(i+1, 0, nameCell.SetExpansion(0))
		s.layout.GetTable().View().SetCell(i+1, 1, versionCell.SetExpansion(0))
		s.layout.GetTable().View().SetCell(i+1, 2, tview.NewTableCell(info.Description).SetSelectable(true).SetExpansion(1))
	}

	// Update the details view with the first item in the list
	if len(*data) > 0 {
		if scrollToTop {
			s.layout.GetTable().View().Select(1, 0)
			s.layout.GetTable().View().ScrollToBeginning()
			s.setDetails(&(*data)[0])
		}

		// Update the filter counter
		s.layout.GetSearch().UpdateCounter(len(*s.packages), len(*s.filteredPackages))
		return
	}

	s.setDetails(nil)
}

func (s *AppService) BuildApp() {
	// Build the layout
	s.layout.Setup()
	s.layout.GetHeader().Update(AppName, AppVersion, s.brewVersion)

	// Evaluate if there is a new version available
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		if latestVersion, err := s.SelfUpdateService.CheckForUpdates(ctx); err == nil && latestVersion != AppVersion {
			s.app.QueueUpdateDraw(func() {
				AppVersion = fmt.Sprintf("%s ([orange]New Version Available: %s[-])", AppVersion, latestVersion)
				s.layout.GetHeader().Update(AppName, AppVersion, s.brewVersion)
			})
		}
	}()

	// Result table section
	tableSelectionChangedFunc := func(row, _ int) {
		if row > 0 && row-1 < len(*s.filteredPackages) {
			s.setDetails(&(*s.filteredPackages)[row-1])
		}
	}
	s.layout.GetTable().View().SetSelectionChangedFunc(tableSelectionChangedFunc)

	// Search field section
	inputDoneFunc := func(key tcell.Key) {
		if key == tcell.KeyEnter || key == tcell.KeyEscape {
			s.app.SetFocus(s.layout.GetTable().View())
		}
	}
	changedFunc := func(text string) {
		s.search(text, true)
	}
	s.layout.GetSearch().SetHandlers(inputDoneFunc, changedFunc)

	// Add key event handler and set the root view
	s.app.SetInputCapture(s.handleKeyEventInput)
	s.app.SetRoot(s.layout.Root(), true)
	s.app.SetFocus(s.layout.GetTable().View())

	go s.updateHomeBrew()          // Update Async the Homebrew formulae
	s.setResults(s.packages, true) // Set the results
}
