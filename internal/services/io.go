package services

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
)

type IOMap struct {
	Key     tcell.Key
	Rune    rune
	Name    string
	KeySlug string
	Action  func()
}

var (
	IO_SEARCH           = IOMap{Key: tcell.KeyRune, Rune: '/', KeySlug: "/", Name: "Search"}
	IO_FILTER_INSTALLED = IOMap{Key: tcell.KeyRune, Rune: 'f', KeySlug: "f", Name: "Filter Installed"}
	IO_FILTER_OUTDATED  = IOMap{Key: tcell.KeyRune, Rune: 'o', KeySlug: "o", Name: "Filter Outdated"}
	IO_INSTALL          = IOMap{Key: tcell.KeyRune, Rune: 'i', KeySlug: "i", Name: "Install"}
	IO_UPDATE           = IOMap{Key: tcell.KeyRune, Rune: 'u', KeySlug: "u", Name: "Update"}
	IO_REMOVE           = IOMap{Key: tcell.KeyRune, Rune: 'r', KeySlug: "r", Name: "Remove"}
	IO_UPDATE_ALL       = IOMap{Key: tcell.KeyCtrlU, Rune: 0, KeySlug: "ctrl+u", Name: "Update All"}
	IO_BACK             = IOMap{Key: tcell.KeyEsc, Rune: 0, KeySlug: "esc", Name: "Back to Table"}
	IO_QUIT             = IOMap{Key: tcell.KeyRune, Rune: 'q', KeySlug: "q", Name: "Quit"}
)

var IOKeys = []*IOMap{&IO_SEARCH, &IO_FILTER_INSTALLED, &IO_FILTER_OUTDATED, &IO_INSTALL, &IO_UPDATE, &IO_UPDATE_ALL, &IO_REMOVE, &IO_BACK, &IO_QUIT}

func (s *AppService) GetLegendText(activeSection string) (legendText string) {
	for i, legend := range IOKeys {
		if legend.KeySlug == activeSection {
			legendText += s.layout.GetLegend().GetFormattedLabel(legend.KeySlug, legend.Name, true)
		} else {
			legendText += s.layout.GetLegend().GetFormattedLabel(legend.KeySlug, legend.Name, false)
		}

		if i < len(IOKeys)-1 {
			legendText += " | "
		}
	}

	return legendText
}

func (s *AppService) handleKeyEventInput(event *tcell.EventKey) *tcell.EventKey {
	if s.layout.GetSearch().Field().HasFocus() {
		return event
	}

	// Define actions for each key input
	IO_QUIT.Action = s.handleQuitEvent
	IO_UPDATE.Action = s.handleUpdatePackageEvent
	IO_UPDATE_ALL.Action = s.handleUpdateAllPackagesEvent
	IO_REMOVE.Action = s.handleRemovePackageEvent
	IO_INSTALL.Action = s.handleInstallPackageEvent
	IO_SEARCH.Action = s.handleSearchFieldEvent
	IO_FILTER_INSTALLED.Action = s.handleFilterPackagesEvent
	IO_FILTER_OUTDATED.Action = s.handleFilterOutdatedPackagesEvent
	IO_BACK.Action = func() {
		s.app.SetRoot(s.layout.Root(), true)
		s.app.SetFocus(s.layout.GetTable().View())
	}

	for _, input := range IOKeys {
		if input.Key == event.Key() && input.Rune == event.Rune() { // Check Rune
			if input.Action != nil {
				input.Action()
				return nil
			}
		} else if input.Key == event.Key() { // Check Key only
			if input.Action != nil {
				input.Action()
				return nil
			}
		}
	}

	return event
}

func (s *AppService) handleSearchFieldEvent() {
	s.app.SetFocus(s.layout.GetSearch().Field())
}

func (s *AppService) handleQuitEvent() {
	s.app.Stop()
}

func (s *AppService) handleFilterPackagesEvent() {
	if s.showOnlyOutdated {
		s.showOnlyOutdated = false
		s.showOnlyInstalled = true
	} else {
		s.showOnlyInstalled = !s.showOnlyInstalled
	}

	s.layout.GetLegend().SetText(s.GetLegendText(IO_FILTER_INSTALLED.KeySlug))

	// Update the search field label
	if s.showOnlyOutdated {
		s.layout.GetSearch().Field().SetLabel("Search (Outdated): ")
	} else if s.showOnlyInstalled {
		s.layout.GetSearch().Field().SetLabel("Search (Installed): ")
	} else {
		s.layout.GetSearch().Field().SetLabel("Search (All): ")
	}

	s.search(s.layout.GetSearch().Field().GetText(), true)
}

