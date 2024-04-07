package tabview

import (
	"image"
	"image/color"

	"github/oligo/gioview/misc"
	"github/oligo/gioview/theme"

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
	"gioui.org/widget/material"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

var (
	horizontalInset = layout.Inset{Left: unit.Dp(2)}
	verticalInset   = layout.Inset{Top: unit.Dp(2)}
	horizontalFlex  = layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}
	verticalFlex    = layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}
)

type TabView struct {
	Axis         layout.Axis
	list         layout.List
	tabItems     []*TabItem
	currentView  int
	headerLength int
}

type TabItem struct {
	title string
	//tabColor color.NRGBA
	// tabWidth int
	inset     layout.Inset
	alignment text.Alignment

	click    gesture.Click
	hovering bool
	selected bool

	wgt func(gtx C, th *theme.Theme) D
}

func (item *TabItem) Update(gtx layout.Context) bool {
	for {
		event, ok := gtx.Event(
			pointer.Filter{Target: item, Kinds: pointer.Enter | pointer.Leave},
		)
		if !ok {
			break
		}

		switch event := event.(type) {
		case pointer.Event:
			switch event.Kind {
			case pointer.Enter:
				item.hovering = true
			case pointer.Leave:
				item.hovering = false
			case pointer.Cancel:
				item.hovering = false
			}
		}
	}

	var clicked bool
	for {
		e, ok := item.click.Update(gtx.Source)
		if !ok {
			break
		}
		if e.Kind == gesture.KindClick {
			clicked = true
			item.selected = true
		}
	}

	return clicked
}

func (item *TabItem) LayoutTab(gtx C, th *theme.Theme) D {
	item.Update(gtx)

	macro := op.Record(gtx.Ops)
	dims := item.layoutTab(gtx, th)
	call := macro.Stop()

	rect := clip.Rect(image.Rectangle{Max: dims.Size})
	defer rect.Push(gtx.Ops).Pop()

	item.click.Add(gtx.Ops)
	// register tag
	event.Op(gtx.Ops, item)
	call.Add(gtx.Ops)

	return dims
}

func (item *TabItem) layoutTab(gtx C, th *theme.Theme) D {
	return layout.Background{}.Layout(gtx,
		func(gtx C) D {
			return item.layoutTabBackground(gtx, th)
		},
		func(gtx C) D {
			return item.inset.Layout(gtx, func(gtx C) D {
				label := material.Label(th.Theme, th.TextSize*0.9, item.title)
				label.Font.Weight = font.Medium
				label.Alignment = item.alignment
				return label.Layout(gtx)
			})
		},
	)
}

func (item *TabItem) layoutTabBackground(gtx C, th *theme.Theme) D {
	var fill color.NRGBA
	if item.hovering {
		fill = misc.WithAlpha(th.Palette.Fg, th.AlphaPalette.Hover)
	} else if item.selected {
		fill = misc.WithAlpha(th.Palette.Fg, th.AlphaPalette.Selected)
	}

	rr := gtx.Dp(unit.Dp(4))
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

func (item *TabItem) LayoutWidget(gtx C, th *theme.Theme) D {
	return item.wgt(gtx, th)
}

func (tv *TabView) Layout(gtx C, th *theme.Theme) D {
	tv.Update(gtx)

	if len(tv.tabItems) <= 0 {
		return layout.Dimensions{}
	}

	maxTabSize := tv.calculateWidth(gtx, th)
	var direction layout.Direction
	var flex layout.Flex
	var tabAlign text.Alignment
	if tv.Axis == layout.Horizontal {
		direction = layout.Center
		flex = horizontalFlex
		tabAlign = text.Middle
	} else {
		direction = layout.N
		flex = verticalFlex
		tabAlign = text.Start
	}

	return flex.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return direction.Layout(gtx, func(gtx C) D {
				tv.list.Axis = tv.Axis
				tv.list.Alignment = layout.Start
				listDims := tv.list.Layout(gtx, len(tv.tabItems), func(gtx C, index int) D {
					gtx.Constraints.Min = maxTabSize
					item := tv.tabItems[index]
					item.alignment = tabAlign

					if index == 0 {
						return item.LayoutTab(gtx, th)
					}

					if tv.Axis == layout.Horizontal {
						return horizontalInset.Layout(gtx, func(gtx C) D {
							return item.LayoutTab(gtx, th)
						})
					} else {
						return verticalInset.Layout(gtx, func(gtx C) D {
							return item.LayoutTab(gtx, th)
						})
					}

				})

				if tv.Axis == layout.Horizontal {
					tv.headerLength = listDims.Size.X
				} else {
					tv.headerLength = listDims.Size.Y
				}
				return listDims
			})
		}),
		layout.Rigid(func(gtx C) D {
			if tv.Axis == layout.Horizontal {
				return layout.Spacer{Height: unit.Dp(2)}.Layout(gtx)
			} else {
				return layout.Spacer{Width: unit.Dp(2)}.Layout(gtx)
			}
		}),

		layout.Rigid(func(gtx C) D {
			if tv.Axis == layout.Horizontal {
				gtx.Constraints.Min.X = tv.headerLength
			} else {
				gtx.Constraints.Min.Y = tv.headerLength
			}
			return misc.Divider(tv.Axis, unit.Dp(0.5)).Layout(gtx, th)
		}),

		layout.Rigid(func(gtx C) D {
			if tv.Axis == layout.Horizontal {
				return layout.Spacer{Height: unit.Dp(20)}.Layout(gtx)
			} else {
				return layout.Spacer{Width: unit.Dp(20)}.Layout(gtx)
			}
		}),

		layout.Rigid(func(gtx C) D {
			return tv.tabItems[tv.currentView].LayoutWidget(gtx, th)
		}),
	)
}

func (tv *TabView) Update(gtx C) {
	for idx, item := range tv.tabItems {
		if item.Update(gtx) {
			// unselect last item
			lastItem := tv.tabItems[tv.currentView]
			if lastItem != nil && idx != tv.currentView {
				lastItem.selected = false
			}

			tv.currentView = idx
		}

		if tv.currentView == idx && !item.selected {
			item.selected = true
		}
	}
}

func (tv *TabView) calculateWidth(gtx C, th *theme.Theme) image.Point {
	fakeOps := new(op.Ops)
	current := gtx.Ops
	gtx.Ops = fakeOps
	maxSize := image.Point{}

	gtx.Constraints.Min = image.Point{}
	for _, item := range tv.tabItems {
		dims := item.layoutTab(gtx, th)
		if dims.Size.X > maxSize.X {
			maxSize.X = dims.Size.X
		}
		// if dims.Size.Y > maxSize.Y {
		// 	maxSize.Y = dims.Size.Y
		// }
	}

	gtx.Ops = current
	return maxSize
}

func NewTabView(inset layout.Inset, item ...*TabItem) *TabView {
	for _, i := range item {
		i.inset = inset
	}

	return &TabView{
		Axis:     layout.Horizontal, //default value
		tabItems: item,
	}
}

func NewTabItem(title string, wgt func(gtx C, th *theme.Theme) D) *TabItem {
	return &TabItem{
		title: title,
		wgt:   wgt,
	}
}
