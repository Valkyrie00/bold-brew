package theme

import "github.com/gdamore/tcell/v2"

type Theme struct {
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
	SearchLabelColor tcell.Color
}

func NewTheme() *Theme {
	return &Theme{
		DefaultTextColor: tcell.ColorWhite,
		DefaultBgColor:   tcell.ColorBlack,
		WarningColor:     tcell.ColorYellow,
		SuccessColor:     tcell.ColorGreen,
		ErrorColor:       tcell.ColorRed,

		TitleColor:      tcell.ColorMediumVioletRed,
		LabelColor:      tcell.ColorYellow,
		ButtonBgColor:   tcell.ColorGray,
		ButtonTextColor: tcell.ColorWhite,

		ModalBgColor:     tcell.ColorDarkSlateGray,
		LegendColor:      tcell.ColorWhite,
		TableHeaderColor: tcell.ColorBlue,
		SearchLabelColor: tcell.ColorMediumVioletRed,
	}
}