func (s *AppService) handleFilterOutdatedPackagesEvent() {
	if s.showOnlyInstalled {
		s.showOnlyInstalled = false
		s.showOnlyOutdated = true
	} else {
		s.showOnlyOutdated = !s.showOnlyOutdated
	}

	// Update the search field label
	if s.showOnlyOutdated {
		s.layout.GetSearch().Field().SetLabel("Search (Outdated): ")
	} else if s.showOnlyInstalled {
		s.layout.GetSearch().Field().SetLabel("Search (Installed): ")
	} else {
		s.layout.GetSearch().Field().SetLabel("Search (All): ")
	}

	s.search(s.layout.GetSearch().Field().GetText(), true)
}

func (s *AppService) showModal(text string, confirmFunc func(), cancelFunc func()) {
	modal := s.layout.GetModal().Build(text, confirmFunc, cancelFunc)
	s.app.SetRoot(modal, true)
}

func (s *AppService) closeModal() {
	s.app.SetRoot(s.layout.Root(), true)
	s.app.SetFocus(s.layout.GetTable().View())
}

func (s *AppService) handleInstallPackageEvent() {
	row, _ := s.layout.GetTable().View().GetSelection()
	if row > 0 {
		info := (*s.filteredPackages)[row-1]
		s.showModal(
			fmt.Sprintf("Are you sure you want to install the package: %s?", info.Name),
			func() {
				s.closeModal()
				s.layout.GetOutput().Clear()
				go func() {
					s.layout.GetNotifier().ShowWarning(fmt.Sprintf("Installing %s...", info.Name))
					if err := s.CommandService.InstallPackage(info, s.app, s.layout.GetOutput().View()); err != nil {
						s.layout.GetNotifier().ShowError(fmt.Sprintf("Failed to install %s", info.Name))
						return
					}
					s.layout.GetNotifier().ShowSuccess(fmt.Sprintf("Installed %s", info.Name))
					s.forceRefreshResults()
				}()
			}, s.closeModal)
	}
}

func (s *AppService) handleRemovePackageEvent() {
	row, _ := s.layout.GetTable().View().GetSelection()
	if row > 0 {
		info := (*s.filteredPackages)[row-1]
		s.showModal(
			fmt.Sprintf("Are you sure you want to remove the package: %s?", info.Name),
			func() {
				s.closeModal()
				s.layout.GetOutput().Clear()
				go func() {
					s.layout.GetNotifier().ShowWarning(fmt.Sprintf("Removing %s...", info.Name))
					if err := s.CommandService.RemovePackage(info, s.app, s.layout.GetOutput().View()); err != nil {
						s.layout.GetNotifier().ShowError(fmt.Sprintf("Failed to remove %s", info.Name))
						return
					}
					s.layout.GetNotifier().ShowSuccess(fmt.Sprintf("Removed %s", info.Name))
					s.forceRefreshResults()
				}()
			}, s.closeModal)
	}
}

func (s *AppService) handleUpdatePackageEvent() {
	row, _ := s.layout.GetTable().View().GetSelection()
	if row > 0 {
		info := (*s.filteredPackages)[row-1]
		s.showModal(
			fmt.Sprintf("Are you sure you want to update the package: %s?", info.Name),
			func() {
				s.closeModal()
				s.layout.GetOutput().Clear()
				go func() {
					s.layout.GetNotifier().ShowWarning(fmt.Sprintf("Updating %s...", info.Name))
					if err := s.CommandService.UpdatePackage(info, s.app, s.layout.GetOutput().View()); err != nil {
						s.layout.GetNotifier().ShowError(fmt.Sprintf("Failed to update %s", info.Name))
						return
					}
					s.layout.GetNotifier().ShowSuccess(fmt.Sprintf("Updated %s", info.Name))
					s.forceRefreshResults()
				}()
			}, s.closeModal)
	}
}

func (s *AppService) handleUpdateAllPackagesEvent() {
	s.showModal("Are you sure you want to update all packages?", func() {
		s.closeModal()
		s.layout.GetOutput().Clear()
		go func() {
			s.layout.GetNotifier().ShowWarning("Updating all packages...")
			if err := s.CommandService.UpdateAllPackages(s.app, s.layout.GetOutput().View()); err != nil {
				s.layout.GetNotifier().ShowError("Failed to update all packages")
				return
			}
			s.layout.GetNotifier().ShowSuccess("Updated all packages")
			s.forceRefreshResults()
		}()
	}, s.closeModal)
}
