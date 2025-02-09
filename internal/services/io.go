package services

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
)

func (s *AppService) handleKeyEventInput(event *tcell.EventKey) *tcell.EventKey {
	if s.LayoutService.GetSearchField().HasFocus() {
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
			}
			if action, exists := runeActions[event.Rune()]; exists {
				action()
			}
		},
		tcell.KeyCtrlU: s.handleUpdateHomebrewEvent,
		tcell.KeyEsc: func() {
			s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetTableResult())
		},
	}

	if action, exists := keyActions[event.Key()]; exists {
		action()
		return nil
	}

	return event
}

func (s *AppService) handleSearchFieldEvent() {
	s.app.SetFocus(s.LayoutService.GetSearchField())
}

func (s *AppService) handleQuitEvent() {
	s.app.Stop()
}

func (s *AppService) handleUpdateHomebrewEvent() {
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
}

func (s *AppService) handleFilterPackagesEvent() {
	s.showOnlyInstalled = !s.showOnlyInstalled
	if s.showOnlyInstalled {
		s.LayoutService.GetSearchField().SetLabel("Search (Installed): ")
	} else {
		s.LayoutService.GetSearchField().SetLabel("Search (All): ")
	}
	s.applySearchFilter(s.LayoutService.GetSearchField().GetText())
	s.LayoutService.GetTableResult().ScrollToBeginning()
}

func (s *AppService) handleInstallPackageEvent() {
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
}

func (s *AppService) handleRemovePackageEvent() {
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
}

func (s *AppService) handleUpdatePackageEvent() {
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
}
