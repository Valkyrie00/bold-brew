package services

import (
	"bbrew/internal/models"
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

// InputAction represents a user action that can be triggered by a key event.
type InputAction struct {
	Key            tcell.Key
	Rune           rune
	Name           string
	KeySlug        string
	Action         func()
	HideFromLegend bool // If true, this action won't appear in the legend bar
}

// InputServiceInterface defines the interface for handling user input actions.
type InputServiceInterface interface {
	HandleKeyEventInput(event *tcell.EventKey) *tcell.EventKey
	EnableBrewfileMode()
}

// InputService implements the InputServiceInterface and handles key events for the application.
type InputService struct {
	appService    *AppService
	layout        ui.LayoutInterface
	brewService   BrewServiceInterface
	keyActions    []*InputAction
	legendEntries []struct{ KeySlug, Name string }

	// Actions for each key input
	ActionSearch          *InputAction
	ActionFilterInstalled *InputAction
	ActionFilterOutdated  *InputAction
	ActionFilterLeaves    *InputAction
	ActionFilterCasks     *InputAction
	ActionInstall         *InputAction
	ActionUpdate          *InputAction
	ActionRemove          *InputAction
	ActionUpdateAll       *InputAction
	ActionInstallAll      *InputAction
	ActionRemoveAll       *InputAction
	ActionHelp            *InputAction
	ActionBack            *InputAction
	ActionQuit            *InputAction
}

var NewInputService = func(appService *AppService, brewService BrewServiceInterface) InputServiceInterface {
	s := &InputService{
		appService:  appService,
		layout:      appService.GetLayout(),
		brewService: brewService,
	}

	// Initialize actions with key bindings and handlers
	s.ActionSearch = &InputAction{
		Key: tcell.KeyRune, Rune: '/', KeySlug: "/", Name: "Search",
		Action: s.handleSearchFieldEvent,
	}
	s.ActionFilterInstalled = &InputAction{
		Key: tcell.KeyRune, Rune: 'f', KeySlug: "f", Name: "Installed",
		Action: s.handleFilterPackagesEvent,
	}
	s.ActionFilterOutdated = &InputAction{
		Key: tcell.KeyRune, Rune: 'o', KeySlug: "o", Name: "Outdated",
		Action: s.handleFilterOutdatedPackagesEvent, HideFromLegend: true,
	}
	s.ActionFilterLeaves = &InputAction{
		Key: tcell.KeyRune, Rune: 'l', KeySlug: "l", Name: "Leaves",
		Action: s.handleFilterLeavesEvent, HideFromLegend: true,
	}
	s.ActionFilterCasks = &InputAction{
		Key: tcell.KeyRune, Rune: 'c', KeySlug: "c", Name: "Casks",
		Action: s.handleFilterCasksEvent, HideFromLegend: true,
	}
	s.ActionInstall = &InputAction{
		Key: tcell.KeyRune, Rune: 'i', KeySlug: "i", Name: "Install",
		Action: s.handleInstallPackageEvent,
	}
	s.ActionUpdate = &InputAction{
		Key: tcell.KeyRune, Rune: 'u', KeySlug: "u", Name: "Update",
		Action: s.handleUpdatePackageEvent,
	}
	s.ActionRemove = &InputAction{
		Key: tcell.KeyRune, Rune: 'r', KeySlug: "r", Name: "Remove",
		Action: s.handleRemovePackageEvent,
	}
	s.ActionUpdateAll = &InputAction{
		Key: tcell.KeyCtrlU, Rune: 0, KeySlug: "ctrl+u", Name: "Update All",
		Action: s.handleUpdateAllPackagesEvent, HideFromLegend: true,
	}
	s.ActionInstallAll = &InputAction{
		Key: tcell.KeyCtrlA, Rune: 0, KeySlug: "ctrl+a", Name: "Install All (Brewfile)",
		Action: s.handleInstallAllPackagesEvent,
	}
	s.ActionRemoveAll = &InputAction{
		Key: tcell.KeyCtrlR, Rune: 0, KeySlug: "ctrl+r", Name: "Remove All (Brewfile)",
		Action: s.handleRemoveAllPackagesEvent,
	}
	s.ActionHelp = &InputAction{
		Key: tcell.KeyRune, Rune: '?', KeySlug: "?", Name: "Help",
		Action: s.handleHelpEvent,
	}
	s.ActionBack = &InputAction{
		Key: tcell.KeyEsc, Rune: 0, KeySlug: "esc", Name: "Back to Table",
		Action: s.handleBack, HideFromLegend: true,
	}
	s.ActionQuit = &InputAction{
		Key: tcell.KeyRune, Rune: 'q', KeySlug: "q", Name: "Quit",
		Action: s.handleQuitEvent, HideFromLegend: true,
	}

	// Build keyActions slice (InstallAll/RemoveAll added dynamically in Brewfile mode)
	s.keyActions = []*InputAction{
		s.ActionSearch, s.ActionFilterInstalled, s.ActionFilterOutdated,
		s.ActionFilterLeaves, s.ActionFilterCasks, s.ActionInstall,
		s.ActionUpdate, s.ActionRemove, s.ActionUpdateAll,
		s.ActionHelp, s.ActionBack, s.ActionQuit,
	}

	// Convert keyActions to legend entries
	s.updateLegendEntries()
	return s
}

// updateLegendEntries updates the legend entries based on current keyActions
func (s *InputService) updateLegendEntries() {
	s.legendEntries = make([]struct{ KeySlug, Name string }, 0, len(s.keyActions))
	for _, input := range s.keyActions {
		if !input.HideFromLegend {
			s.legendEntries = append(s.legendEntries, struct{ KeySlug, Name string }{KeySlug: input.KeySlug, Name: input.Name})
		}
	}
	s.layout.GetLegend().SetLegend(s.legendEntries, "")
}

// EnableBrewfileMode enables Brewfile mode, adding Install All and Remove All actions to the legend
func (s *InputService) EnableBrewfileMode() {
	// Add Install All and Remove All actions after Update All
	newActions := []*InputAction{}
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
func (s *InputService) HandleKeyEventInput(event *tcell.EventKey) *tcell.EventKey {
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
func (s *InputService) handleBack() {
	s.appService.GetApp().SetRoot(s.layout.Root(), true)
	s.appService.GetApp().SetFocus(s.layout.GetTable().View())
}

// handleSearchFieldEvent is called when the user presses the search key (/).
func (s *InputService) handleSearchFieldEvent() {
	s.appService.GetApp().SetFocus(s.layout.GetSearch().Field())
}

// handleQuitEvent is called when the user presses the quit key (q).
func (s *InputService) handleQuitEvent() {
	s.appService.GetApp().Stop()
}

// handleHelpEvent shows the help screen with all keyboard shortcuts.
func (s *InputService) handleHelpEvent() {
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
func (s *InputService) handleFilterEvent(filterType FilterType) {
	s.layout.GetLegend().SetLegend(s.legendEntries, "")

	// Get current state of the target filter
	isCurrentlyActive := s.isFilterActive(filterType)

	// Reset all filters
	s.appService.showOnlyInstalled = false
	s.appService.showOnlyOutdated = false
	s.appService.showOnlyLeaves = false
	s.appService.showOnlyCasks = false

	// If target wasn't active, activate it (otherwise it stays off = toggle off)
	if !isCurrentlyActive {
		s.setFilterActive(filterType, true)
	}

	// Update UI based on active filter
	s.updateFilterUI()
	s.appService.search(s.layout.GetSearch().Field().GetText(), true)
}

// isFilterActive checks if a specific filter is currently active.
func (s *InputService) isFilterActive(filterType FilterType) bool {
	switch filterType {
	case FilterInstalled:
		return s.appService.showOnlyInstalled
	case FilterOutdated:
		return s.appService.showOnlyOutdated
	case FilterLeaves:
		return s.appService.showOnlyLeaves
	case FilterCasks:
		return s.appService.showOnlyCasks
	}
	return false
}

// setFilterActive sets a specific filter's active state.
func (s *InputService) setFilterActive(filterType FilterType, active bool) {
	switch filterType {
	case FilterInstalled:
		s.appService.showOnlyInstalled = active
	case FilterOutdated:
		s.appService.showOnlyOutdated = active
	case FilterLeaves:
		s.appService.showOnlyLeaves = active
	case FilterCasks:
		s.appService.showOnlyCasks = active
	}
}

// updateFilterUI updates the search label and legend based on the current filter state.
func (s *InputService) updateFilterUI() {
	baseLabel := "Search"
	if s.appService.IsBrewfileMode() {
		baseLabel = "Search (Brewfile"
	}

	// Map filter states to their labels and legend keys
	filterConfig := []struct {
		active  bool
		suffix  string
		keySlug string
	}{
		{s.appService.showOnlyInstalled, "Installed", s.ActionFilterInstalled.KeySlug},
		{s.appService.showOnlyOutdated, "Outdated", s.ActionFilterOutdated.KeySlug},
		{s.appService.showOnlyLeaves, "Leaves", s.ActionFilterLeaves.KeySlug},
		{s.appService.showOnlyCasks, "Casks", s.ActionFilterCasks.KeySlug},
	}

	for _, cfg := range filterConfig {
		if cfg.active {
			if s.appService.IsBrewfileMode() {
				s.layout.GetSearch().Field().SetLabel(baseLabel + " - " + cfg.suffix + "): ")
			} else {
				s.layout.GetSearch().Field().SetLabel("Search (" + cfg.suffix + "): ")
			}
			s.layout.GetLegend().SetLegend(s.legendEntries, cfg.keySlug)
			return
		}
	}

	// No filter active
	if s.appService.IsBrewfileMode() {
		s.layout.GetSearch().Field().SetLabel(baseLabel + "): ")
	} else {
		s.layout.GetSearch().Field().SetLabel("Search (All): ")
	}
}

// handleFilterPackagesEvent toggles the filter for installed packages
func (s *InputService) handleFilterPackagesEvent() {
	s.handleFilterEvent(FilterInstalled)
}

// handleFilterOutdatedPackagesEvent toggles the filter for outdated packages
func (s *InputService) handleFilterOutdatedPackagesEvent() {
	s.handleFilterEvent(FilterOutdated)
}

// handleFilterLeavesEvent toggles the filter for leaf packages (installed on request)
func (s *InputService) handleFilterLeavesEvent() {
	s.handleFilterEvent(FilterLeaves)
}

// handleFilterCasksEvent toggles the filter for cask packages only
func (s *InputService) handleFilterCasksEvent() {
	s.handleFilterEvent(FilterCasks)
}

// showModal displays a modal dialog with the specified text and confirmation/cancellation actions.
// This is used for actions like installing, removing, or updating packages, invoking user confirmation.
func (s *InputService) showModal(text string, confirmFunc func(), cancelFunc func()) {
	modal := s.layout.GetModal().Build(text, confirmFunc, cancelFunc)
	s.appService.app.SetRoot(modal, true)
}

// closeModal closes the currently displayed modal dialog and returns focus to the main table view.
func (s *InputService) closeModal() {
	s.appService.app.SetRoot(s.layout.Root(), true)
	s.appService.app.SetFocus(s.layout.GetTable().View())
}

// handleInstallPackageEvent is called when the user presses the installation key (i).
func (s *InputService) handleInstallPackageEvent() {
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
func (s *InputService) handleRemovePackageEvent() {
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
func (s *InputService) handleUpdatePackageEvent() {
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
func (s *InputService) handleUpdateAllPackagesEvent() {
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

// batchOperation defines the configuration for a batch package operation.
type batchOperation struct {
	actionVerb    string // "Installing" or "Removing"
	actionTag     string // "INSTALL" or "REMOVE"
	skipCondition func(pkg models.Package) bool
	skipReason    string
	execute       func(pkg models.Package) error
}

// handleBatchPackageOperation processes multiple packages with progress notifications.
func (s *InputService) handleBatchPackageOperation(op batchOperation) {
	if !s.appService.IsBrewfileMode() {
		return
	}

	packages := *s.appService.GetBrewfilePackages()
	if len(packages) == 0 {
		s.layout.GetNotifier().ShowError("No packages found in Brewfile")
		return
	}

	// Count relevant packages
	actionable := 0
	for _, pkg := range packages {
		if !op.skipCondition(pkg) {
			actionable++
		}
	}

	if actionable == 0 {
		s.layout.GetNotifier().ShowWarning(fmt.Sprintf("No packages to process (%s)", op.skipReason))
		return
	}

	message := fmt.Sprintf("%s all packages from Brewfile?\n\nTotal: %d packages\nTo process: %d",
		op.actionVerb, len(packages), actionable)

	s.showModal(message, func() {
		s.closeModal()
		s.layout.GetOutput().Clear()
		go func() {
			current := 0
			total := len(packages)

			for _, pkg := range packages {
				current++
				pkgName := pkg.Name // Capture for closures

				if op.skipCondition(pkg) {
					s.layout.GetNotifier().ShowWarning(fmt.Sprintf("[%d/%d] Skipping %s (%s)", current, total, pkgName, op.skipReason))
					s.appService.app.QueueUpdateDraw(func() {
						fmt.Fprintf(s.layout.GetOutput().View(), "[SKIP] %s (%s)\n", pkgName, op.skipReason)
					})
					continue
				}

				s.layout.GetNotifier().ShowWarning(fmt.Sprintf("[%d/%d] %s %s...", current, total, op.actionVerb, pkgName))
				s.appService.app.QueueUpdateDraw(func() {
					fmt.Fprintf(s.layout.GetOutput().View(), "\n[%s] %s %s...\n", op.actionTag, op.actionVerb, pkgName)
				})

				if err := op.execute(pkg); err != nil {
					s.layout.GetNotifier().ShowError(fmt.Sprintf("[%d/%d] Failed to process %s", current, total, pkgName))
					s.appService.app.QueueUpdateDraw(func() {
						fmt.Fprintf(s.layout.GetOutput().View(), "[ERROR] Failed to process %s: %v\n", pkgName, err)
					})
					continue
				}

				s.appService.app.QueueUpdateDraw(func() {
					fmt.Fprintf(s.layout.GetOutput().View(), "[SUCCESS] %s processed successfully\n", pkgName)
				})
			}

			s.layout.GetNotifier().ShowSuccess(fmt.Sprintf("Completed! Processed %d packages", total))
			s.appService.forceRefreshResults()
		}()
	}, s.closeModal)
}

// handleInstallAllPackagesEvent is called when the user presses the install all key (Ctrl+A).
func (s *InputService) handleInstallAllPackagesEvent() {
	s.handleBatchPackageOperation(batchOperation{
		actionVerb:    "Installing",
		actionTag:     "INSTALL",
		skipCondition: func(pkg models.Package) bool { return pkg.LocallyInstalled },
		skipReason:    "already installed",
		execute: func(pkg models.Package) error {
			return s.brewService.InstallPackage(pkg, s.appService.app, s.layout.GetOutput().View())
		},
	})
}

// handleRemoveAllPackagesEvent is called when the user presses the remove all key (Ctrl+R).
func (s *InputService) handleRemoveAllPackagesEvent() {
	s.handleBatchPackageOperation(batchOperation{
		actionVerb:    "Removing",
		actionTag:     "REMOVE",
		skipCondition: func(pkg models.Package) bool { return !pkg.LocallyInstalled },
		skipReason:    "not installed",
		execute: func(pkg models.Package) error {
			return s.brewService.RemovePackage(pkg, s.appService.app, s.layout.GetOutput().View())
		},
	})
}
