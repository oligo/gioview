package navi

import (
	"image"
	"image/color"

	"github.com/oligo/gioview/misc"
	"github.com/oligo/gioview/theme"
	"github.com/oligo/gioview/view"

	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type TabEvent string

var (
	arrowIcon, _ = widget.NewIcon(icons.NavigationArrowBack)
	closeIcon, _ = widget.NewIcon(icons.NavigationClose)
)

const (
	TabSelectedEvent = TabEvent("TabSelected")
	TabClosedEvent   = TabEvent("TabClosed")
)

type TabbarOptions struct {
	Height            unit.Dp
	MaxVisibleActions int
}

type Tabbar struct {
	vm       view.ViewManager
	arrowBtn widget.Clickable
	list     *layout.List
	tabs     []*Tab
	options  *TabbarOptions
}

type Tab struct {
	vw         view.View
	tabClick   gesture.Click
	closeBtn   widget.Clickable
	isSelected bool
	hovering   bool
	events     []TabEvent

	// action bar for the current view.
	actionBar *ActionBar
}

func (tb *Tabbar) Layout(gtx C, th *theme.Theme) D {
	if tb.arrowBtn.Clicked(gtx) {
		tb.vm.NavBack()
	}

	arrowAlpha := 0xb6
	if !tb.vm.HasPrev() {
		arrowAlpha = 0x30
	}

	tabViews := tb.vm.OpenedViews()
	if len(tb.tabs) != len(tabViews) {
		// rebuilding tabs
		if len(tb.tabs) > 0 {
			tb.tabs = tb.tabs[:0]
		}
		for _, v := range tabViews {
			tb.tabs = append(tb.tabs, newTab(v, tb.options.MaxVisibleActions))
		}
	}

	var currentTab *Tab

	for idx, v := range tabViews {
		tab := tb.tabs[idx]
		for _, evt := range tab.Update(gtx) {
			switch evt {
			case TabSelectedEvent:
				tb.vm.SwitchTab(idx)
			case TabClosedEvent:
				// wait for the next frame to rebuild tabs
				tb.vm.CloseTab(idx)
			}
		}
		// sync tab state
		tab.isSelected = tb.vm.CurrentViewIndex() == idx
		tab.vw = v
		if tab.IsSelected() {
			currentTab = tab
		}
	}

	gtx.Constraints.Max.Y = gtx.Dp(tb.options.Height)
	gtx.Constraints.Min = gtx.Constraints.Max
	rect := clip.Rect(image.Rectangle{Max: gtx.Constraints.Max})
	paint.FillShape(gtx.Ops, misc.WithAlpha(th.Bg, 0x20), rect.Op())

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Left:  unit.Dp(10),
						Right: unit.Dp(10),
					}.Layout(gtx, func(gtx C) D {
						// arrow symbol
						return layout.Center.Layout(gtx, func(gtx C) D {
							return material.Clickable(gtx, &tb.arrowBtn, func(gtx C) D {
								return misc.Icon{Icon: arrowIcon, Color: misc.WithAlpha(th.Fg, uint8(arrowAlpha))}.Layout(gtx, th)
							})
						})
					})
				}),
				layout.Flexed(0.8, func(gtx C) D {
					// FIXME: As pointed out in this todo, layout.List does not scroll when laid out horizontally:
					// https://todo.sr.ht/~eliasnaur/gio/530
					return tb.list.Layout(gtx, len(tb.tabs), func(gtx C, index int) D {
						return tb.tabs[index].Layout(gtx, th)
					})
				}),

				layout.Flexed(0.2, func(gtx C) D {
					return layout.E.Layout(gtx, func(gtx C) D {
						return tb.layoutActions(gtx, th, currentTab)
					})
				}),
			)
		}),
		layout.Rigid(func(gtx C) D {
			return misc.Divider(layout.Horizontal, unit.Dp(0.5)).Layout(gtx, th)
		}),
	)

}

func (tb *Tabbar) layoutActions(gtx C, th *theme.Theme, tab *Tab) D {
	if tab == nil || len(tab.vw.Actions()) <= 0 {
		return layout.Dimensions{}
	}

	return layout.Inset{Right: unit.Dp(10)}.Layout(gtx, func(gtx C) D {
		return tab.actionBar.Layout(gtx, th)
	})
}

