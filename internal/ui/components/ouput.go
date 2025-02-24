package components

import (
	"bbrew/internal/ui/theme"
	"github.com/rivo/tview"
)

type Output struct {
	view  *tview.TextView
	theme *theme.Theme
}

func NewOutput(theme *theme.Theme) *Output {
	output := &Output{
		view:  tview.NewTextView(),
		theme: theme,
	}

	output.view.SetDynamicColors(true)
	output.view.SetScrollable(true)
	output.view.SetWrap(true)
	output.view.SetTextAlign(tview.AlignLeft)
	output.view.SetBorder(true)
	output.view.SetTitle("Output")
	output.view.SetTitleColor(theme.TitleColor)
	output.view.SetTitleAlign(tview.AlignLeft)
	return output
}

func (o *Output) View() *tview.TextView {
	return o.view
}

func (o *Output) Clear() {
	o.view.Clear()
}

func (o *Output) Write(text string) {
	o.view.SetText(text)
}

func (o *Output) Append(text string) {
	currentText := o.view.GetText(true)
	o.view.SetText(currentText + text)
}

func (o *Output) ScrollToEnd() {
	o.view.ScrollToEnd()
}
