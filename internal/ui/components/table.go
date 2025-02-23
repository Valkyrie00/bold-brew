package components

import (
	"bbrew/internal/ui/theme"
	"github.com/rivo/tview"
)

type Table struct {
	view  *tview.Table
	theme *theme.ThemeService
}

func NewTable(theme *theme.ThemeService) *Table {
	table := &Table{
		view:  tview.NewTable(),
		theme: theme,
	}
	table.view.SetBorders(false)
	table.view.SetSelectable(true, false)
	table.view.SetFixed(1, 0)
	return table
}

func (t *Table) SetSelectionHandler(handler func(row, column int)) {
	t.view.SetSelectionChangedFunc(handler)
}

func (t *Table) View() *tview.Table {
	return t.view
}
