package theme

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Theme struct {
	// Application-specific colors
	DefaultTextColor tcell.Color
	DefaultBgColor   tcell.Color
	WarningColor     tcell.Color
	SuccessColor     tcell.Color
	ErrorColor       tcell.Color

	TitleColor      tcell.Color
	LabelColor      tcell.Color
	ButtonBgColor   tcell.Color
	ButtonTextColor tcell.Color

	ModalBgColor     tcell.Color
	LegendColor      tcell.Color
	TableHeaderColor tcell.Color
	SearchLabelColor       tcell.Color
	SearchBorderColor      tcell.Color
	SearchFocusBorderColor tcell.Color

	// tview global styles (mapped to tview.Styles)
	PrimitiveBackgroundColor    tcell.Color
	ContrastBackgroundColor     tcell.Color
	MoreContrastBackgroundColor tcell.Color
	BorderColor                 tcell.Color
	GraphicsColor               tcell.Color
	PrimaryTextColor            tcell.Color
	SecondaryTextColor          tcell.Color
	TertiaryTextColor           tcell.Color
	InverseTextColor            tcell.Color
	ContrastSecondaryTextColor  tcell.Color
}

func NewTheme() *Theme {
	theme := &Theme{
		// Application-specific colors
		DefaultTextColor: tcell.ColorDefault,
		DefaultBgColor:   tcell.ColorDefault,

		// Use standard ANSI colors that work well on both light and dark themes
		WarningColor: tcell.ColorYellow,
		SuccessColor: tcell.ColorGreen,
		ErrorColor:   tcell.ColorRed,

		// Component colors
		TitleColor:      tcell.ColorPurple,
		LabelColor:      tcell.ColorYellow,
		ButtonBgColor:   tcell.ColorDefault,
		ButtonTextColor: tcell.ColorDefault,

		ModalBgColor:     tcell.ColorDefault,
		LegendColor:      tcell.ColorDefault,
		TableHeaderColor: tcell.ColorBlue,
		SearchLabelColor:       tcell.ColorPurple,
		SearchBorderColor:      tcell.ColorWhite,
		SearchFocusBorderColor: tcell.ColorGreen,

		// tview global styles - use terminal default colors for better compatibility
		// By default, tview uses hardcoded colors (like tcell.ColorBlack) which don't
		// adapt to the terminal's theme. We set them all to ColorDefault.
		PrimitiveBackgroundColor:    tcell.ColorDefault,
		ContrastBackgroundColor:     tcell.ColorDefault,
		MoreContrastBackgroundColor: tcell.ColorDefault,
		BorderColor:                 tcell.ColorDefault,
		GraphicsColor:               tcell.ColorDefault,
		PrimaryTextColor:            tcell.ColorDefault,
		SecondaryTextColor:          tcell.ColorDefault,
		TertiaryTextColor:           tcell.ColorDefault,
		InverseTextColor:            tcell.ColorDefault,
		ContrastSecondaryTextColor:  tcell.ColorDefault,
	}

	// Apply theme to tview global styles
	tview.Styles.PrimitiveBackgroundColor = theme.PrimitiveBackgroundColor
	tview.Styles.ContrastBackgroundColor = theme.ContrastBackgroundColor
	tview.Styles.MoreContrastBackgroundColor = theme.MoreContrastBackgroundColor
	tview.Styles.BorderColor = theme.BorderColor
	tview.Styles.TitleColor = theme.TitleColor
	tview.Styles.GraphicsColor = theme.GraphicsColor
	tview.Styles.PrimaryTextColor = theme.PrimaryTextColor
	tview.Styles.SecondaryTextColor = theme.SecondaryTextColor
	tview.Styles.TertiaryTextColor = theme.TertiaryTextColor
	tview.Styles.InverseTextColor = theme.InverseTextColor
	tview.Styles.ContrastSecondaryTextColor = theme.ContrastSecondaryTextColor

	return theme
}
