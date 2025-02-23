package components

import (
	"bbrew/internal/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Details struct {
	view  *tview.TextView
	theme *theme.ThemeService
}

func NewDetails(theme *theme.ThemeService) *Details {
	details := &Details{
		view:  tview.NewTextView(),
		theme: theme,
	}
	details.view.SetDynamicColors(true)
	details.view.SetTextAlign(tview.AlignLeft)
	details.view.SetTitle("Details")
	details.view.SetTitleColor(tcell.ColorYellowGreen)
	details.view.SetTitleAlign(tview.AlignLeft)
	details.view.SetBorder(true)
	return details
}

func (d *Details) SetContent(text string) {
	d.view.SetText(text)
}

func (d *Details) View() *tview.TextView {
	return d.view
}
