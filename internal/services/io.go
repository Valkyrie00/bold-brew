package services

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
)

func (s *AppService) handleKeyEventInput(event *tcell.EventKey) *tcell.EventKey {
	if s.layout.GetSearch().Field().HasFocus() {
		return event
	}

	keyActions := map[tcell.Key]func(){
		tcell.KeyRune: func() {
			runeActions := map[rune]func(){
				'q': s.handleQuitEvent,
				'u': s.handleUpdatePackageEvent,
				'r': s.handleRemovePackageEvent,
				'i': s.handleInstallPackageEvent,
				'/': s.handleSearchFieldEvent,
				'f': s.handleFilterPackagesEvent,
				'o': s.handleFilterOutdatedPackagesEvent, // New key binding for filtering outdated packages
			}
			if action, exists := runeActions[event.Rune()]; exists {
				action()
			}
		},
		tcell.KeyCtrlU: s.handleUpdateAllPackagesEvent,
		tcell.KeyEsc: func() {
			s.app.SetRoot(s.layout.Root(), true)
			s.app.SetFocus(s.layout.GetTable().View())
		},
	}

	if action, exists := keyActions[event.Key()]; exists {
		action()
		return nil
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
	s.showOnlyInstalled = !s.showOnlyInstalled
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
	s.showOnlyOutdated = !s.showOnlyOutdated
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
