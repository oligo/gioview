package image

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

const (
	MaxSize       = 60
	defaultRadius = 10
)

// Avatar is a widget displaying a square image.
// If radius is specified, the image is cropped with an outer rounded rectangle.
type GioImg struct {
	src   *ImageSource
	imgOp paint.ImageOp
	//Size   image.Point
	Radius unit.Dp
}

func NewGioImg(src *ImageSource) *GioImg {
	return &GioImg{
		src:    src,
		Radius: unit.Dp(defaultRadius),
	}
}

func (gi *GioImg) Layout(gtx layout.Context) layout.Dimensions {
	size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Max.Y}

	if gi.imgOp != (paint.ImageOp{}) {
		size = gi.imgOp.Size()
	}

	defer clip.UniformRRect(image.Rectangle{Max: size}, gtx.Dp(gi.Radius)).Push(gtx.Ops).Pop()

	if gi.src == nil {
		return gi.layoutEmptyImg(gtx)
	}

	return gi.layoutImg(gtx)
}

func (gi *GioImg) layoutImg(gtx layout.Context) layout.Dimensions {
	if gi.imgOp == (paint.ImageOp{}) {
		imgOp, err := gi.src.ImageOp(gtx.Constraints.Max)
		if err != nil {
			return gi.layoutEmptyImg(gtx)
		}
		gi.imgOp = imgOp
	}

	img := widget.Image{
		Src:      gi.imgOp,
		Scale:    1.0 / gtx.Metric.PxPerDp,
		Fit:      widget.Cover,
		Position: layout.Center,
	}

	gtx.Constraints.Max = gi.imgOp.Size()
	return img.Layout(gtx)
}

func (gi *GioImg) layoutEmptyImg(gtx layout.Context) layout.Dimensions {
	src := image.NewUniform(color.Black)
	img := widget.Image{Src: paint.NewImageOp(src)}
	img.Scale = 1.0 / gtx.Metric.PxPerDp
	return img.Layout(gtx)
}
