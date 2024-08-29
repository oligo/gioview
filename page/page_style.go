package page

import (
	"image"
	"image/color"
	"reflect"

	"github.com/oligo/gioview/theme"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

var (
	// white
	defaultBackgroundColor = color.NRGBA{}
	maxWidth               = unit.Dp(760)
)

type PageStyle struct {
	Background color.NRGBA
	// Minimun padding of the left and right side.
	Padding   unit.Dp
	MaxWidth  unit.Dp
	listState *widget.List
}

func (p *PageStyle) Layout(gtx C, th *theme.Theme, items ...layout.Widget) D {
	if reflect.ValueOf(p.MaxWidth).IsZero() {
		p.MaxWidth = unit.Dp(maxWidth)
	}

	if reflect.ValueOf(p.Background).IsZero() {
		p.Background = defaultBackgroundColor
	}

	if p.listState == nil {
		p.listState = &widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		}
	}

	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	rect := clip.Rect(image.Rectangle{Max: gtx.Constraints.Max})
	paint.FillShape(gtx.Ops, p.Background, rect.Op())

	width := gtx.Constraints.Max.X
	padding := p.Padding
	if width-2*gtx.Dp(padding) > gtx.Dp(p.MaxWidth) {
		paddingVal := float32(width-gtx.Dp(p.MaxWidth)) / 2.0
		padding = unit.Dp(paddingVal / gtx.Metric.PxPerDp)
	}

	//log.Printf("width: %v, padding: %v, content width: %v", width, padding, width-2*gtx.Dp(padding))

	return material.List(th.Theme, p.listState).Layout(gtx, len(items), func(gtx C, index int) D {
		return layout.Inset{
			Left:  padding,
			Right: padding,
		}.Layout(gtx, items[index])
	})

}
