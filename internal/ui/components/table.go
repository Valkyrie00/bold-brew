package components

import (
	"bbrew/internal/ui/theme"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Table struct {
	view         *tview.Table
	theme        *theme.Theme
	selectedRows map[int]bool
}

func NewTable(theme *theme.Theme) *Table {
	table := &Table{
		view:         tview.NewTable(),
		theme:        theme,
		selectedRows: make(map[int]bool),
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
	t.selectedRows = make(map[int]bool)
}

func (t *Table) ClearSelection() {
	t.selectedRows = make(map[int]bool)
}

func (t *Table) ToggleSelection(row int, highlightColor tcell.Color) {
	isSelected := false
	if t.selectedRows[row] {
		delete(t.selectedRows, row)
	} else {
		t.selectedRows[row] = true
		isSelected = true
	}

	// Update visual style for the row
	colCount := t.view.GetColumnCount()
	for i := 0; i < colCount; i++ {
		cell := t.view.GetCell(row, i)
		if cell != nil {
			if isSelected {
				cell.SetBackgroundColor(highlightColor)
			} else {
				cell.SetBackgroundColor(t.theme.DefaultBgColor) // Or tcell.ColorDefault
			}
		}
	}
}

func (t *Table) IsSelected(row int) bool {
	return t.selectedRows[row]
}

func (t *Table) GetSelectedRows() []int {
	rows := make([]int, 0, len(t.selectedRows))
	for row := range t.selectedRows {
		rows = append(rows, row)
	}
	return rows
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
