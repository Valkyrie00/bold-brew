package components

import (
	"bbrew/internal/ui/theme"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Search struct {
	field   *tview.InputField
	counter *tview.TextView
	theme   *theme.Theme
}

func NewSearch(theme *theme.Theme) *Search {
	search := &Search{
		field:   tview.NewInputField(),
		counter: tview.NewTextView(),
		theme:   theme,
	}

	search.field.SetLabel("Search (All): ")
	search.field.SetFieldBackgroundColor(theme.DefaultBgColor)
	search.field.SetFieldTextColor(theme.DefaultTextColor)
	search.field.SetLabelColor(theme.LabelColor)
	search.field.SetFieldWidth(30)
	search.counter.SetDynamicColors(true)
	search.counter.SetTextAlign(tview.AlignRight)
	return search
}

func (s *Search) SetHandlers(done func(key tcell.Key), changed func(text string)) {
	s.field.SetDoneFunc(done)
	s.field.SetChangedFunc(changed)
}

func (s *Search) UpdateCounter(total, filtered int) {
	s.counter.SetText(fmt.Sprintf("Total: %d | Filtered: %d", total, filtered))
}

func (s *Search) Field() *tview.InputField {
	return s.field
}

func (s *Search) Counter() *tview.TextView {
	return s.counter
}
