package services

import (
	"bbrew/internal/models"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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
	if s.showOnlyInstalled {
		s.layout.GetSearch().Field().SetLabel("Search (Installed): ")
	} else {
		s.layout.GetSearch().Field().SetLabel("Search (All): ")
	}

	s.search(s.layout.GetSearch().Field().GetText(), true)
}

func (s *AppService) showModal(text string, confirmFunc func(), cancelFunc func()) {
	modal := s.layout.GenerateModal(text, func() {
		s.app.SetRoot(s.layout.Root(), true)
		confirmFunc()
	}, func() {
		s.app.SetRoot(s.layout.Root(), true)
		cancelFunc()
	})
	s.app.SetRoot(modal, true)
}

func (s *AppService) handleInstallPackageEvent() {
	row, _ := s.layout.GetTable().View().GetSelection()
	if row > 0 {
		info := (*s.filteredPackages)[row-1]
		s.showModal(
			fmt.Sprintf("Are you sure you want to install the package: %s?", info.Name),
			s.createModalConfirmHandler(info, "Installing", s.CommandService.InstallPackage, "Installed"),
			s.resetViewAfterModal,
		)
	}
}

func (s *AppService) handleRemovePackageEvent() {
	row, _ := s.layout.GetTable().View().GetSelection()
	if row > 0 {
		info := (*s.filteredPackages)[row-1]
		s.showModal(
			fmt.Sprintf("Are you sure you want to remove the package: %s?", info.Name),
			s.createModalConfirmHandler(info, "Removing", s.CommandService.RemovePackage, "Removed"),
			s.resetViewAfterModal,
		)
	}
}

func (s *AppService) handleUpdatePackageEvent() {
	row, _ := s.layout.GetTable().View().GetSelection()
	if row > 0 {
		info := (*s.filteredPackages)[row-1]
		s.showModal(
			fmt.Sprintf("Are you sure you want to update the package: %s?", info.Name),
			s.createModalConfirmHandler(info, "Updating", s.CommandService.UpdatePackage, "Updated"),
			s.resetViewAfterModal,
		)
	}
}

func (s *AppService) handleUpdateAllPackagesEvent() {
	s.showModal("Are you sure you want to update all packages?", func() {
		s.layout.GetDetails().View().Clear()
		go func() {
			s.layout.ShowWarningNotification("Updating all packages...")
			if err := s.CommandService.UpdateAllPackages(s.app, s.layout.GetDetails().View()); err != nil {
				s.layout.ShowWarningNotification("Failed to update all packages")
				return
			}
			s.layout.ShowSuccessNotification("Updated all packages")
			s.forceRefreshResults()
		}()
	}, s.resetViewAfterModal)
}

func (s *AppService) resetViewAfterModal() {
	s.app.SetFocus(s.layout.GetTable().View())
}

func (s *AppService) createModalConfirmHandler(info models.Formula, actionName string, action func(models.Formula, *tview.Application, *tview.TextView) error, completedAction string) func() {
	return func() {
		s.resetViewAfterModal()
		s.layout.GetOutput().Clear()
		go func() {
			s.layout.ShowWarningNotification(fmt.Sprintf("%s %s...", actionName, info.Name))
			if err := action(info, s.app, s.layout.GetOutput().View()); err != nil {
				s.layout.ShowWarningNotification(fmt.Sprintf("Failed to %s %s", actionName, info.Name))
				return
			}
			s.layout.ShowSuccessNotification(fmt.Sprintf("%s %s", info.Name, completedAction))
			s.forceRefreshResults()
		}()
	}
}
