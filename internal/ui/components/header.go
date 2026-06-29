package components

import (
	"fmt"

	"github.com/rivo/tview"

	"bbrew/internal/ui/theme"
)

type Header struct {
	view  *tview.TextView
	theme *theme.Theme
}

func NewHeader(theme *theme.Theme) *Header {
	header := &Header{
		view:  tview.NewTextView(),
		theme: theme,
	}

	header.view.SetDynamicColors(true)
	header.view.SetTextAlign(tview.AlignLeft)
	return header
}

func (h *Header) Update(name, version, brewVersion string) {
	h.view.SetText(fmt.Sprintf(" %s %s - %s", name, version, brewVersion))
}

func (h *Header) View() *tview.TextView {
	return h.view
}
