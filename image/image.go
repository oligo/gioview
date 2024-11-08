package image

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// ImageStyle is a widget displaying an image from an ImageSource.
// Styling parameters can be set after construction. Displayed size
// is specified by the max constraints of the widget.
type ImageStyle struct {
	Src *ImageSource
	//Size   image.Point
	Radius   unit.Dp
	Fit      widget.Fit
	Scale    float32
	Position layout.Direction
}

func (img ImageStyle) Layout(gtx layout.Context) layout.Dimensions {
	if img.Src != nil && img.Src.onLoadedRedraw == nil {
		img.Src.onLoadedRedraw = func() {
			gtx.Execute(op.InvalidateCmd{})
		}
	}

	if img.Scale <= 0 {
		img.Scale = 1.0 / gtx.Metric.PxPerDp
	}

	macro := op.Record(gtx.Ops)
	dims := func() layout.Dimensions {
		if img.Src == nil {
			return img.layoutEmptyImg(gtx)
		}

		dims := img.layoutImg(gtx)
		return dims
	}()
	call := macro.Stop()

	defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, gtx.Dp(img.Radius)).Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)

	return dims
}

func (img ImageStyle) layoutImg(gtx layout.Context) layout.Dimensions {
	imgOp := img.Src.ImageOp(gtx.Constraints.Max)
	_img := widget.Image{
		Src:      *imgOp,
		Scale:    img.Scale,
		Fit:      img.Fit,
		Position: img.Position,
	}

	return _img.Layout(gtx)
}

func (img ImageStyle) layoutEmptyImg(gtx layout.Context) layout.Dimensions {
	src := image.NewUniform(color.Gray{Y: 232})
	_img := widget.Image{
		Src:      paint.NewImageOp(src),
		Fit:      widget.Unscaled,
		Scale:    1.0 / gtx.Metric.PxPerDp,
		Position: layout.Center,
	}
	return _img.Layout(gtx)
}
