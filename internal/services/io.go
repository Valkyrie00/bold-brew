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
	FilterCasks
)

// IOAction represents an input/output action that can be triggered by a key event.
type IOAction struct {
	Key            tcell.Key
	Rune           rune
	Name           string
	KeySlug        string
	Action         func()
	HideFromLegend bool // If true, this action won't appear in the legend bar
}

func (k *IOAction) SetAction(action func()) {
	k.Action = action
}

// IOServiceInterface defines the interface for handling input/output actions in the application.
type IOServiceInterface interface {
	HandleKeyEventInput(event *tcell.EventKey) *tcell.EventKey
	EnableBrewfileMode()
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
	ActionFilterCasks     *IOAction
	ActionInstall         *IOAction
	ActionUpdate          *IOAction
	ActionRemove          *IOAction
	ActionUpdateAll       *IOAction
	ActionInstallAll      *IOAction
	ActionRemoveAll       *IOAction
	ActionHelp            *IOAction
	ActionBack            *IOAction
	ActionQuit            *IOAction
}

var NewIOService = func(appService *AppService, brewService BrewServiceInterface) IOServiceInterface {
	s := &IOService{
		appService:  appService,
		layout:      appService.GetLayout(),
		brewService: brewService,
	}

	// Initialize key actions with their respective keys, runes, and names.
	s.ActionSearch = &IOAction{Key: tcell.KeyRune, Rune: '/', KeySlug: "/", Name: "Search"}
	s.ActionFilterInstalled = &IOAction{Key: tcell.KeyRune, Rune: 'f', KeySlug: "f", Name: "Installed"}
	s.ActionFilterOutdated = &IOAction{Key: tcell.KeyRune, Rune: 'o', KeySlug: "o", Name: "Outdated", HideFromLegend: true}
	s.ActionFilterLeaves = &IOAction{Key: tcell.KeyRune, Rune: 'l', KeySlug: "l", Name: "Leaves", HideFromLegend: true}
	s.ActionFilterCasks = &IOAction{Key: tcell.KeyRune, Rune: 'c', KeySlug: "c", Name: "Casks", HideFromLegend: true}
	s.ActionInstall = &IOAction{Key: tcell.KeyRune, Rune: 'i', KeySlug: "i", Name: "Install"}
	s.ActionUpdate = &IOAction{Key: tcell.KeyRune, Rune: 'u', KeySlug: "u", Name: "Update"}
	s.ActionRemove = &IOAction{Key: tcell.KeyRune, Rune: 'r', KeySlug: "r", Name: "Remove"}
	s.ActionUpdateAll = &IOAction{Key: tcell.KeyCtrlU, Rune: 0, KeySlug: "ctrl+u", Name: "Update All", HideFromLegend: true}
	s.ActionInstallAll = &IOAction{Key: tcell.KeyCtrlA, Rune: 0, KeySlug: "ctrl+a", Name: "Install All (Brewfile)"}
	s.ActionRemoveAll = &IOAction{Key: tcell.KeyCtrlR, Rune: 0, KeySlug: "ctrl+r", Name: "Remove All (Brewfile)"}
	s.ActionHelp = &IOAction{Key: tcell.KeyRune, Rune: '?', KeySlug: "?", Name: "Help"}
	s.ActionBack = &IOAction{Key: tcell.KeyEsc, Rune: 0, KeySlug: "esc", Name: "Back to Table", HideFromLegend: true}
	s.ActionQuit = &IOAction{Key: tcell.KeyRune, Rune: 'q', KeySlug: "q", Name: "Quit", HideFromLegend: true}

	// Define actions for each key input,
	s.ActionSearch.SetAction(s.handleSearchFieldEvent)
	s.ActionFilterInstalled.SetAction(s.handleFilterPackagesEvent)
	s.ActionFilterOutdated.SetAction(s.handleFilterOutdatedPackagesEvent)
	s.ActionFilterLeaves.SetAction(s.handleFilterLeavesEvent)
	s.ActionFilterCasks.SetAction(s.handleFilterCasksEvent)
	s.ActionInstall.SetAction(s.handleInstallPackageEvent)
	s.ActionUpdate.SetAction(s.handleUpdatePackageEvent)
	s.ActionRemove.SetAction(s.handleRemovePackageEvent)
	s.ActionUpdateAll.SetAction(s.handleUpdateAllPackagesEvent)
	s.ActionInstallAll.SetAction(s.handleInstallAllPackagesEvent)
	s.ActionRemoveAll.SetAction(s.handleRemoveAllPackagesEvent)
	s.ActionHelp.SetAction(s.handleHelpEvent)
	s.ActionBack.SetAction(s.handleBack)
	s.ActionQuit.SetAction(s.handleQuitEvent)

	// Add all actions to the keyActions slice
	// Note: ActionInstallAll and ActionRemoveAll will be added dynamically if in Brewfile mode
	s.keyActions = []*IOAction{
		s.ActionSearch,
		s.ActionFilterInstalled,
		s.ActionFilterOutdated,
		s.ActionFilterLeaves,
		s.ActionFilterCasks,
		s.ActionInstall,
		s.ActionUpdate,
		s.ActionRemove,
		s.ActionUpdateAll,
		s.ActionHelp,
		s.ActionBack,
		s.ActionQuit,
	}

	// Convert keyActions to legend entries
	s.updateLegendEntries()
	return s
}

