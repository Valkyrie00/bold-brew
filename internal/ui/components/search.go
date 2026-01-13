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

	search.field.SetTitle("Search (All)")
	search.field.SetTitleColor(theme.TitleColor)
	search.field.SetTitleAlign(tview.AlignLeft)
	search.field.SetFieldStyle(tcell.StyleDefault.Italic(true).Underline(true))
	search.field.SetFieldBackgroundColor(theme.DefaultBgColor)
	search.field.SetFieldTextColor(theme.DefaultTextColor)
	search.field.SetBorder(true)
	search.field.SetBorderColor(theme.SearchBorderColor)
	search.field.SetFocusFunc(func() {
		search.field.SetBorderColor(theme.SearchFocusBorderColor)
	})
	search.field.SetBlurFunc(func() {
		search.field.SetBorderColor(theme.SearchBorderColor)
	})

	search.counter.SetDynamicColors(true)
	search.counter.SetTextAlign(tview.AlignRight)
	search.counter.SetBorderPadding(1, 0, 0, 0)
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
