package view

import (
	"image"
	"image/color"
	"net/url"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/oligo/gioview/misc"
	"github.com/oligo/gioview/theme"
)

// A link between views, as anchor in HTML.
type Link struct {
	Title  string
	Src    ViewID
	Params map[string]interface{}
	// This is the url for webview only.
	Location  string
	Referer   url.URL
	UseNewTab bool

	click    gesture.Click
	hovering bool
	// whether the link has been clicked ever.
	clicked bool
}

func (link *Link) OnClick() Intent {
	intent := Intent{
		Target:     link.Src,
		Params:     link.Params,
		Referer:    link.Referer,
		RequireNew: link.UseNewTab,
	}

	if link.Location != "" {
		if intent.Params == nil {
			intent.Params = make(map[string]interface{})
		}

		intent.Params["url"] = link.Location
	}

	return intent
}

func (link *Link) Clicked() bool {
	return link.clicked
}

func (link *Link) Layout(gtx C, lt *text.Shaper, font font.Font, size unit.Sp, textMaterial op.CallOp) D {
	link.Update(gtx)

	tl := widget.Label{
		Alignment: text.Start,
		MaxLines:  1,
	}

	// todo: should be displayed as inline text that allows text to flow to the next line.
	return tl.Layout(gtx, lt, font, size, link.Title, textMaterial)
}

// Update handles link events and reports if the link was clicked.
func (link *Link) Update(gtx C) bool {
	for {
		event, ok := gtx.Event(
			pointer.Filter{Target: link, Kinds: pointer.Enter | pointer.Leave},
		)
		if !ok {
			break
		}

		switch event := event.(type) {
		case pointer.Event:
			switch event.Kind {
			case pointer.Enter:
				link.hovering = true
			case pointer.Leave:
				link.hovering = false
			}
		}
	}

	var clicked bool
	for {
		e, ok := link.click.Update(gtx.Source)
		if !ok {
			break
		}
		if e.Kind == gesture.KindClick {
			clicked = true
		}
	}

	if clicked {
		link.clicked = true
	}

	return clicked
}

type LinkStyle struct {
	state *Link

	// Face defines the text style.
	Font         font.Font
	Color        color.NRGBA
	ClickedColor color.NRGBA
	// show as button or normal text?
	Style string
}

func NewLink(link *Link, style string) *LinkStyle {
	return &LinkStyle{
		state: link,
		Style: style,
	}
}

func (ls *LinkStyle) Layout(gtx C, th *theme.Theme) D {
	textColorMacro := op.Record(gtx.Ops)
	paint.ColorOp{Color: ls.Color}.Add(gtx.Ops)
	textColor := textColorMacro.Stop()

	clickedColorMacro := op.Record(gtx.Ops)
	paint.ColorOp{Color: ls.ClickedColor}.Add(gtx.Ops)
	clickedColor := clickedColorMacro.Stop()

	var textMaterial op.CallOp
	if ls.state.Clicked() {
		textMaterial = clickedColor
	} else {
		textMaterial = textColor
	}

	if ls.Color == (color.NRGBA{}) {
		ls.Color = th.Fg
	}

	ls.Color = color.NRGBA{B: 255, A: 255}
	ls.ClickedColor = color.NRGBA{R: 255, A: 255}

	if ls.Font == (font.Font{}) {
		ls.Font.Typeface = th.Face
	}

	macro := op.Record(gtx.Ops)
	dims := layout.Background{}.Layout(gtx,
		func(gtx C) D { return ls.layoutBackground(gtx, th) },
		func(gtx C) D { return ls.state.Layout(gtx, th.Shaper, ls.Font, th.TextSize, textMaterial) },
	)

	linkCall := macro.Stop()

	defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()

	// draw a underline below the text.
	var path clip.Path
	path.Begin(gtx.Ops)
	path.Move(f32.Point{Y: float32(dims.Size.Y)})
	path.Line(f32.Point{X: float32(dims.Size.X)})
	path.Close()
	paint.FillShape(gtx.Ops, ls.Color,
		clip.Stroke{
			Path:  path.End(),
			Width: 2,
		}.Op())

	event.Op(gtx.Ops, ls.state)
	linkCall.Add(gtx.Ops)

	return dims

}

func (ls *LinkStyle) layoutBackground(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	var fill color.NRGBA
	if ls.state.hovering {
		fill = misc.WithAlpha(th.Palette.Fg, th.HoverAlpha)
	} //else if ls.isSelected {
	// 	fill = misc.WithAlpha(th.Palette.ContrastBg, uint8(255))
	// }

	rr := gtx.Dp(unit.Dp(2))
	rect := clip.RRect{
		Rect: image.Rectangle{
			Max: image.Point{X: gtx.Constraints.Min.X, Y: gtx.Constraints.Min.Y},
		},
		NE: rr,
		SE: rr,
		NW: rr,
		SW: rr,
	}
	paint.FillShape(gtx.Ops, fill, rect.Op(gtx.Ops))
	return layout.Dimensions{Size: gtx.Constraints.Min}
}
