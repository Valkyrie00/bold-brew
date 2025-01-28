package services

import (
	"bbrew/internal/models"
	"fmt"
	"github.com/gdamore/tcell/v2"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rivo/tview"
	"os"
	"strings"
)

var (
	AppName    = os.Getenv("APP_NAME")
	AppVersion = os.Getenv("APP_VERSION")
)

type AppServiceInterface interface {
	GetApp() *tview.Application
	InitData() (err error)
	BuildApp()
}

type AppService struct {
	// Components IO
	app              *tview.Application
	table            *tview.Table
	detailsView      *tview.TextView
	outputView       *tview.TextView
	searchInput      *tview.InputField
	packageCountView *tview.TextView
	currentModal     *tview.Modal
	grid             *tview.Grid

	// Data
	packages          *[]models.Formula
	filteredPackages  *[]models.Formula
	showOnlyInstalled bool
	brewVersion       string

	// Services
	BrewService    BrewServiceInterface
	CommandService CommandServiceInterface
}

var NewAppService = func() AppServiceInterface {
	return &AppService{
		app:               tview.NewApplication(), // Initialize the application
		packages:          new([]models.Formula),
		filteredPackages:  new([]models.Formula),
		showOnlyInstalled: false, // Default to show all packages
		brewVersion:       "-",

		// Services
		BrewService:    NewBrewService(),
		CommandService: NewCommandService(),
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

	s.packageCountView.SetText(fmt.Sprintf("Total: %d | Filtered: %d", len(*s.packages), len(*s.filteredPackages)))
}

func (s *AppService) updateDetailsView(
	info *models.Formula,
) {
	if info != nil {
		installedVersion := "Not installed"
		packagePrefix := "-"
		if len(info.Installed) > 0 {
			installedVersion = info.Installed[0].Version
			packagePrefix, _ = s.BrewService.GetPrefixPath(info.Name)
		}

		s.detailsView.SetText(
			fmt.Sprintf("Name: %s\nDescription: %s\nInstalled: %s\nAvailable Version: %s\nInstall Path: %s",
				info.Name, info.Description, installedVersion, info.Versions.Stable, packagePrefix),
		)
		return
	}

	s.detailsView.SetText("")
}

func (s *AppService) createModal(text string, confirmFunc func()) *tview.Modal {
	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"Confirm", "Cancel"}).
		SetBackgroundColor(tcell.ColorDarkSlateGray).
		SetTextColor(tcell.ColorWhite).
		SetButtonBackgroundColor(tcell.ColorGray).
		SetButtonTextColor(tcell.ColorWhite).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			s.app.SetRoot(s.grid, true).SetFocus(s.table)
			if buttonLabel == "Confirm" {
				confirmFunc()
			}
		})

	s.currentModal = modal
	return modal
}

func (s *AppService) updateTableView() {
	s.app.QueueUpdateDraw(func() {
		_ = s.BrewService.LoadAllFormulae()
		s.applySearchFilter(s.searchInput.GetText())
		s.app.SetRoot(s.grid, true).SetFocus(s.table)
	})
}

func (s *AppService) fillTable(data *[]models.Formula) {
	headers := []string{"Name", "Description", "Installed", "Available Version"}
	s.table.Clear()

	for i, header := range headers {
		s.table.SetCell(0, i, tview.NewTableCell(header).
			SetTextColor(tcell.ColorBlue).
			SetAlign(tview.AlignLeft).
			SetSelectable(false).
			SetExpansion(1))
	}

	for i, info := range *data {
		installedVersion := ""
		if len(info.Installed) > 0 {
			installedVersion = info.Installed[0].Version
		}

		nameCell := tview.NewTableCell(info.Name).SetSelectable(true)
		if installedVersion != "" {
			nameCell.SetTextColor(tcell.ColorGreen)
		}

		installedVersionCell := tview.NewTableCell(installedVersion).SetSelectable(true)
		if installedVersion != "" && installedVersion < info.Versions.Stable {
			installedVersionCell.SetTextColor(tcell.ColorOrange)
		}

		s.table.SetCell(i+1, 0, nameCell)
		s.table.SetCell(i+1, 1, tview.NewTableCell(info.Description).SetSelectable(true))
		s.table.SetCell(i+1, 2, installedVersionCell)
		s.table.SetCell(i+1, 3, tview.NewTableCell(info.Versions.Stable).SetSelectable(true))
	}

	// Update the details view with the first item in the list
	if len(*data) > 0 {
		s.table.Select(1, 0)
		s.table.ScrollToBeginning()
		s.updateDetailsView(&(*data)[0])
		return
	}

	s.updateDetailsView(nil)
}

