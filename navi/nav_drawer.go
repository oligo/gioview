package navi

import (
	"image/color"
	"log"

	"github/oligo/gioview/misc"
	"github/oligo/gioview/theme"
	"github/oligo/gioview/view"

	"gioui.org/font"
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

type NavDrawer struct {
	vm           view.ViewManager
	selectedItem *NavItemStyle
	listItems    []NavSection
	listState    *widget.List
}

type NaviDrawerStyle struct {
	*NavDrawer
	Inset layout.Inset
	Bg    color.NRGBA
	Width unit.Dp
}

func NewNavDrawer(vm view.ViewManager) *NavDrawer {
	return &NavDrawer{
		vm: vm,
		listState: &widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
	}
}

func (nv *NavDrawer) AddSection(item NavSection) {
	nv.listItems = append(nv.listItems, item)
}

func (nv *NavDrawer) Layout(gtx C, th *theme.Theme) D {
	return material.List(th.Theme, nv.listState).Layout(gtx, len(nv.listItems), func(gtx C, index int) D {
		item := nv.listItems[index]
		dims := layout.Inset{
			Bottom: unit.Dp(5),
		}.Layout(gtx, func(gtx C) D {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					if item.Title() == "" {
						return layout.Dimensions{}
					}

					return layout.Inset{
						Bottom: unit.Dp(5),
					}.Layout(gtx, func(gtx C) D {
						label := material.Label(th.Theme, th.TextSize, item.Title())
						label.Color = misc.WithAlpha(th.Fg, 0xb6)
						label.TextSize = th.TextSize * 0.7
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx C) D {
					return item.Layout(gtx, th)
				}),
			)
		})

		return dims
	})
}

func (nv *NavDrawer) OnItemSelected(gtx C, item *NavItemStyle) {
	if item != nv.selectedItem {
		if nv.selectedItem != nil {
			nv.selectedItem.Unselect()
		}
		nv.selectedItem = item
	}

	if item != nil {
		intent := item.item.OnSelect(gtx)
		// An empty also refresh the UI so do not drop it.
		if err := nv.vm.RequestSwitch(intent); err != nil {
			log.Printf("switching to view %s error: %v", intent.Target, err)
		}
	}
}

func (ns NaviDrawerStyle) Layout(gtx C, th *theme.Theme) D {
	if ns.Inset == (layout.Inset{}) {
		ns.Inset = layout.Inset{
			Top:    unit.Dp(20),
			Bottom: unit.Dp(20),
			Left:   unit.Dp(20),
		}
	}
	if ns.Width <= 0 {
		ns.Width = unit.Dp(220)
	}

	gtx.Constraints.Max.X = gtx.Dp(ns.Width)
	gtx.Constraints.Min = gtx.Constraints.Max
	rect := clip.Rect{
		Max: gtx.Constraints.Max,
	}
	paint.FillShape(gtx.Ops, ns.Bg, rect.Op())

	return ns.Inset.Layout(gtx, func(gtx C) D {
		return ns.NavDrawer.Layout(gtx, th)
	})

}
