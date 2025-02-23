package theme

import "github.com/gdamore/tcell/v2"

type ThemeService struct {
	HeaderColor      tcell.Color
	HighlightColor   tcell.Color
	WarningColor     tcell.Color
	SuccessColor     tcell.Color
	ErrorColor       tcell.Color
	DefaultTextColor tcell.Color
}

func NewThemeService() *ThemeService {
	return &ThemeService{
		HeaderColor:      tcell.ColorBlue,
		HighlightColor:   tcell.ColorOrange,
		WarningColor:     tcell.ColorYellow,
		SuccessColor:     tcell.ColorGreen,
		ErrorColor:       tcell.ColorRed,
		DefaultTextColor: tcell.ColorWhite,
	}
}
