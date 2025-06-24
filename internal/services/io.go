package services

import (
	"bbrew/internal/ui"
	"fmt"
	"github.com/gdamore/tcell/v2"
)

type IOKey struct {
	Key     tcell.Key
	Rune    rune
	Name    string
	KeySlug string
	Action  func()
}

var (
	IO_SEARCH           = IOKey{Key: tcell.KeyRune, Rune: '/', KeySlug: "/", Name: "Search"}
	IO_FILTER_INSTALLED = IOKey{Key: tcell.KeyRune, Rune: 'f', KeySlug: "f", Name: "Filter Installed"}
	IO_FILTER_OUTDATED  = IOKey{Key: tcell.KeyRune, Rune: 'o', KeySlug: "o", Name: "Filter Outdated"}
	IO_INSTALL          = IOKey{Key: tcell.KeyRune, Rune: 'i', KeySlug: "i", Name: "Install"}
	IO_UPDATE           = IOKey{Key: tcell.KeyRune, Rune: 'u', KeySlug: "u", Name: "Update"}
	IO_REMOVE           = IOKey{Key: tcell.KeyRune, Rune: 'r', KeySlug: "r", Name: "Remove"}
	IO_UPDATE_ALL       = IOKey{Key: tcell.KeyCtrlU, Rune: 0, KeySlug: "ctrl+u", Name: "Update All"}
	IO_BACK             = IOKey{Key: tcell.KeyEsc, Rune: 0, KeySlug: "esc", Name: "Back to Table"}
	IO_QUIT             = IOKey{Key: tcell.KeyRune, Rune: 'q', KeySlug: "q", Name: "Quit"}
)

type IOServiceInterface interface {
	HandleKeyEventInput(event *tcell.EventKey) *tcell.EventKey
}

type IOService struct {
	appService     *AppService
	layout         ui.LayoutInterface
	commandService CommandServiceInterface
	IOMap          []*IOKey
}

var NewIOService = func(appService *AppService) IOServiceInterface {
	s := &IOService{
		appService:     appService,
		layout:         appService.GetLayout(),
		commandService: appService.CommandService,
	}

	// Define actions for each key input
	s.IOMap = []*IOKey{&IO_SEARCH, &IO_FILTER_INSTALLED, &IO_FILTER_OUTDATED, &IO_INSTALL, &IO_UPDATE, &IO_UPDATE_ALL, &IO_REMOVE, &IO_BACK, &IO_QUIT}
	IO_QUIT.Action = s.handleQuitEvent
	IO_UPDATE.Action = s.handleUpdatePackageEvent
	IO_UPDATE_ALL.Action = s.handleUpdateAllPackagesEvent
	IO_REMOVE.Action = s.handleRemovePackageEvent
	IO_INSTALL.Action = s.handleInstallPackageEvent
	IO_SEARCH.Action = s.handleSearchFieldEvent
	IO_FILTER_INSTALLED.Action = s.handleFilterPackagesEvent
	IO_FILTER_OUTDATED.Action = s.handleFilterOutdatedPackagesEvent
	IO_BACK.Action = s.handleBack

	// Initialize the legend text
	s.layout.GetLegend().SetText(s.getLegendText(""))

	return s
}

func (s *IOService) getLegendText(activeSection string) (legendText string) {
	for i, legend := range s.IOMap {
		if legend.KeySlug == activeSection {
			legendText += s.layout.GetLegend().GetFormattedLabel(legend.KeySlug, legend.Name, true)
		} else {
			legendText += s.layout.GetLegend().GetFormattedLabel(legend.KeySlug, legend.Name, false)
		}

		if i < len(s.IOMap)-1 {
			legendText += " | "
		}
	}

	return legendText
}

func (s *IOService) HandleKeyEventInput(event *tcell.EventKey) *tcell.EventKey {
	if s.layout.GetSearch().Field().HasFocus() {
		return event
	}

	for _, input := range s.IOMap {
		if event.Modifiers() == tcell.ModNone && input.Key == event.Key() && input.Rune == event.Rune() { // Check Rune
			if input.Action != nil {
				input.Action()
				return nil
			}
		} else if event.Modifiers() != tcell.ModNone && input.Key == event.Key() { // Check Key only
			if input.Action != nil {
				input.Action()
				return nil
			}
		}
	}

	return event
}

func (s *IOService) handleBack() {
	s.layout.GetLegend().SetText(s.getLegendText(""))
	s.appService.GetApp().SetRoot(s.layout.Root(), true)
	s.appService.GetApp().SetFocus(s.layout.GetTable().View())
}

func (s *IOService) handleSearchFieldEvent() {
	s.layout.GetLegend().SetText(s.getLegendText(IO_SEARCH.KeySlug))
	s.appService.GetApp().SetFocus(s.layout.GetSearch().Field())
}

func (s *IOService) handleQuitEvent() {
	s.appService.GetApp().Stop()
}

func (s *IOService) handleFilterPackagesEvent() {
	s.layout.GetLegend().SetText(s.getLegendText(""))

	if s.appService.showOnlyOutdated {
		s.appService.showOnlyOutdated = false
		s.appService.showOnlyInstalled = true
	} else {
		s.appService.showOnlyInstalled = !s.appService.showOnlyInstalled
	}

	// Update the search field label
	if s.appService.showOnlyOutdated {
		s.layout.GetSearch().Field().SetLabel("Search (Outdated): ")
		s.layout.GetLegend().SetText(s.getLegendText(IO_FILTER_OUTDATED.KeySlug))
	} else if s.appService.showOnlyInstalled {
		s.layout.GetSearch().Field().SetLabel("Search (Installed): ")
		s.layout.GetLegend().SetText(s.getLegendText(IO_FILTER_INSTALLED.KeySlug))
	} else {
		s.layout.GetSearch().Field().SetLabel("Search (All): ")
	}

	s.appService.search(s.layout.GetSearch().Field().GetText(), true)
}

