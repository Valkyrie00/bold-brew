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
	InitData() (err error)
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

func (s *AppService) InitData() (err error) {
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

func (s *AppService) applySearchFilter(
	searchText string,
) {
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
	s.fillTable(s.filteredPackages)

	s.LayoutService.GetFilterCounter().SetText(fmt.Sprintf("Total: %d | Filtered: %d", len(*s.packages), len(*s.filteredPackages)))
}

func (s *AppService) updateDetailsView(info *models.Formula) {
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

func (s *AppService) updateTableView() {
	s.app.QueueUpdateDraw(func() {
		_ = s.BrewService.LoadAllFormulae()
		s.applySearchFilter(s.LayoutService.GetSearchField().GetText())
		s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetTableResult())
	})
}

func (s *AppService) fillTable(data *[]models.Formula) {
	headers := []string{"Name", "Description", "Version"}
	s.LayoutService.GetTableResult().Clear()

	for i, header := range headers {
		s.LayoutService.GetTableResult().SetCell(0, i, tview.NewTableCell(header).
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

		s.LayoutService.GetTableResult().SetCell(i+1, 0, nameCell)
		s.LayoutService.GetTableResult().SetCell(i+1, 1, tview.NewTableCell(info.Description).SetSelectable(true))
		s.LayoutService.GetTableResult().SetCell(i+1, 2, versionCell)
	}

	// Update the details view with the first item in the list
	if len(*data) > 0 {
		s.LayoutService.GetTableResult().Select(1, 0)
		s.LayoutService.GetTableResult().ScrollToBeginning()
		s.updateDetailsView(&(*data)[0])
		return
	}

	s.updateDetailsView(nil)
}

func (s *AppService) BuildApp() {
	latestVersion, err := s.SelfUpdateService.CheckForUpdates()
	if err == nil && latestVersion != AppVersion {
		AppVersion = fmt.Sprintf("%s ([orange]Update available: %s[-])", AppVersion, latestVersion)
	}

	s.LayoutService.SetHeader(AppName, AppVersion, s.brewVersion)
	s.LayoutService.SetLegend()

	tableSelectionChangedFunc := func(row, column int) {
		if row > 0 && row-1 < len(*s.filteredPackages) {
			s.updateDetailsView(&(*s.filteredPackages)[row-1])
		}
	}

	s.LayoutService.SetTableResult(tableSelectionChangedFunc)
	s.LayoutService.SetDetailsView()
	s.LayoutService.SetBuildOutputView()

	// Search input to filter packages
	inputDoneFunc := func(key tcell.Key) {
		if key == tcell.KeyEnter || key == tcell.KeyEscape {
			s.app.SetFocus(s.LayoutService.GetTableResult())
		}
	}

	changedFunc := func(text string) {
		s.applySearchFilter(s.LayoutService.GetSearchField().GetText())
	}

	s.LayoutService.SetSearchField(inputDoneFunc, changedFunc)

	s.LayoutService.SetFilterCounter(len(*s.packages), len(*s.filteredPackages))

	s.LayoutService.SetGrid()

	// Add key event handler
	s.app.SetInputCapture(s.handleKeyEventInput)

	// Set the grid as the root of the application
	s.app.SetRoot(s.LayoutService.GetGrid(), true)
	s.app.SetFocus(s.LayoutService.GetTableResult())

	// Fill the table with the initial data
	s.fillTable(s.packages)
}

func (s *AppService) handleKeyEventInput(event *tcell.EventKey) *tcell.EventKey {
	if s.LayoutService.GetSearchField().HasFocus() {
		return event
	}

	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'q': // Quit the application
			s.app.Stop()
			return nil
		case 'u': // Update the selected package
			row, _ := s.LayoutService.GetTableResult().GetSelection()
			if row > 0 {
				info := (*s.filteredPackages)[row-1]
				modal := s.LayoutService.GenerateModal(fmt.Sprintf("Are you sure you want to update the package: %s?", info.Name), func() {
					s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetTableResult())
					s.LayoutService.GetOutputView().Clear()
					go func() {
						_ = s.CommandService.UpdatePackage(info, s.app, s.LayoutService.GetOutputView())
						s.updateTableView()
					}()
				}, func() {
					s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetTableResult())
				})

				s.app.SetRoot(modal, true).SetFocus(modal)
			}
			return nil
		case 'r': // Remove the selected package
			row, _ := s.LayoutService.GetTableResult().GetSelection()
			if row > 0 {
				info := (*s.filteredPackages)[row-1]
				modal := s.LayoutService.GenerateModal(fmt.Sprintf("Are you sure you want to remove the package: %s?", info.Name), func() {
					s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetTableResult())
					s.LayoutService.GetOutputView().Clear()
					go func() {
						_ = s.CommandService.RemovePackage(info, s.app, s.LayoutService.GetOutputView())
						s.updateTableView()
					}()
				}, func() {
					s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetTableResult())
				})
				s.app.SetRoot(modal, true).SetFocus(modal)
			}
			return nil
		case 'i': // Install the selected package
			row, _ := s.LayoutService.GetTableResult().GetSelection()
			if row > 0 {
				info := (*s.filteredPackages)[row-1]
				modal := s.LayoutService.GenerateModal(fmt.Sprintf("Are you sure you want to install the package: %s?", info.Name), func() {
					s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetTableResult())
					s.LayoutService.GetOutputView().Clear()
					go func() {
						_ = s.CommandService.InstallPackage(info, s.app, s.LayoutService.GetOutputView())
						s.updateTableView()
					}()
				}, func() {
					s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetTableResult())
				})
				s.app.SetRoot(modal, true).SetFocus(modal)
			}
			return nil
		case '/':
			s.app.SetFocus(s.LayoutService.GetSearchField())
			return nil
		case 'f':
			s.showOnlyInstalled = !s.showOnlyInstalled
			if s.showOnlyInstalled {
				s.LayoutService.GetSearchField().SetLabel("Search (Installed): ")
			} else {
				s.LayoutService.GetSearchField().SetLabel("Search (All): ")
			}
			s.applySearchFilter(s.LayoutService.GetSearchField().GetText())
			s.LayoutService.GetTableResult().ScrollToBeginning()
			return nil
		}
	case tcell.KeyCtrlU:
		// Update homebrew
		modal := s.LayoutService.GenerateModal("Are you sure you want to update Homebrew?", func() {
			s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetTableResult())
			s.LayoutService.GetOutputView().Clear()
			go func() {
				_ = s.CommandService.UpdateHomebrew(s.app, s.LayoutService.GetOutputView())
				s.updateTableView()
			}()
		}, func() {
			s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetTableResult())
		})
		s.app.SetRoot(modal, true).SetFocus(modal)
		return nil
	case tcell.KeyEsc:
		s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetTableResult())
		return nil
	}

	return event
}
