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
		tcell.KeyCtrlU: s.handleUpdateAllPackagesEvent,
		tcell.KeyEsc: func() {
			s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetResultTable())
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

//lint:ignore U1000 Temporarily unused
func (s *AppService) handleUpdateHomebrewEvent() {
	modal := s.LayoutService.GenerateModal("Are you sure you want to update Homebrew?", func() {
		s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetResultTable())
		s.LayoutService.GetOutputView().Clear()
		go func() {
			if err := s.CommandService.UpdateHomebrew(s.app, s.LayoutService.GetOutputView()); err == nil {
				s.forceRefreshResults()
			}
		}()
	}, func() {
		s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetResultTable())
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
	s.search(s.LayoutService.GetSearchField().GetText())
	s.LayoutService.GetResultTable().ScrollToBeginning()
}

func (s *AppService) handleInstallPackageEvent() {
	row, _ := s.LayoutService.GetResultTable().GetSelection()
	if row > 0 {
		info := (*s.filteredPackages)[row-1]
		modal := s.LayoutService.GenerateModal(fmt.Sprintf("Are you sure you want to install the package: %s?", info.Name), func() {
			s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetResultTable())
			s.LayoutService.GetOutputView().Clear()
			go func() {
				if err := s.CommandService.InstallPackage(info, s.app, s.LayoutService.GetOutputView()); err == nil {
					s.forceRefreshResults()
				}
			}()
		}, func() {
			s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetResultTable())
		})
		s.app.SetRoot(modal, true).SetFocus(modal)
	}
}

func (s *AppService) handleRemovePackageEvent() {
	row, _ := s.LayoutService.GetResultTable().GetSelection()
	if row > 0 {
		info := (*s.filteredPackages)[row-1]
		modal := s.LayoutService.GenerateModal(fmt.Sprintf("Are you sure you want to remove the package: %s?", info.Name), func() {
			s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetResultTable())
			s.LayoutService.GetOutputView().Clear()
			go func() {
				if err := s.CommandService.RemovePackage(info, s.app, s.LayoutService.GetOutputView()); err == nil {
					s.forceRefreshResults()
				}
			}()
		}, func() {
			s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetResultTable())
		})
		s.app.SetRoot(modal, true).SetFocus(modal)
	}
}

func (s *AppService) handleUpdatePackageEvent() {
	row, _ := s.LayoutService.GetResultTable().GetSelection()
	if row > 0 {
		info := (*s.filteredPackages)[row-1]
		modal := s.LayoutService.GenerateModal(fmt.Sprintf("Are you sure you want to update the package: %s?", info.Name), func() {
			s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetResultTable())
			s.LayoutService.GetOutputView().Clear()
			go func() {
				if err := s.CommandService.UpdatePackage(info, s.app, s.LayoutService.GetOutputView()); err == nil {
					s.forceRefreshResults()
				}
			}()
		}, func() {
			s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetResultTable())
		})
		s.app.SetRoot(modal, true).SetFocus(modal)
	}
}

func (s *AppService) handleUpdateAllPackagesEvent() {
	modal := s.LayoutService.GenerateModal("Are you sure you want to update all packages?", func() {
		s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetResultTable())
		s.LayoutService.GetOutputView().Clear()
		go func() {
			if err := s.CommandService.UpdateAllPackages(s.app, s.LayoutService.GetOutputView()); err == nil {
				s.forceRefreshResults()
			}
		}()
	}, func() {
		s.app.SetRoot(s.LayoutService.GetGrid(), true).SetFocus(s.LayoutService.GetResultTable())
	})
	s.app.SetRoot(modal, true).SetFocus(modal)
}