func NewTabbar(vm view.ViewManager, options *TabbarOptions) *Tabbar {
	tb := &Tabbar{
		vm:      vm,
		list:    &layout.List{Axis: layout.Horizontal, Alignment: layout.Middle},
		options: options,
	}
	if options == nil {
		tb.options = &TabbarOptions{
			Height:            unit.Dp(28),
			MaxVisibleActions: 999, // unlimited visible actions
		}
	}
	if tb.options.Height == 0 {
		tb.options.Height = unit.Dp(28)
	}
	if tb.options.MaxVisibleActions < 0 {
		tb.options.MaxVisibleActions = 0
	}

	return tb
}

func newTab(vw view.View, maxVisibleActions int) *Tab {
	tab := &Tab{vw: vw, actionBar: &ActionBar{}}
	tab.actionBar.SetActions(vw.Actions(), maxVisibleActions)
	return tab
}

func (tab *Tab) IsSelected() bool {
	return tab.isSelected
}

func (tab *Tab) Layout(gtx C, th *theme.Theme) D {
	tab.Update(gtx)

	macro := op.Record(gtx.Ops)
	dims := layout.Background{}.Layout(gtx,
		func(gtx C) D { return tab.layoutBackground(gtx, th) },
		func(gtx C) D {
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
			color := th.Fg
			if tab.isSelected {
				color = th.ContrastFg
			}
			return layout.Center.Layout(gtx, func(gtx C) D {
				return layout.Inset{
					Left:  unit.Dp(18),
					Right: unit.Dp(2),
				}.Layout(gtx, func(gtx C) D {
					return layout.Flex{
						Axis:      layout.Horizontal,
						Alignment: layout.Middle,
					}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							label := material.Label(th.Theme, th.TextSize*0.9, tab.vw.Title())
							label.Color = color
							return label.Layout(gtx)
						}),
						layout.Rigid(func(gtx C) D {
							iconAlpha := uint8(1)
							if tab.hovering {
								iconAlpha = uint8(200)
							}
							return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx C) D {
								return material.Clickable(gtx, &tab.closeBtn, func(gtx C) D {
									return misc.Icon{Icon: closeIcon,
										Color: misc.WithAlpha(color, iconAlpha),
										Size:  max(16, unit.Dp(16*th.TextSize/14)),
									}.Layout(gtx, th)
								})
							})

						}),
					)

				})
			})
		},
	)

	tabOps := macro.Stop()

	defer clip.Rect(image.Rectangle{
		Max: dims.Size,
	}).Push(gtx.Ops).Pop()

	tab.tabClick.Add(gtx.Ops)
	// register event tag
	event.Op(gtx.Ops, tab)
	tabOps.Add(gtx.Ops)
	return dims
}

func (tab *Tab) layoutBackground(gtx C, th *theme.Theme) D {
	if !tab.isSelected && !tab.hovering {
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}

	var fill color.NRGBA
	if tab.isSelected {
		fill = th.Palette.ContrastBg
	} else if tab.hovering {
		fill = misc.WithAlpha(th.Palette.ContrastBg, th.HoverAlpha)
	}
	rect := clip.Rect{
		Max: image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Max.Y},
	}
	paint.FillShape(gtx.Ops, fill, rect.Op())
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

func (tab *Tab) Update(gtx C) []TabEvent {
	for {
		event, ok := gtx.Event(
			pointer.Filter{Target: tab, Kinds: pointer.Enter | pointer.Leave},
		)
		if !ok {
			break
		}

		switch event := event.(type) {
		case pointer.Event:
			switch event.Kind {
			case pointer.Enter:
				tab.hovering = true
			case pointer.Leave:
				tab.hovering = false
			case pointer.Cancel:
				tab.hovering = false
			}
		}
	}

	tab.events = tab.events[:0]
	for {
		e, ok := tab.tabClick.Update(gtx.Source)
		if !ok {
			break
		}

		if e.Kind == gesture.KindClick {
			tab.isSelected = true
			tab.events = append(tab.events, TabSelectedEvent)
		}
	}

	if tab.closeBtn.Clicked(gtx) {
		tab.events = append(tab.events, TabClosedEvent)
	}

	return tab.events
}
