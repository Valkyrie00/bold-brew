package components

import (
	"bbrew/internal/ui/theme"
	"fmt"
	"github.com/rivo/tview"
)

type Notifier struct {
	view  *tview.TextView
	theme *theme.Theme
}

func NewNotifier(theme *theme.Theme) *Notifier {
	notifierView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight)

	return &Notifier{
		view:  notifierView,
		theme: theme,
	}
}

func (n *Notifier) View() *tview.TextView {
	return n.view
}

func (n *Notifier) ShowSuccess(message string) {
	n.view.SetTextColor(n.theme.SuccessColor).SetText(fmt.Sprintf(" %s ", message))
}

func (n *Notifier) ShowWarning(message string) {
	n.view.SetTextColor(n.theme.WarningColor).SetText(fmt.Sprintf(" %s ", message))
}

func (n *Notifier) ShowError(message string) {
	n.view.SetTextColor(n.theme.ErrorColor).SetText(fmt.Sprintf(" %s ", message))
}

func (n *Notifier) Clear() {
	n.view.Clear()
}