func (s *IOService) handleFilterOutdatedPackagesEvent() {
	s.layout.GetLegend().SetText(s.getLegendText(""))

	if s.appService.showOnlyInstalled {
		s.appService.showOnlyInstalled = false
		s.appService.showOnlyOutdated = true
	} else {
		s.appService.showOnlyOutdated = !s.appService.showOnlyOutdated
	}

	// Update the search field label
	if s.appService.showOnlyOutdated {
		s.layout.GetSearch().Field().SetLabel("Search (Outdated): ")
		s.layout.GetLegend().SetText(s.getLegendText(IO_FILTER_OUTDATED.KeySlug))
	} else if s.appService.showOnlyInstalled {
		s.layout.GetSearch().Field().SetLabel("Search (Installed): ")
		s.layout.GetLegend().SetText(s.getLegendText(IO_FILTER_INSTALLED.KeySlug))
	} else {
		s.layout.GetSearch().Field().SetLabel("Search (All): ")
	}

	s.appService.search(s.layout.GetSearch().Field().GetText(), true)
}

func (s *IOService) showModal(text string, confirmFunc func(), cancelFunc func()) {
	modal := s.layout.GetModal().Build(text, confirmFunc, cancelFunc)
	s.appService.app.SetRoot(modal, true)
}

func (s *IOService) closeModal() {
	s.appService.app.SetRoot(s.layout.Root(), true)
	s.appService.app.SetFocus(s.layout.GetTable().View())
}

func (s *IOService) handleInstallPackageEvent() {
	row, _ := s.layout.GetTable().View().GetSelection()
	if row > 0 {
		info := (*s.appService.filteredPackages)[row-1]
		s.showModal(
			fmt.Sprintf("Are you sure you want to install the package: %s?", info.Name),
			func() {
				s.closeModal()
				s.layout.GetOutput().Clear()
				go func() {
					s.layout.GetNotifier().ShowWarning(fmt.Sprintf("Installing %s...", info.Name))
					if err := s.appService.CommandService.InstallPackage(info, s.appService.app, s.layout.GetOutput().View()); err != nil {
						s.layout.GetNotifier().ShowError(fmt.Sprintf("Failed to install %s", info.Name))
						return
					}
					s.layout.GetNotifier().ShowSuccess(fmt.Sprintf("Installed %s", info.Name))
					s.appService.forceRefreshResults()
				}()
			}, s.closeModal)
	}
}

func (s *IOService) handleRemovePackageEvent() {
	row, _ := s.layout.GetTable().View().GetSelection()
	if row > 0 {
		info := (*s.appService.filteredPackages)[row-1]
		s.showModal(
			fmt.Sprintf("Are you sure you want to remove the package: %s?", info.Name),
			func() {
				s.closeModal()
				s.layout.GetOutput().Clear()
				go func() {
					s.layout.GetNotifier().ShowWarning(fmt.Sprintf("Removing %s...", info.Name))
					if err := s.appService.CommandService.RemovePackage(info, s.appService.app, s.layout.GetOutput().View()); err != nil {
						s.layout.GetNotifier().ShowError(fmt.Sprintf("Failed to remove %s", info.Name))
						return
					}
					s.layout.GetNotifier().ShowSuccess(fmt.Sprintf("Removed %s", info.Name))
					s.appService.forceRefreshResults()
				}()
			}, s.closeModal)
	}
}

func (s *IOService) handleUpdatePackageEvent() {
	row, _ := s.layout.GetTable().View().GetSelection()
	if row > 0 {
		info := (*s.appService.filteredPackages)[row-1]
		s.showModal(
			fmt.Sprintf("Are you sure you want to update the package: %s?", info.Name),
			func() {
				s.closeModal()
				s.layout.GetOutput().Clear()
				go func() {
					s.layout.GetNotifier().ShowWarning(fmt.Sprintf("Updating %s...", info.Name))
					if err := s.appService.CommandService.UpdatePackage(info, s.appService.app, s.layout.GetOutput().View()); err != nil {
						s.layout.GetNotifier().ShowError(fmt.Sprintf("Failed to update %s", info.Name))
						return
					}
					s.layout.GetNotifier().ShowSuccess(fmt.Sprintf("Updated %s", info.Name))
					s.appService.forceRefreshResults()
				}()
			}, s.closeModal)
	}
}

func (s *IOService) handleUpdateAllPackagesEvent() {
	s.showModal("Are you sure you want to update all Packages?", func() {
		s.closeModal()
		s.layout.GetOutput().Clear()
		go func() {
			s.layout.GetNotifier().ShowWarning("Updating all Packages...")
			if err := s.appService.CommandService.UpdateAllPackages(s.appService.app, s.layout.GetOutput().View()); err != nil {
				s.layout.GetNotifier().ShowError("Failed to update all Packages")
				return
			}
			s.layout.GetNotifier().ShowSuccess("Updated all Packages")
			s.appService.forceRefreshResults()
		}()
	}, s.closeModal)
}
