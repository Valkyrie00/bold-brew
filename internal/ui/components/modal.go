package components

import (
	"bbrew/internal/ui/theme"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Modal struct {
	view  *tview.Modal
	theme *theme.Theme
}

func NewModal(theme *theme.Theme) *Modal {
	// Use green background with black text for activated button
	// Black text ensures consistent visibility across all terminal themes
	activatedStyle := tcell.StyleDefault.
		Background(theme.SuccessColor).
		Foreground(tcell.ColorBlack).
		Bold(true)

	modal := tview.NewModal().
		SetBackgroundColor(theme.ModalBgColor).
		SetTextColor(theme.DefaultTextColor).
		SetButtonBackgroundColor(theme.ButtonBgColor).
		SetButtonTextColor(theme.ButtonTextColor).
		SetButtonActivatedStyle(activatedStyle)

	return &Modal{
		view:  modal,
		theme: theme,
	}
}

func (m *Modal) View() *tview.Modal {
	return m.view
}

func (m *Modal) Build(text string, confirmFunc func(), cancelFunc func()) *tview.Modal {
	m.view.ClearButtons()
	m.view.
		SetText(text).
		// Add padding to button labels with spaces for better visual appearance
		AddButtons([]string{"  Confirm  ", "  Cancel  "}).
		SetDoneFunc(func(buttonIndex int, _ string) {
			switch buttonIndex {
			case 0:
				confirmFunc()
			case 1:
				cancelFunc()
			}
		})

	return m.view
}
