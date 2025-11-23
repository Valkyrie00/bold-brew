package components

import (
	"bbrew/internal/ui/theme"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Table struct {
	view  *tview.Table
	theme *theme.Theme
}

func NewTable(theme *theme.Theme) *Table {
	table := &Table{
		view:  tview.NewTable(),
		theme: theme,
	}
	table.view.SetBorders(false)
	table.view.SetSelectable(true, false)
	table.view.SetFixed(1, 0)

	// Use reverse video for selection to ensure visibility on any terminal theme
	table.view.SetSelectedStyle(tcell.StyleDefault.Reverse(true))

	return table
}

func (t *Table) SetSelectionHandler(handler func(row, column int)) {
	t.view.SetSelectionChangedFunc(handler)
}

func (t *Table) View() *tview.Table {
	return t.view
}

func (t *Table) Clear() {
	t.view.Clear()
}

func (t *Table) SetTableHeaders(headers ...string) {
	for i, header := range headers {
		t.view.SetCell(0, i, &tview.TableCell{
			Text:            header,
			NotSelectable:   true,
			Align:           tview.AlignLeft,
			Color:           t.theme.TableHeaderColor,
			BackgroundColor: t.theme.DefaultBgColor,
		})
	}
}
