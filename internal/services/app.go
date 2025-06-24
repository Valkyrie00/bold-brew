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
	showOnlyOutdated  bool
	brewVersion       string

	BrewService       BrewServiceInterface
	CommandService    CommandServiceInterface
	SelfUpdateService SelfUpdateServiceInterface
}

func NewAppService() AppServiceInterface {
	app := tview.NewApplication()
	themeService := theme.NewTheme()
	brewService := NewBrewService()

	appService := &AppService{
		app:    app,
		theme:  themeService,
		layout: ui.NewLayout(themeService),

		packages:          new([]models.Formula),
		filteredPackages:  new([]models.Formula),
		showOnlyInstalled: false,
		showOnlyOutdated:  false,
		brewVersion:       "-",

		BrewService:       brewService,
		CommandService:    NewCommandService(),
		SelfUpdateService: NewSelfUpdateService(),
	}

	return appService
}

func (s *AppService) GetApp() *tview.Application {
	return s.app
}

func (s *AppService) Boot() (err error) {
	if s.brewVersion, err = s.BrewService.GetBrewVersion(); err != nil {
		// This error is critical, as we need Homebrew to function
		return fmt.Errorf("failed to get Homebrew version: %v", err)
	}

	// Download and parse Homebrew formulae data
	if err = s.BrewService.SetupData(false); err != nil {
		return fmt.Errorf("failed to load Homebrew formulae: %v", err)
	}

	s.packages = s.BrewService.GetFormulae()
	*s.filteredPackages = *s.packages

	return nil
}

func (s *AppService) updateHomeBrew() {
	s.layout.GetNotifier().ShowWarning("Updating Homebrew formulae...")
	if err := s.BrewService.UpdateHomebrew(); err != nil {
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

func (s *AppService) setDetails(info *models.Formula) {
	if info == nil {
		s.layout.GetDetails().SetContent(nil)
		return
	}

	s.layout.GetDetails().SetContent(info)
}

func (s *AppService) forceRefreshResults() {
	_ = s.BrewService.SetupData(true)
	s.packages = s.BrewService.GetFormulae()
	*s.filteredPackages = *s.packages

	s.app.QueueUpdateDraw(func() {
		s.search(s.layout.GetSearch().Field().GetText(), false)
	})
}

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
	s.layout.GetLegend().SetText(s.GetLegendText("")) // Initialize the legend text
	s.app.SetRoot(s.layout.Root(), true)
	s.app.SetFocus(s.layout.GetTable().View())

	go s.updateHomeBrew()          // Update Async the Homebrew formulae
	s.setResults(s.packages, true) // Set the results
}
