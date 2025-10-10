package services

import (
	"bbrew/internal/ui"
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type FilterType int

const (
	FilterInstalled FilterType = iota
	FilterOutdated
	FilterLeaves
)

// IOAction represents an input/output action that can be triggered by a key event.
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

// IOServiceInterface defines the interface for handling input/output actions in the application.
type IOServiceInterface interface {
	HandleKeyEventInput(event *tcell.EventKey) *tcell.EventKey
}

// IOService implements the IOServiceInterface and handles key events for the application.
type IOService struct {
	appService    *AppService
	layout        ui.LayoutInterface
	brewService   BrewServiceInterface
	keyActions    []*IOAction
	legendEntries []struct{ KeySlug, Name string }

	// Actions for each key input
	ActionSearch          *IOAction
	ActionFilterInstalled *IOAction
	ActionFilterOutdated  *IOAction
	ActionFilterLeaves    *IOAction
	ActionInstall         *IOAction
	ActionUpdate          *IOAction
	ActionRemove          *IOAction
	ActionUpdateAll       *IOAction
	ActionBack            *IOAction
	ActionQuit            *IOAction
}

var NewIOService = func(appService *AppService) IOServiceInterface {
	s := &IOService{
		appService:  appService,
		layout:      appService.GetLayout(),
		brewService: NewBrewService(),
	}

	// Initialize key actions with their respective keys, runes, and names.
	s.ActionSearch = &IOAction{Key: tcell.KeyRune, Rune: '/', KeySlug: "/", Name: "Search"}
	s.ActionFilterInstalled = &IOAction{Key: tcell.KeyRune, Rune: 'f', KeySlug: "f", Name: "Filter Installed"}
	s.ActionFilterOutdated = &IOAction{Key: tcell.KeyRune, Rune: 'o', KeySlug: "o", Name: "Filter Outdated"}
	s.ActionFilterLeaves = &IOAction{Key: tcell.KeyRune, Rune: 'l', KeySlug: "l", Name: "Filter Leaves"}
	s.ActionInstall = &IOAction{Key: tcell.KeyRune, Rune: 'i', KeySlug: "i", Name: "Install"}
	s.ActionUpdate = &IOAction{Key: tcell.KeyRune, Rune: 'u', KeySlug: "u", Name: "Update"}
	s.ActionRemove = &IOAction{Key: tcell.KeyRune, Rune: 'r', KeySlug: "r", Name: "Remove"}
	s.ActionUpdateAll = &IOAction{Key: tcell.KeyCtrlU, Rune: 0, KeySlug: "ctrl+u", Name: "Update All"}
	s.ActionBack = &IOAction{Key: tcell.KeyEsc, Rune: 0, KeySlug: "esc", Name: "Back to Table"}
	s.ActionQuit = &IOAction{Key: tcell.KeyRune, Rune: 'q', KeySlug: "q", Name: "Quit"}

	// Define actions for each key input,
	s.ActionSearch.SetAction(s.handleSearchFieldEvent)
	s.ActionFilterInstalled.SetAction(s.handleFilterPackagesEvent)
	s.ActionFilterOutdated.SetAction(s.handleFilterOutdatedPackagesEvent)
	s.ActionFilterLeaves.SetAction(s.handleFilterLeavesEvent)
	s.ActionInstall.SetAction(s.handleInstallPackageEvent)
	s.ActionUpdate.SetAction(s.handleUpdatePackageEvent)
	s.ActionRemove.SetAction(s.handleRemovePackageEvent)
	s.ActionUpdateAll.SetAction(s.handleUpdateAllPackagesEvent)
	s.ActionBack.SetAction(s.handleBack)
	s.ActionQuit.SetAction(s.handleQuitEvent)

	// Add all actions to the keyActions slice
	s.keyActions = []*IOAction{
		s.ActionSearch,
		s.ActionFilterInstalled,
		s.ActionFilterOutdated,
		s.ActionFilterLeaves,
		s.ActionInstall,
		s.ActionUpdate,
		s.ActionRemove,
		s.ActionUpdateAll,
		s.ActionBack,
		s.ActionQuit,
	}

	// Convert keyActions to legend entries
	s.legendEntries = make([]struct{ KeySlug, Name string }, len(s.keyActions))
	for i, input := range s.keyActions {
		s.legendEntries[i] = struct{ KeySlug, Name string }{KeySlug: input.KeySlug, Name: input.Name}
	}

	// Initialize the legend text, literally the UI component that displays the key bindings
	s.layout.GetLegend().SetLegend(s.legendEntries, "")
	return s
}

// HandleKeyEventInput processes key events and triggers the corresponding actions.
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

// handleBack is called when the user presses the back key (Esc).
func (s *IOService) handleBack() {
	s.appService.GetApp().SetRoot(s.layout.Root(), true)
	s.appService.GetApp().SetFocus(s.layout.GetTable().View())
}

// handleSearchFieldEvent is called when the user presses the search key (/).
func (s *IOService) handleSearchFieldEvent() {
	s.appService.GetApp().SetFocus(s.layout.GetSearch().Field())
}

// handleQuitEvent is called when the user presses the quit key (q).
func (s *IOService) handleQuitEvent() {
	s.appService.GetApp().Stop()
}

// handleFilterEvent toggles the filter for installed or outdated packages based on the provided filter type.
func (s *IOService) handleFilterEvent(filterType FilterType) {
	s.layout.GetLegend().SetLegend(s.legendEntries, "")

	switch filterType {
	case FilterInstalled:
		if s.appService.showOnlyOutdated || s.appService.showOnlyLeaves {
			s.appService.showOnlyOutdated = false
			s.appService.showOnlyLeaves = false
			s.appService.showOnlyInstalled = true
		} else {
			s.appService.showOnlyInstalled = !s.appService.showOnlyInstalled
		}
	case FilterOutdated:
		if s.appService.showOnlyInstalled || s.appService.showOnlyLeaves {
			s.appService.showOnlyInstalled = false
			s.appService.showOnlyLeaves = false
			s.appService.showOnlyOutdated = true
		} else {
			s.appService.showOnlyOutdated = !s.appService.showOnlyOutdated
		}
	case FilterLeaves:
		if s.appService.showOnlyInstalled || s.appService.showOnlyOutdated {
			s.appService.showOnlyInstalled = false
			s.appService.showOnlyOutdated = false
			s.appService.showOnlyLeaves = true
		} else {
			s.appService.showOnlyLeaves = !s.appService.showOnlyLeaves
		}
	}

	// Update the search field label and legend based on the current filter state
	if s.appService.showOnlyOutdated {
		s.layout.GetSearch().Field().SetLabel("Search (Outdated): ")
		s.layout.GetLegend().SetLegend(s.legendEntries, s.ActionFilterOutdated.KeySlug)
	} else if s.appService.showOnlyInstalled {
		s.layout.GetSearch().Field().SetLabel("Search (Installed): ")
		s.layout.GetLegend().SetLegend(s.legendEntries, s.ActionFilterInstalled.KeySlug)
	} else if s.appService.showOnlyLeaves {
		s.layout.GetSearch().Field().SetLabel("Search (Leaves): ")
		s.layout.GetLegend().SetLegend(s.legendEntries, s.ActionFilterLeaves.KeySlug)
	} else {
		s.layout.GetSearch().Field().SetLabel("Search (All): ")
	}

	s.appService.search(s.layout.GetSearch().Field().GetText(), true)
}

// handleFilterPackagesEvent toggles the filter for installed packages
func (s *IOService) handleFilterPackagesEvent() {
	s.handleFilterEvent(FilterInstalled)
}

// handleFilterOutdatedPackagesEvent toggles the filter for outdated packages
func (s *IOService) handleFilterOutdatedPackagesEvent() {
	s.handleFilterEvent(FilterOutdated)
}

// handleFilterLeavesEvent toggles the filter for leaf packages (installed on request)
func (s *IOService) handleFilterLeavesEvent() {
	s.handleFilterEvent(FilterLeaves)
}

// showModal displays a modal dialog with the specified text and confirmation/cancellation actions.
// This is used for actions like installing, removing, or updating packages, invoking user confirmation.
func (s *IOService) showModal(text string, confirmFunc func(), cancelFunc func()) {
	modal := s.layout.GetModal().Build(text, confirmFunc, cancelFunc)
	s.appService.app.SetRoot(modal, true)
}

// closeModal closes the currently displayed modal dialog and returns focus to the main table view.
func (s *IOService) closeModal() {
	s.appService.app.SetRoot(s.layout.Root(), true)
	s.appService.app.SetFocus(s.layout.GetTable().View())
}

// handleInstallPackageEvent is called when the user presses the installation key (i).
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
					if err := s.brewService.InstallPackage(info, s.appService.app, s.layout.GetOutput().View()); err != nil {
						s.layout.GetNotifier().ShowError(fmt.Sprintf("Failed to install %s", info.Name))
						return
					}
					s.layout.GetNotifier().ShowSuccess(fmt.Sprintf("Installed %s", info.Name))
					s.appService.forceRefreshResults()
				}()
			}, s.closeModal)
	}
}

