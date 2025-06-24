package components

import (
	"bbrew/internal/ui/theme"
	"fmt"
	"github.com/rivo/tview"
)

type Legend struct {
	view  *tview.TextView
	theme *theme.Theme
}

func NewLegend(theme *theme.Theme) *Legend {
	legendView := tview.NewTextView().
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

func (l *Legend) GetFormattedLabel(keySlug, label string, active bool) string {
	if active {
		return fmt.Sprintf("[yellow::b]%s[-]", tview.Escape(fmt.Sprintf("[%s] %s", keySlug, label)))
	}

	return tview.Escape(fmt.Sprintf("[%s] %s", keySlug, label))
}

func (l *Legend) SetText(text string) {
	l.view.SetText(text)
}

func (l *Legend) Clear() {
	l.view.Clear()
}
