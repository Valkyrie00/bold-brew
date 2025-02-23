package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Modal struct {
	view *tview.Modal
}

func NewModal() *Modal {
	modal := tview.NewModal().
		SetBackgroundColor(tcell.ColorDarkSlateGray).
		SetTextColor(tcell.ColorWhite).
		SetButtonBackgroundColor(tcell.ColorGray).
		SetButtonTextColor(tcell.ColorWhite)

	return &Modal{
		view: modal,
	}
}

func (m *Modal) View() *tview.Modal {
	return m.view
}

func (m *Modal) Generate(text string, confirmFunc func(), cancelFunc func()) *tview.Modal {
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
