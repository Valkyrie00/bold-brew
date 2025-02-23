package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Output struct {
	view *tview.TextView
}

func NewOutput() *Output {
	outputView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true).
		SetTextAlign(tview.AlignLeft)

	outputView.SetBorder(true).
		SetTitle("Output").
		SetTitleColor(tcell.ColorYellowGreen).
		SetTitleAlign(tview.AlignLeft)

	return &Output{
		view: outputView,
	}
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
