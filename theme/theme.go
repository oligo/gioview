package theme

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/text"
	"gioui.org/unit"
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

type EditorStyle struct {
	// typeface for editing
	TypeFace        font.Typeface
	TextSize        unit.Sp
	Weight          font.Weight
	LineHeight      unit.Sp
	LineHeightScale float32
	//May be helpful for code syntax highlighting.
	ColorScheme string
}

type Theme struct {
	*material.Theme
	AlphaPalette
	Editor      EditorStyle
	NaviBgColor color.NRGBA
}

// NewTheme instantiates a theme, extending material theme.
func NewTheme(fontDir string, noSystemFonts bool) *Theme {
	th := material.NewTheme()

	var options = []text.ShaperOption{
		text.WithCollection(LoadBuiltin(fontDir)),
	}

	if noSystemFonts {
		options = append(options, text.NoSystemFonts())
	}

	th.Shaper = text.NewShaper(options...)

	theme := &Theme{
		Theme:        th,
		AlphaPalette: DefaultAlphaPalette,
	}

	// set defaults
	theme.Editor = EditorStyle{
		TypeFace:        th.Face,
		TextSize:        th.TextSize,
		LineHeightScale: 1.5,
		Weight:          font.Medium,
	}

	return theme
}

func (th *Theme) SetEditorStyle(style EditorStyle) {
	th.Editor = style
}
