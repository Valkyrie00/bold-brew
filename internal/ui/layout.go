package ui

import (
	"bbrew/internal/ui/components"
	"bbrew/internal/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type LayoutInterface interface {
	Setup()
	Root() tview.Primitive

	GetHeader() *components.Header
	GetSearch() *components.Search
	GetTable() *components.Table
	GetDetails() *components.Details
	GetOutput() *components.Output
	GetLegend() *components.Legend
	GetNotifier() *components.Notifier

	//UpdateHeader(appName, version, brewVersion string)
	UpdateDetails(text string)
	UpdateSearchCounter(total, filtered int)

	ShowSuccessNotification(message string)
	ShowWarningNotification(message string)
	ShowErrorNotification(message string)
	ClearNotification()

	SetSearchHandlers(doneFunc func(key tcell.Key), changedFunc func(text string))
	SetTableSelectionHandler(handler func(row, column int))

	GenerateModal(text string, confirmFunc func(), cancelFunc func()) *tview.Modal
}

type Layout struct {
	mainContent *tview.Grid
	header      *components.Header
	search      *components.Search
	table       *components.Table
	details     *components.Details
	output      *components.Output
	legend      *components.Legend
	notifier    *components.Notifier
	modal       *components.Modal
	theme       *theme.ThemeService
}

func NewLayout(theme *theme.ThemeService) *Layout {
	l := &Layout{
		mainContent: tview.NewGrid(),
		header:      components.NewHeader(theme),
		search:      components.NewSearch(theme),
		table:       components.NewTable(theme),
		details:     components.NewDetails(theme),
		output:      components.NewOutput(),
		legend:      components.NewLegend(),
		notifier:    components.NewNotifier(),
		modal:       components.NewModal(),
		theme:       theme,
	}
	l.setupLayout()
	return l
}

func (l *Layout) setupLayout() {
	// Header
	headerContent := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(l.header.View(), 0, 1, false).
		AddItem(l.notifier.View(), 0, 1, false)

	// Search and filters
	searchRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(l.search.Field(), 0, 1, false).
		AddItem(l.search.Counter(), 0, 1, false)

	filtersArea := tview.NewFrame(searchRow).
		SetBorders(0, 0, 0, 0, 3, 3)

	tableFrame := tview.NewFrame(l.table.View()).
		SetBorders(0, 0, 0, 0, 3, 3)

	// Left column with search and table
	leftColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(filtersArea, 2, 0, false).
		AddItem(tableFrame, 0, 4, false)

	// Right column with details and output
	rightColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(l.details.View(), 0, 2, false).
		AddItem(l.output.View(), 0, 1, false)

	// Central content
	mainContent := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftColumn, 0, 2, false).
		AddItem(rightColumn, 0, 1, false)

	// Footer
	footerContent := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(l.legend.View(), 0, 1, false)

	// Final layout
	l.mainContent.
		SetRows(1, 0, 1).
		SetColumns(0).
		SetBorders(true).
		AddItem(headerContent, 0, 0, 1, 1, 0, 0, false).
		AddItem(mainContent, 1, 0, 1, 1, 0, 0, true).
		AddItem(footerContent, 2, 0, 1, 1, 0, 0, false)
}

func (l *Layout) Root() tview.Primitive {
	return l.mainContent
}

func (l *Layout) GetHeader() *components.Header     { return l.header }
func (l *Layout) GetSearch() *components.Search     { return l.search }
func (l *Layout) GetTable() *components.Table       { return l.table }
func (l *Layout) GetDetails() *components.Details   { return l.details }
func (l *Layout) GetOutput() *components.Output     { return l.output }
func (l *Layout) GetLegend() *components.Legend     { return l.legend }
func (l *Layout) GetNotifier() *components.Notifier { return l.notifier }

func (l *Layout) GenerateModal(text string, confirmFunc func(), cancelFunc func()) *tview.Modal {
	return l.modal.Generate(text, confirmFunc, cancelFunc)
}

func (l *Layout) Setup() {
	l.setupLayout()
}

//func (l *Layout) UpdateHeader(appName, version, brewVersion string) {
//	l.header.Update(appName, version, brewVersion)
//}

func (l *Layout) UpdateDetails(text string) {
	l.details.SetContent(text)
}

func (l *Layout) UpdateSearchCounter(total, filtered int) {
	l.search.UpdateCounter(total, filtered)
}

func (l *Layout) ShowSuccessNotification(message string) {
	l.notifier.ShowSuccess(message)
}

func (l *Layout) ShowWarningNotification(message string) {
	l.notifier.ShowWarning(message)
}

func (l *Layout) ShowErrorNotification(message string) {
	l.notifier.ShowError(message)
}

func (l *Layout) ClearNotification() {
	l.notifier.Clear()
}

func (l *Layout) SetSearchHandlers(doneFunc func(key tcell.Key), changedFunc func(text string)) {
	l.search.Field().SetDoneFunc(doneFunc)
	l.search.Field().SetChangedFunc(changedFunc)
}

func (l *Layout) SetTableSelectionHandler(handler func(row, column int)) {
	l.table.View().SetSelectionChangedFunc(handler)
}
