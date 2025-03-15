package components

import (
	"bbrew/internal/ui/theme"
	"github.com/rivo/tview"
)

type Legend struct {
	view  *tview.TextView
	theme *theme.Theme
}

func NewLegend(theme *theme.Theme) *Legend {
	legendText := tview.Escape(
		"[/] Search | " +
			"[f] Filter Installed | " +
			"[o] Filter Outdated | " +
			"[i] Install | " +
			"[u] Update | " +
			"[ctrl+u] Update All | " +
			"[r] Remove | " +
			"[Esc] Back to Table | " +
			"[q] Quit",
	)

	legendView := tview.NewTextView().
		SetText(legendText).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(theme.LegendColor)

	return &Legend{
		view:  legendView,
		theme: theme,
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
