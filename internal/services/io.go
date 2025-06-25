package services

import (
	"bbrew/internal/ui"
	"fmt"
	"github.com/gdamore/tcell/v2"
)

var (
	IoSearch          = IOAction{Key: tcell.KeyRune, Rune: '/', KeySlug: "/", Name: "Search"}
	IoFilterInstalled = IOAction{Key: tcell.KeyRune, Rune: 'f', KeySlug: "f", Name: "Filter Installed"}
	IoFilterOutdated  = IOAction{Key: tcell.KeyRune, Rune: 'o', KeySlug: "o", Name: "Filter Outdated"}
	IoInstall         = IOAction{Key: tcell.KeyRune, Rune: 'i', KeySlug: "i", Name: "Install"}
	IoUpdate          = IOAction{Key: tcell.KeyRune, Rune: 'u', KeySlug: "u", Name: "Update"}
	IoRemove          = IOAction{Key: tcell.KeyRune, Rune: 'r', KeySlug: "r", Name: "Remove"}
	IoUpdateAll       = IOAction{Key: tcell.KeyCtrlU, Rune: 0, KeySlug: "ctrl+u", Name: "Update All"}
	IoBack            = IOAction{Key: tcell.KeyEsc, Rune: 0, KeySlug: "esc", Name: "Back to Table"}
	IoQuit            = IOAction{Key: tcell.KeyRune, Rune: 'q', KeySlug: "q", Name: "Quit"}
)

type IOAction struct {
	Key     tcell.Key
	Rune    rune
	Name    string
	KeySlug string
	Action  func()
}

func (k *IOAction) SetAction(action func()) {
	k.Action = action
}

type IOServiceInterface interface {
	HandleKeyEventInput(event *tcell.EventKey) *tcell.EventKey
}

type IOService struct {
	appService     *AppService
	layout         ui.LayoutInterface
	commandService CommandServiceInterface
	keyActions     []*IOAction
	legendEntries  []struct{ KeySlug, Name string }
}

var NewIOService = func(appService *AppService) IOServiceInterface {
	s := &IOService{
		appService:     appService,
		layout:         appService.GetLayout(),
		commandService: NewCommandService(),
	}

	// Define actions for each key input
	s.keyActions = []*IOAction{&IoSearch, &IoFilterInstalled, &IoFilterOutdated, &IoInstall, &IoUpdate, &IoUpdateAll, &IoRemove, &IoBack, &IoQuit}
	IoQuit.SetAction(s.handleQuitEvent)
	IoUpdate.SetAction(s.handleUpdatePackageEvent)
	IoUpdateAll.SetAction(s.handleUpdateAllPackagesEvent)
	IoRemove.SetAction(s.handleRemovePackageEvent)
	IoInstall.SetAction(s.handleInstallPackageEvent)
	IoSearch.SetAction(s.handleSearchFieldEvent)
	IoFilterInstalled.SetAction(s.handleFilterPackagesEvent)
	IoFilterOutdated.SetAction(s.handleFilterOutdatedPackagesEvent)
	IoBack.SetAction(s.handleBack)

	// Convert IOMap to a map for easier access
	s.legendEntries = make([]struct{ KeySlug, Name string }, len(s.keyActions))
	for i, input := range s.keyActions {
		s.legendEntries[i] = struct{ KeySlug, Name string }{KeySlug: input.KeySlug, Name: input.Name}
	}

	// Initialize the legend text
	s.layout.GetLegend().SetLegend(s.legendEntries, "")
	return s
}

func (s *IOService) HandleKeyEventInput(event *tcell.EventKey) *tcell.EventKey {
	if s.layout.GetSearch().Field().HasFocus() {
		return event
	}

	for _, input := range s.keyActions {
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
	s.appService.GetApp().SetRoot(s.layout.Root(), true)
	s.appService.GetApp().SetFocus(s.layout.GetTable().View())
}

func (s *IOService) handleSearchFieldEvent() {
	s.appService.GetApp().SetFocus(s.layout.GetSearch().Field())
}

func (s *IOService) handleQuitEvent() {
	s.appService.GetApp().Stop()
}

func (s *IOService) handleFilterPackagesEvent() {
	s.layout.GetLegend().SetLegend(s.legendEntries, "")

	if s.appService.showOnlyOutdated {
		s.appService.showOnlyOutdated = false
		s.appService.showOnlyInstalled = true
	} else {
		s.appService.showOnlyInstalled = !s.appService.showOnlyInstalled
	}

	// Update the search field label
	if s.appService.showOnlyOutdated {
		s.layout.GetSearch().Field().SetLabel("Search (Outdated): ")
		s.layout.GetLegend().SetLegend(s.legendEntries, IoFilterOutdated.KeySlug)
	} else if s.appService.showOnlyInstalled {
		s.layout.GetSearch().Field().SetLabel("Search (Installed): ")
		s.layout.GetLegend().SetLegend(s.legendEntries, IoFilterInstalled.KeySlug)
	} else {
		s.layout.GetSearch().Field().SetLabel("Search (All): ")
	}

	s.appService.search(s.layout.GetSearch().Field().GetText(), true)
}

func (s *IOService) handleFilterOutdatedPackagesEvent() {
	s.layout.GetLegend().SetLegend(s.legendEntries, "")

	if s.appService.showOnlyInstalled {
		s.appService.showOnlyInstalled = false
		s.appService.showOnlyOutdated = true
	} else {
		s.appService.showOnlyOutdated = !s.appService.showOnlyOutdated
	}

	// Update the search field label
	if s.appService.showOnlyOutdated {
		s.layout.GetSearch().Field().SetLabel("Search (Outdated): ")
		s.layout.GetLegend().SetLegend(s.legendEntries, IoFilterOutdated.KeySlug)
	} else if s.appService.showOnlyInstalled {
		s.layout.GetSearch().Field().SetLabel("Search (Installed): ")
		s.layout.GetLegend().SetLegend(s.legendEntries, IoFilterInstalled.KeySlug)
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
					if err := s.commandService.InstallPackage(info, s.appService.app, s.layout.GetOutput().View()); err != nil {
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
					if err := s.commandService.RemovePackage(info, s.appService.app, s.layout.GetOutput().View()); err != nil {
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
					if err := s.commandService.UpdatePackage(info, s.appService.app, s.layout.GetOutput().View()); err != nil {
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
			if err := s.commandService.UpdateAllPackages(s.appService.app, s.layout.GetOutput().View()); err != nil {
				s.layout.GetNotifier().ShowError("Failed to update all Packages")
				return
			}
			s.layout.GetNotifier().ShowSuccess("Updated all Packages")
			s.appService.forceRefreshResults()
		}()
	}, s.closeModal)
}
