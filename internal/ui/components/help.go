package components

import (
	"bbrew/internal/ui/theme"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// HelpScreen displays a modal overlay with all keyboard shortcuts
type HelpScreen struct {
	pages      *tview.Pages
	theme      *theme.Theme
	isBrewfile bool
}

// NewHelpScreen creates a new help screen component
func NewHelpScreen(theme *theme.Theme) *HelpScreen {
	return &HelpScreen{
		theme: theme,
	}
}

// View returns the help screen pages (for overlay functionality)
func (h *HelpScreen) View() *tview.Pages {
	return h.pages
}

// SetBrewfileMode sets whether Brewfile-specific commands should be shown
func (h *HelpScreen) SetBrewfileMode(enabled bool) {
	h.isBrewfile = enabled
}

// Build creates the help screen as an overlay on top of the main content
func (h *HelpScreen) Build(mainContent tview.Primitive) *tview.Pages {
	content := h.buildHelpContent()

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetText(content).
		SetTextAlign(tview.AlignLeft)

	textView.SetBackgroundColor(h.theme.ModalBgColor)
	textView.SetTextColor(h.theme.DefaultTextColor)

	// Create a frame around the text
	frame := tview.NewFrame(textView).
		SetBorders(1, 1, 1, 1, 2, 2)
	frame.SetBackgroundColor(h.theme.ModalBgColor)
	frame.SetBorderColor(h.theme.BorderColor)
	frame.SetBorder(true).
		SetTitle(" Help ").
		SetTitleAlign(tview.AlignCenter)

	// Calculate box dimensions
	boxHeight := 22
	boxWidth := 55
	if h.isBrewfile {
		boxHeight = 26 // Extra space for Brewfile section
	}

	// Center the frame in a flex layout
	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(frame, boxHeight, 0, true).
			AddItem(nil, 0, 1, false),
			boxWidth, 0, true).
		AddItem(nil, 0, 1, false)

	// Create pages with main content as background and help as overlay
	h.pages = tview.NewPages().
		AddPage("main", mainContent, true, true).
		AddPage("help", centered, true, true)

	return h.pages
}

// buildHelpContent generates the formatted help text
func (h *HelpScreen) buildHelpContent() string {
	var sb strings.Builder

	// Navigation section
	sb.WriteString(h.formatSection("NAVIGATION"))
	sb.WriteString(h.formatKey("↑/↓, j/k", "Navigate list"))
	sb.WriteString(h.formatKey("/", "Focus search"))
	sb.WriteString(h.formatKey("Shift+T", "Sort by Type"))
	sb.WriteString(h.formatKey("Esc", "Back to table"))
	sb.WriteString(h.formatKey("q", "Quit"))
	sb.WriteString("\n")

	// Filters section
	sb.WriteString(h.formatSection("FILTERS"))
	sb.WriteString(h.formatKey("Shift+F", "Toggle installed"))
	sb.WriteString(h.formatKey("Shift+O", "Toggle outdated"))
	sb.WriteString(h.formatKey("Shift+L", "Toggle leaves"))
	sb.WriteString(h.formatKey("Shift+C", "Toggle casks"))
	sb.WriteString("\n")

	// Actions section
	sb.WriteString(h.formatSection("ACTIONS"))
	sb.WriteString(h.formatKey("o", "Open Homepage"))
	sb.WriteString(h.formatKey("i", "Install selected"))
	sb.WriteString(h.formatKey("u", "Update selected"))
	sb.WriteString(h.formatKey("r", "Remove selected"))
	sb.WriteString(h.formatKey("Ctrl+U", "Update all"))

	// Brewfile section (only if in Brewfile mode)
	if h.isBrewfile {
		sb.WriteString("\n")
		sb.WriteString(h.formatSection("BREWFILE"))
		sb.WriteString(h.formatKey("Ctrl+A", "Install all"))
		sb.WriteString(h.formatKey("Ctrl+R", "Remove all"))
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("[%s]Press any key to close[-]", h.getColorTag(h.theme.LegendColor)))

	return sb.String()
}

// formatSection formats a section header
func (h *HelpScreen) formatSection(title string) string {
	return fmt.Sprintf("[%s::b]%s[-:-:-]\n", h.getColorTag(h.theme.SuccessColor), title)
}

// formatKey formats a key-description pair
func (h *HelpScreen) formatKey(key, description string) string {
	return fmt.Sprintf("  [%s]%-12s[-] %s\n", h.getColorTag(h.theme.WarningColor), key, description)
}

// getColorTag converts a tcell.Color to a tview color tag
func (h *HelpScreen) getColorTag(color tcell.Color) string {
	return fmt.Sprintf("#%06x", color.Hex())
}
