package services

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type LayoutServiceInterface interface {
	GetGrid() *tview.Grid
	SetGrid()

	GetHeaderView() *tview.TextView
	SetHeaderView(name, version, brewVersion string)

	GetLegendView() *tview.TextView
	SetLegendView()

	GetResultTable() *tview.Table
	SetResultTable(selectionChanged func(row, column int))

	GetDetailsView() *tview.TextView
	SetDetailsView()

	GetOutputView() *tview.TextView
	SetBuildOutputView()

	GetSearchField() *tview.InputField
	SetSearchField(done func(key tcell.Key), changed func(text string))

	GetFilterCounterView() *tview.TextView
	SetFilterCounterView()
	UpdateFilterCounterView(total, filtered int)

	GenerateModal(text string, confirmFunc func(), cancelFunc func()) *tview.Modal
}

type LayoutService struct {
	header        *tview.TextView
	legend        *tview.TextView
	table         *tview.Table
	detailsView   *tview.TextView
	outputView    *tview.TextView
	searchField   *tview.InputField
	filterCounter *tview.TextView
	grid          *tview.Grid
}

var NewLayoutService = func() LayoutServiceInterface {
	return &LayoutService{}
}

func (s *LayoutService) GetGrid() *tview.Grid {
	return s.grid
}

func (s *LayoutService) SetGrid() {
	searchRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(s.searchField, 0, 1, false).
		AddItem(s.filterCounter, 0, 1, false)

	filtersArea := tview.NewFrame(searchRow).
		SetBorders(0, 0, 0, 0, 3, 3)

	tableFrame := tview.NewFrame(s.table).
		SetBorders(0, 0, 0, 0, 3, 3)

	leftColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(filtersArea, 2, 0, false). // Fixed height of 3 rows
		AddItem(tableFrame, 0, 4, false)

	rightColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(s.detailsView, 0, 2, false).
		AddItem(s.outputView, 0, 1, false)

	mainContent := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftColumn, 0, 1, false).
		AddItem(rightColumn, 0, 1, false)

	s.grid = tview.NewGrid().
		SetRows(1, 0, 1).
		SetColumns(0).
		SetBorders(true).
		AddItem(s.header, 0, 0, 1, 1, 0, 0, false).
		AddItem(mainContent, 1, 0, 1, 1, 0, 0, true).
		AddItem(s.legend, 2, 0, 1, 1, 0, 0, false)
}

func (s *LayoutService) GetHeaderView() *tview.TextView {
	return s.header
}

func (s *LayoutService) SetHeaderView(name, version, brewVersion string) {
	s.header = tview.NewTextView().
		SetText(fmt.Sprintf("%s %s - %s", name, version, brewVersion)).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
}

func (s *LayoutService) GetLegendView() *tview.TextView {
	return s.legend
}

func (s *LayoutService) SetLegendView() {
	s.legend = tview.NewTextView().
		SetText(tview.Escape("[/] Search | [f] Filter Installed | [i] Install | [u] Update | [r] Remove | [Esc] Back to Table | [ctrl+u] Update Homebrew | [q] Quit")).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
}

func (s *LayoutService) GetResultTable() *tview.Table {
	return s.table
}

func (s *LayoutService) SetResultTable(selectionChanged func(row, column int)) {
	s.table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectionChangedFunc(selectionChanged)
}

func (s *LayoutService) GetDetailsView() *tview.TextView {
	return s.detailsView
}

func (s *LayoutService) SetDetailsView() {
	s.detailsView = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft)
	s.detailsView.SetTitle("Details").SetTitleColor(tcell.ColorYellowGreen).SetTitleAlign(tview.AlignLeft).SetBorder(true)
}

func (s *LayoutService) GetOutputView() *tview.TextView {
	return s.outputView
}

func (s *LayoutService) SetBuildOutputView() {
	s.outputView = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft)
	s.outputView.SetBorder(true).SetTitle("Output").SetTitleColor(tcell.ColorYellowGreen).SetTitleAlign(tview.AlignLeft)
}

func (s *LayoutService) GetSearchField() *tview.InputField {
	return s.searchField
}

func (s *LayoutService) SetSearchField(done func(key tcell.Key), changed func(text string)) {
	s.searchField = tview.NewInputField().
		SetLabel("Search (All): ").
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorYellow).
		SetFieldWidth(30).
		SetDoneFunc(done).
		SetChangedFunc(changed)
}

func (s *LayoutService) GetFilterCounterView() *tview.TextView {
	return s.filterCounter
}

func (s *LayoutService) SetFilterCounterView() {
	s.filterCounter = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight)
}

func (s *LayoutService) UpdateFilterCounterView(total, filtered int) {
	s.filterCounter.SetText(fmt.Sprintf("Total: %d | Filtered: %d", total, filtered))
}

func (s *LayoutService) GenerateModal(text string, confirmFunc func(), cancelFunc func()) *tview.Modal {
	return tview.NewModal().
		SetText(text).
		AddButtons([]string{"Confirm", "Cancel"}).
		SetBackgroundColor(tcell.ColorDarkSlateGray).
		SetTextColor(tcell.ColorWhite).
		SetButtonBackgroundColor(tcell.ColorGray).
		SetButtonTextColor(tcell.ColorWhite).
		SetDoneFunc(func(_ int, buttonLabel string) {
			if buttonLabel == "Confirm" {
				confirmFunc()
			} else if buttonLabel == "Cancel" {
				cancelFunc()
			}
		})
}