// handleRemovePackageEvent is called when the user presses the removal key (r).
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
					if err := s.brewService.RemovePackage(info, s.appService.app, s.layout.GetOutput().View()); err != nil {
						s.layout.GetNotifier().ShowError(fmt.Sprintf("Failed to remove %s", info.Name))
						return
					}
					s.layout.GetNotifier().ShowSuccess(fmt.Sprintf("Removed %s", info.Name))
					s.appService.forceRefreshResults()
				}()
			}, s.closeModal)
	}
}

// handleUpdatePackageEvent is called when the user presses the update key (u).
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
					if err := s.brewService.UpdatePackage(info, s.appService.app, s.layout.GetOutput().View()); err != nil {
						s.layout.GetNotifier().ShowError(fmt.Sprintf("Failed to update %s", info.Name))
						return
					}
					s.layout.GetNotifier().ShowSuccess(fmt.Sprintf("Updated %s", info.Name))
					s.appService.forceRefreshResults()
				}()
			}, s.closeModal)
	}
}

// handleUpdateAllPackagesEvent is called when the user presses the update all key (Ctrl+U).
func (s *IOService) handleUpdateAllPackagesEvent() {
	s.showModal("Are you sure you want to update all Packages?", func() {
		s.closeModal()
		s.layout.GetOutput().Clear()
		go func() {
			s.layout.GetNotifier().ShowWarning("Updating all Packages...")
			if err := s.brewService.UpdateAllPackages(s.appService.app, s.layout.GetOutput().View()); err != nil {
				s.layout.GetNotifier().ShowError("Failed to update all Packages")
				return
			}
			s.layout.GetNotifier().ShowSuccess("Updated all Packages")
			s.appService.forceRefreshResults()
		}()
	}, s.closeModal)
}
