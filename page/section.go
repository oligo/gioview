package page

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"looz.ws/gioview/theme"
	"looz.ws/gioview/misc"
)

type Section struct {
	Title string
}

type Labelpair struct {
	key   string
	value string
}

func (p Labelpair) Layout(gtx C, th *theme.Theme) D {
	return layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			label := material.Label(th.Theme, th.TextSize*0.8, p.key)
			label.Color = misc.WithAlpha(th.Palette.Fg, 0xb6)
			return label.Layout(gtx)
		}),

		layout.Rigid(func(gtx C) D {
			return layout.Spacer{Width: unit.Dp(15)}.Layout(gtx)
		}),

		layout.Rigid(func(gtx C) D {
			label := material.Label(th.Theme, th.TextSize*0.8, p.value)
			label.Color = th.Palette.Fg
			return label.Layout(gtx)
		}),
	)
}

type RowItem struct {
	Alignment layout.Alignment
}

func (item RowItem) Layout(gtx C, th *theme.Theme, labelDesc string, w layout.Widget) D {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: item.Alignment,
		Spacing:   layout.SpaceAround,
	}.Layout(gtx,
		layout.Flexed(0.25, func(gtx C) D {
			return layout.UniformInset(unit.Dp(3)).Layout(gtx,
				material.Label(th.Theme, th.TextSize*0.9, labelDesc).Layout)
		}),

		layout.Flexed(0.75, func(gtx C) D {
			return layout.UniformInset(unit.Dp(3)).Layout(gtx, w)
		}),
	)
}

func (s Section) Layout(gtx C, th *theme.Theme, w layout.Widget) D {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Spacer{Height: unit.Dp(4)}.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			header := material.Label(th.Theme, th.TextSize*1.2, s.Title)
			header.Color = misc.WithAlpha(th.Fg, 0xb6)
			return header.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Spacer{Height: unit.Dp(4)}.Layout(gtx)
		}),

		layout.Rigid(func(gtx C) D {
			return misc.Divider(layout.Horizontal, unit.Dp(1)).Layout(gtx, th)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Spacer{Height: unit.Dp(8)}.Layout(gtx)
		}),

		layout.Rigid(w),
	)
}
