package theme

import (
	"gioui.org/text"
	"gioui.org/widget/material"
)

// AlphaPalette is the set of alpha values to be applied for certain
// material design states like hover, selected, etc...
type AlphaPalette struct {
	Hover, Selected uint8
}

var DefaultAlphaPalette = AlphaPalette{
	Hover:    48,
	Selected: 96,
}

type Theme struct {
	*material.Theme
	AlphaPalette
}

// NewTheme instantiates a theme, extending material theme.
func NewTheme(fontDir string, embeddedFonts [][]byte, noSystemFonts bool) *Theme {
	th := material.NewTheme()

	var options = []text.ShaperOption{
		text.WithCollection(LoadBuiltin(fontDir, embeddedFonts)),
	}

	if noSystemFonts {
		options = append(options, text.NoSystemFonts())
	}

	th.Shaper = text.NewShaper(options...)

	theme := &Theme{
		Theme:        th,
		AlphaPalette: DefaultAlphaPalette,
	}

	return theme
}
