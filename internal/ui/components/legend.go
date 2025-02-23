package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Legend struct {
	view *tview.TextView
}

func NewLegend() *Legend {
	legendText := "[yellow]" + tview.Escape(
		"[/] Search | "+
			"[f] Filter Installed | "+
			"[i] Install | "+
			"[u] Update | "+
			"[ctrl+u] Update All | "+
			"[r] Remove | "+
			"[Esc] Back to Table | "+
			"[q] Quit",
	) + "[-]"

	legendView := tview.NewTextView().
		SetText(legendText).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorWhite)

	return &Legend{
		view: legendView,
	}
}

func (l *Legend) View() *tview.TextView {
	return l.view
}

func (l *Legend) SetText(text string) {
	l.view.SetText(text)
}

func (l *Legend) Clear() {
	l.view.Clear()
}