// updateLegendEntries updates the legend entries based on current keyActions
func (s *IOService) updateLegendEntries() {
	s.legendEntries = make([]struct{ KeySlug, Name string }, 0, len(s.keyActions))
	for _, input := range s.keyActions {
		if !input.HideFromLegend {
			s.legendEntries = append(s.legendEntries, struct{ KeySlug, Name string }{KeySlug: input.KeySlug, Name: input.Name})
		}
	}
	s.layout.GetLegend().SetLegend(s.legendEntries, "")
}

// EnableBrewfileMode enables Brewfile mode, adding Install All and Remove All actions to the legend
func (s *IOService) EnableBrewfileMode() {
	// Add Install All and Remove All actions after Update All
	newActions := []*IOAction{}
	for _, action := range s.keyActions {
		newActions = append(newActions, action)
		if action == s.ActionUpdateAll {
			newActions = append(newActions, s.ActionInstallAll, s.ActionRemoveAll)
		}
	}
	s.keyActions = newActions
	s.updateLegendEntries()
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

// handleHelpEvent shows the help screen with all keyboard shortcuts.
func (s *IOService) handleHelpEvent() {
	helpScreen := s.layout.GetHelpScreen()
	helpScreen.SetBrewfileMode(s.appService.IsBrewfileMode())
	helpPages := helpScreen.Build(s.layout.Root())

	// Set up key handler to close help on any key press
	helpPages.SetInputCapture(func(_ *tcell.EventKey) *tcell.EventKey {
		// Close help and return to main view
		s.appService.GetApp().SetRoot(s.layout.Root(), true)
		s.appService.GetApp().SetFocus(s.layout.GetTable().View())
		return nil
	})

	s.appService.GetApp().SetRoot(helpPages, true)
}

// handleFilterEvent toggles the filter for installed or outdated packages based on the provided filter type.
func (s *IOService) handleFilterEvent(filterType FilterType) {
	s.layout.GetLegend().SetLegend(s.legendEntries, "")

	switch filterType {
	case FilterInstalled:
		if s.appService.showOnlyOutdated || s.appService.showOnlyLeaves || s.appService.showOnlyCasks {
			s.appService.showOnlyOutdated = false
			s.appService.showOnlyLeaves = false
			s.appService.showOnlyCasks = false
			s.appService.showOnlyInstalled = true
		} else {
			s.appService.showOnlyInstalled = !s.appService.showOnlyInstalled
		}
	case FilterOutdated:
		if s.appService.showOnlyInstalled || s.appService.showOnlyLeaves || s.appService.showOnlyCasks {
			s.appService.showOnlyInstalled = false
			s.appService.showOnlyLeaves = false
			s.appService.showOnlyCasks = false
			s.appService.showOnlyOutdated = true
		} else {
			s.appService.showOnlyOutdated = !s.appService.showOnlyOutdated
		}
	case FilterLeaves:
		if s.appService.showOnlyInstalled || s.appService.showOnlyOutdated || s.appService.showOnlyCasks {
			s.appService.showOnlyInstalled = false
			s.appService.showOnlyOutdated = false
			s.appService.showOnlyCasks = false
			s.appService.showOnlyLeaves = true
		} else {
			s.appService.showOnlyLeaves = !s.appService.showOnlyLeaves
		}
	case FilterCasks:
		if s.appService.showOnlyInstalled || s.appService.showOnlyOutdated || s.appService.showOnlyLeaves {
			s.appService.showOnlyInstalled = false
			s.appService.showOnlyOutdated = false
			s.appService.showOnlyLeaves = false
			s.appService.showOnlyCasks = true
		} else {
			s.appService.showOnlyCasks = !s.appService.showOnlyCasks
		}
	}

	// Update the search field label and legend based on the current filter state
	baseLabel := "Search"
	if s.appService.IsBrewfileMode() {
		baseLabel = "Search (Brewfile"
	}

	if s.appService.showOnlyOutdated {
		if s.appService.IsBrewfileMode() {
			s.layout.GetSearch().Field().SetLabel(baseLabel + " - Outdated): ")
		} else {
			s.layout.GetSearch().Field().SetLabel("Search (Outdated): ")
		}
		s.layout.GetLegend().SetLegend(s.legendEntries, s.ActionFilterOutdated.KeySlug)
	} else if s.appService.showOnlyInstalled {
		if s.appService.IsBrewfileMode() {
			s.layout.GetSearch().Field().SetLabel(baseLabel + " - Installed): ")
		} else {
			s.layout.GetSearch().Field().SetLabel("Search (Installed): ")
		}
		s.layout.GetLegend().SetLegend(s.legendEntries, s.ActionFilterInstalled.KeySlug)
	} else if s.appService.showOnlyLeaves {
		if s.appService.IsBrewfileMode() {
			s.layout.GetSearch().Field().SetLabel(baseLabel + " - Leaves): ")
		} else {
			s.layout.GetSearch().Field().SetLabel("Search (Leaves): ")
		}
		s.layout.GetLegend().SetLegend(s.legendEntries, s.ActionFilterLeaves.KeySlug)
	} else if s.appService.showOnlyCasks {
		if s.appService.IsBrewfileMode() {
			s.layout.GetSearch().Field().SetLabel(baseLabel + " - Casks): ")
		} else {
			s.layout.GetSearch().Field().SetLabel("Search (Casks): ")
		}
		s.layout.GetLegend().SetLegend(s.legendEntries, s.ActionFilterCasks.KeySlug)
	} else {
		if s.appService.IsBrewfileMode() {
			s.layout.GetSearch().Field().SetLabel(baseLabel + "): ")
		} else {
			s.layout.GetSearch().Field().SetLabel("Search (All): ")
		}
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

// handleFilterCasksEvent toggles the filter for cask packages only
func (s *IOService) handleFilterCasksEvent() {
	s.handleFilterEvent(FilterCasks)
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

// handleInstallAllPackagesEvent is called when the user presses the install all key (Ctrl+A).
// This is only available in Brewfile mode and installs all packages from the Brewfile.
func (s *IOService) handleInstallAllPackagesEvent() {
	if !s.appService.IsBrewfileMode() {
		return // Only available in Brewfile mode
	}

	packages := *s.appService.GetBrewfilePackages()
	if len(packages) == 0 {
		s.layout.GetNotifier().ShowError("No packages found in Brewfile")
		return
	}

	// Count how many packages are not yet installed
	notInstalled := 0
	for _, pkg := range packages {
		if !pkg.LocallyInstalled {
			notInstalled++
		}
	}

	message := fmt.Sprintf("Install all packages from Brewfile?\n\nTotal: %d packages\nNot installed: %d", len(packages), notInstalled)

	s.showModal(message, func() {
		s.closeModal()
		s.layout.GetOutput().Clear()
		go func() {
			// Install all packages with progress notifications
			current := 0
			total := len(packages)

			for _, pkg := range packages {
				current++

				// Check if package is already installed
				if pkg.LocallyInstalled {
					s.layout.GetNotifier().ShowWarning(fmt.Sprintf("[%d/%d] Skipping %s (already installed)", current, total, pkg.Name))
					s.appService.app.QueueUpdateDraw(func() {
						fmt.Fprintf(s.layout.GetOutput().View(), "[SKIP] %s (already installed)\n", pkg.Name)
					})
					continue
				}

				// Show progress in notifier
				s.layout.GetNotifier().ShowWarning(fmt.Sprintf("[%d/%d] Installing %s...", current, total, pkg.Name))
				s.appService.app.QueueUpdateDraw(func() {
					fmt.Fprintf(s.layout.GetOutput().View(), "\n[INSTALL] Installing %s...\n", pkg.Name)
				})

				if err := s.brewService.InstallPackage(pkg, s.appService.app, s.layout.GetOutput().View()); err != nil {
					s.layout.GetNotifier().ShowError(fmt.Sprintf("[%d/%d] Failed to install %s", current, total, pkg.Name))
					s.appService.app.QueueUpdateDraw(func() {
						fmt.Fprintf(s.layout.GetOutput().View(), "[ERROR] Failed to install %s: %v\n", pkg.Name, err)
					})
					continue
				}

				s.appService.app.QueueUpdateDraw(func() {
					fmt.Fprintf(s.layout.GetOutput().View(), "[SUCCESS] %s installed successfully\n", pkg.Name)
				})
			}

			s.layout.GetNotifier().ShowSuccess(fmt.Sprintf("Completed! Processed %d packages", total))
			s.appService.forceRefreshResults()
		}()
	}, s.closeModal)
}

// handleRemoveAllPackagesEvent is called when the user presses the remove all key (Ctrl+R).
// This is only available in Brewfile mode and removes all installed packages from the Brewfile.
func (s *IOService) handleRemoveAllPackagesEvent() {
	if !s.appService.IsBrewfileMode() {
		return // Only available in Brewfile mode
	}

	packages := *s.appService.GetBrewfilePackages()
	if len(packages) == 0 {
		s.layout.GetNotifier().ShowError("No packages found in Brewfile")
		return
	}

	// Count how many packages are installed
	installed := 0
	for _, pkg := range packages {
		if pkg.LocallyInstalled {
			installed++
		}
	}

	if installed == 0 {
		s.layout.GetNotifier().ShowWarning("No packages to remove (none are installed)")
		return
	}

	message := fmt.Sprintf("Remove all installed packages from Brewfile?\n\nTotal: %d packages\nInstalled: %d", len(packages), installed)

	s.showModal(message, func() {
		s.closeModal()
		s.layout.GetOutput().Clear()
		go func() {
			// Remove all packages with progress notifications
			current := 0
			total := len(packages)

			for _, pkg := range packages {
				current++

				// Check if package is not installed
				if !pkg.LocallyInstalled {
					s.layout.GetNotifier().ShowWarning(fmt.Sprintf("[%d/%d] Skipping %s (not installed)", current, total, pkg.Name))
					s.appService.app.QueueUpdateDraw(func() {
						fmt.Fprintf(s.layout.GetOutput().View(), "[SKIP] %s (not installed)\n", pkg.Name)
					})
					continue
				}

				// Show progress in notifier
				s.layout.GetNotifier().ShowWarning(fmt.Sprintf("[%d/%d] Removing %s...", current, total, pkg.Name))
				s.appService.app.QueueUpdateDraw(func() {
					fmt.Fprintf(s.layout.GetOutput().View(), "\n[REMOVE] Removing %s...\n", pkg.Name)
				})

				if err := s.brewService.RemovePackage(pkg, s.appService.app, s.layout.GetOutput().View()); err != nil {
					s.layout.GetNotifier().ShowError(fmt.Sprintf("[%d/%d] Failed to remove %s", current, total, pkg.Name))
					s.appService.app.QueueUpdateDraw(func() {
						fmt.Fprintf(s.layout.GetOutput().View(), "[ERROR] Failed to remove %s: %v\n", pkg.Name, err)
					})
					continue
				}

				s.appService.app.QueueUpdateDraw(func() {
					fmt.Fprintf(s.layout.GetOutput().View(), "[SUCCESS] %s removed successfully\n", pkg.Name)
				})
			}

			s.layout.GetNotifier().ShowSuccess(fmt.Sprintf("Completed! Processed %d packages", total))
			s.appService.forceRefreshResults()
		}()
	}, s.closeModal)
}