func (s *AppService) BuildApp() {
	header := tview.NewTextView().
		SetText(fmt.Sprintf("%s %s - %s", AppName, AppVersion, s.brewVersion)).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	legend := tview.NewTextView().
		SetText(tview.Escape("[Up/Down] Navigate | [/] Search | [f] Filter Installed Only | [i] Install | [u] Update | [r] Remove | [Esc] Back to Table | [q] Quit")).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	s.table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

	s.table.SetSelectionChangedFunc(func(row, column int) {
		if row > 0 && row-1 < len(*s.filteredPackages) {
			s.updateDetailsView(&(*s.filteredPackages)[row-1])
		}
	})

	// Details view to show package information
	s.detailsView = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft)
	s.detailsView.SetTitle("Details").SetTitleAlign(tview.AlignLeft).SetBorder(true)

	// Output view to show command output
	s.outputView = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft)
	s.outputView.SetBorder(true).SetTitle("Output").SetTitleAlign(tview.AlignLeft)

	// Search input to filter packages
	s.searchInput = tview.NewInputField().
		SetLabel("Search (All): ").
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorYellow).
		SetFieldWidth(30).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				s.app.SetFocus(s.table)
			}
		})

	s.searchInput.SetChangedFunc(func(text string) {
		s.applySearchFilter(s.searchInput.GetText())
	})

	s.packageCountView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight).
		SetText(fmt.Sprintf("Total: %d | Filtered: %d", len(*s.packages), len(*s.filteredPackages)))

	// Create a grid layout to hold the header, table, search input, and the legend
	searchRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(s.searchInput, 0, 1, false).
		AddItem(s.packageCountView, 0, 1, false)

	filtersArea := tview.NewFrame(searchRow).
		SetBorders(0, 0, 0, 0, 3, 3)

	tableFrame := tview.NewFrame(s.table).
		SetBorders(0, 0, 0, 0, 3, 3)

	leftColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(filtersArea, 2, 0, false). // Fixed height of 3 rows
		AddItem(tableFrame, 0, 4, false)

	rightColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(s.detailsView, 0, 2, false).
		AddItem(s.outputView, 0, 1, false)

	mainContent := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftColumn, 0, 1, false).
		AddItem(rightColumn, 0, 1, false)

	s.grid = tview.NewGrid().
		SetRows(1, 0, 1).
		SetColumns(0).
		SetBorders(true).
		AddItem(header, 0, 0, 1, 1, 0, 0, false).
		AddItem(mainContent, 1, 0, 1, 1, 0, 0, true).
		AddItem(legend, 2, 0, 1, 1, 0, 0, false)

	// Add key event handler
	s.app.SetInputCapture(s.handleKeyEventInput)

	// Set the grid as the root of the application
	s.app.SetRoot(s.grid, true)
	s.app.SetFocus(s.table)

	// Fill the table with the initial data
	s.fillTable(s.packages)
}

func (s *AppService) handleKeyEventInput(event *tcell.EventKey) *tcell.EventKey {
	if s.searchInput.HasFocus() {
		return event
	}

	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'q': // Quit the application
			s.app.Stop()
			return nil
		case 'u': // Update the selected package
			row, _ := s.table.GetSelection()
			if row > 0 {
				info := (*s.filteredPackages)[row-1]
				modal := s.createModal(fmt.Sprintf("Are you sure you want to update the package: %s?", info.Name), func() {
					s.outputView.Clear()
					go func() {
						err := s.CommandService.UpdatePackage(info, s.app, s.outputView)
						if err != nil {
							s.app.QueueUpdateDraw(func() {
								errorModal := s.createModal(fmt.Sprintf("Failed to update package: %s\nError: %v", info.Name, err), nil)
								s.app.SetRoot(errorModal, true).SetFocus(errorModal)
							})
						} else {
							s.updateTableView()
						}
					}()
				})
				s.app.SetRoot(modal, true).SetFocus(modal)
			}
			return nil
		case 'r': // Remove the selected package
			row, _ := s.table.GetSelection()
			if row > 0 {
				info := (*s.filteredPackages)[row-1]
				modal := s.createModal(fmt.Sprintf("Are you sure you want to remove the package: %s?", info.Name), func() {
					s.outputView.Clear()
					go func() {
						err := s.CommandService.RemovePackage(info, s.app, s.outputView)
						if err != nil {
							s.app.QueueUpdateDraw(func() {
								errorModal := s.createModal(fmt.Sprintf("Failed to remove package: %s\nError: %v", info.Name, err), nil)
								s.app.SetRoot(errorModal, true).SetFocus(errorModal)
							})
						} else {
							s.updateTableView()
						}
					}()
				})
				s.app.SetRoot(modal, true).SetFocus(modal)
			}
			return nil
		case 'i': // Install the selected package
			row, _ := s.table.GetSelection()
			if row > 0 {
				info := (*s.filteredPackages)[row-1]
				modal := s.createModal(fmt.Sprintf("Are you sure you want to install the package: %s?", info.Name), func() {
					s.outputView.Clear()
					go func() {
						err := s.CommandService.InstallPackage(info, s.app, s.outputView)
						if err != nil {
							s.app.QueueUpdateDraw(func() {
								errorModal := s.createModal(fmt.Sprintf("Failed to install package: %s\nError: %v", info.Name, err), nil)
								s.app.SetRoot(errorModal, true).SetFocus(errorModal)
							})
						} else {
							s.updateTableView()
						}
					}()
				})
				s.app.SetRoot(modal, true).SetFocus(modal)
			}
			return nil
		case '/':
			s.app.SetFocus(s.searchInput)
			return nil
		case 'f':
			s.showOnlyInstalled = !s.showOnlyInstalled
			if s.showOnlyInstalled {
				s.searchInput.SetLabel("Search (Installed): ")
			} else {
				s.searchInput.SetLabel("Search (All): ")
			}
			s.applySearchFilter(s.searchInput.GetText())
			s.table.ScrollToBeginning()
			return nil
		}
	case tcell.KeyEsc:
		// Remove the modal if it is currently displayed
		if s.currentModal != nil {
			s.currentModal = nil
		}

		s.app.SetRoot(s.grid, true).SetFocus(s.table)
		return nil
	}
	return event
}
