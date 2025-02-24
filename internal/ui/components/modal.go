package components

import (
	"bbrew/internal/ui/theme"
	"github.com/rivo/tview"
)

type Modal struct {
	view  *tview.Modal
	theme *theme.Theme
}

func NewModal(theme *theme.Theme) *Modal {
	modal := tview.NewModal().
		SetBackgroundColor(theme.ModalBgColor).
		SetTextColor(theme.DefaultTextColor).
		SetButtonBackgroundColor(theme.ButtonBgColor).
		SetButtonTextColor(theme.ButtonTextColor)

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
		AddButtons([]string{"Confirm", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, _ string) {
			if buttonIndex == 0 {
				confirmFunc()
			} else if buttonIndex == 1 {
				cancelFunc()
			}
		})

	return m.view
}
