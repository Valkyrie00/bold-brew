package ui

import (
	"bbrew/internal/ui/components"
	"bbrew/internal/ui/theme"

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
	GetModal() *components.Modal
	GetHelpScreen() *components.HelpScreen
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
	helpScreen  *components.HelpScreen
	theme       *theme.Theme
}

func NewLayout(theme *theme.Theme) LayoutInterface {
	return &Layout{
		mainContent: tview.NewGrid(),
		header:      components.NewHeader(theme),
		search:      components.NewSearch(theme),
		table:       components.NewTable(theme),
		details:     components.NewDetails(theme),
		output:      components.NewOutput(theme),
		legend:      components.NewLegend(theme),
		notifier:    components.NewNotifier(theme),
		modal:       components.NewModal(theme),
		helpScreen:  components.NewHelpScreen(theme),
		theme:       theme,
	}
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
		AddItem(filtersArea, 3, 0, false).
		AddItem(tableFrame, 0, 4, false)

	// Right column with details and output
	rightColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(l.details.View(), 0, 2, false).
		AddItem(l.output.View(), 0, 1, false)

	// Central content (left 75%, right 25%)
	mainContent := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftColumn, 0, 3, false).
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

func (l *Layout) Setup() {
	l.setupLayout()
}

func (l *Layout) Root() tview.Primitive {
	return l.mainContent
}

func (l *Layout) GetHeader() *components.Header         { return l.header }
func (l *Layout) GetSearch() *components.Search         { return l.search }
func (l *Layout) GetTable() *components.Table           { return l.table }
func (l *Layout) GetDetails() *components.Details       { return l.details }
func (l *Layout) GetOutput() *components.Output         { return l.output }
func (l *Layout) GetLegend() *components.Legend         { return l.legend }
func (l *Layout) GetNotifier() *components.Notifier     { return l.notifier }
func (l *Layout) GetModal() *components.Modal           { return l.modal }
func (l *Layout) GetHelpScreen() *components.HelpScreen { return l.helpScreen }
