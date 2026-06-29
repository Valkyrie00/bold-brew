package ui

import (
	"github.com/rivo/tview"
)

// ThreadSafeWriter wraps a tview.TextView to implement io.Writer with
// thread-safe UI updates via QueueUpdateDraw. This allows services to
// write output without knowing about the UI framework.
type ThreadSafeWriter struct {
	app  *tview.Application
	view *tview.TextView
}

// NewThreadSafeWriter creates a writer that safely updates a tview.TextView from any goroutine.
func NewThreadSafeWriter(app *tview.Application, view *tview.TextView) *ThreadSafeWriter {
	return &ThreadSafeWriter{app: app, view: view}
}

func (w *ThreadSafeWriter) Write(p []byte) (n int, err error) {
	data := make([]byte, len(p))
	copy(data, p)
	w.app.QueueUpdateDraw(func() {
		_, _ = w.view.Write(data)
		w.view.ScrollToEnd()
	})
	return len(p), nil
}
