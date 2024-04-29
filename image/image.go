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
	defaultRadius   = 0
	defaultFit      = widget.Cover
	defaultPosition = layout.Center
)

// ImageStyle is a widget displaying an image from an ImageSource.
// Styling parameters can be set after construction. Displayed size
// is specified by the max constraints of the widget.
type ImageStyle struct {
	src *ImageSource
	//Size   image.Point
	Radius   unit.Dp
	Fit      widget.Fit
	Position layout.Direction
}

func NewImage(src *ImageSource) *ImageStyle {
	return &ImageStyle{
		src:      src,
		Radius:   unit.Dp(defaultRadius),
		Fit:      defaultFit,
		Position: defaultPosition,
	}
}

func (img *ImageStyle) Layout(gtx layout.Context) layout.Dimensions {
	size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Max.Y}

	defer clip.UniformRRect(image.Rectangle{Max: size}, gtx.Dp(img.Radius)).Push(gtx.Ops).Pop()

	if img.src == nil {
		return img.layoutEmptyImg(gtx)
	}

	return img.layoutImg(gtx)
}

func (img *ImageStyle) layoutImg(gtx layout.Context) layout.Dimensions {
	imgOp, err := img.src.ImageOp(gtx.Constraints.Max)
	if err != nil {
		return img.layoutEmptyImg(gtx)
	}

	_img := widget.Image{
		Src:      *imgOp,
		Scale:    1.0 / gtx.Metric.PxPerDp,
		Fit:      img.Fit,
		Position: img.Position,
	}

	gtx.Constraints.Max = imgOp.Size()
	return _img.Layout(gtx)
}

func (img *ImageStyle) layoutEmptyImg(gtx layout.Context) layout.Dimensions {
	src := image.NewUniform(color.Black)
	_img := widget.Image{Src: paint.NewImageOp(src)}
	_img.Scale = 1.0 / gtx.Metric.PxPerDp
	return _img.Layout(gtx)
}
