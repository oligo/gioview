package widget

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/oligo/gioview/theme"
)

type TagVariant uint8

const (
	Solid TagVariant = iota
	Outline
)

// Tag is used for items that need to be labeled using keywords
// that describe them.
type Tag struct {
	TextColor  color.NRGBA
	Background color.NRGBA
	Radius     unit.Dp
	Inset      layout.Inset
	Variant    TagVariant
}

func (t Tag) Layout(gtx C, th *theme.Theme, textSize unit.Sp, txt string) D {
	textColorMacro := op.Record(gtx.Ops)
	paint.ColorOp{Color: t.TextColor}.Add(gtx.Ops)
	textColor := textColorMacro.Stop()

	textFont := font.Font{
		Typeface: th.Face,
		Weight:   font.Normal,
	}

	switch t.Variant {
	case Outline:
		return widget.Border{
			Color:        t.TextColor,
			CornerRadius: t.Radius,
			Width:        gtx.Metric.PxToDp(1),
		}.Layout(gtx, func(gtx C) D {
			return t.Inset.Layout(gtx, func(gtx C) D {
				return t.layoutText(gtx, th.Shaper, textFont, th.TextSize, textColor, txt)
			})
		})

	case Solid:
		return t.layoutSolid(gtx, th.Shaper, textFont, th.TextSize, textColor, txt)
	}

	return D{}
}

func (t Tag) layoutText(gtx C, shaper *text.Shaper, font font.Font, size unit.Sp, textMaterial op.CallOp, txt string) D {

	tl := widget.Label{
		Alignment: text.Start,
		MaxLines:  1,
	}

	return tl.Layout(gtx, shaper, font, size, txt, textMaterial)
}

func (t Tag) layoutSolid(gtx C, shaper *text.Shaper, font font.Font, size unit.Sp, textMaterial op.CallOp, txt string) D {
	macro := op.Record(gtx.Ops)
	dims := t.Inset.Layout(gtx, func(gtx C) D {
		return t.layoutText(gtx, shaper, font, size, textMaterial, txt)
	})
	callOps := macro.Stop()

	defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, int(t.Radius)).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, t.Background)
	callOps.Add(gtx.Ops)

	return dims
}
