package components

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Notifier struct {
	view *tview.TextView
}

func NewNotifier() *Notifier {
	notifierView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight)

	return &Notifier{
		view: notifierView,
	}
}

func (n *Notifier) View() *tview.TextView {
	return n.view
}

func (n *Notifier) ShowSuccess(message string) {
	n.view.SetTextColor(tcell.ColorGreen).SetText(fmt.Sprintf(" %s ", message))
}

func (n *Notifier) ShowWarning(message string) {
	n.view.SetTextColor(tcell.ColorYellow).SetText(fmt.Sprintf(" %s ", message))
}

func (n *Notifier) ShowError(message string) {
	n.view.SetTextColor(tcell.ColorRed).SetText(fmt.Sprintf(" %s ", message))
}

func (n *Notifier) Clear() {
	n.view.Clear()
}
