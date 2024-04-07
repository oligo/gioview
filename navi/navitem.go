package navi

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"

	"github/oligo/gioview/list"
	"github/oligo/gioview/menu"
	"github/oligo/gioview/misc"
	"github/oligo/gioview/theme"
	"github/oligo/gioview/view"
)

var (
	moreIcon, _ = widget.NewIcon(icons.NavigationMoreHoriz)
)

type NavSection interface {
	Title() string
	Layout(gtx C, th *theme.Theme) D
	Attach(d *NavDrawer)
}

type NavItem interface {
	OnSelect(gtx layout.Context) view.Intent
	Icon() *widget.Icon
	Layout(gtx layout.Context, th *theme.Theme, textColor color.NRGBA) D
	// when there's menu options, a context menu should be attached to this navItem.
	// The returned boolean value suggest the position of the popup menu should be at
	// fixed position or not. NavItemStyle should place a clickable icon to guide user interactions.
	ContextMenuOptions() ([][]menu.MenuOption, bool)
	Children() []NavItem
}

type NavItemStyle struct {
	drawer      *NavDrawer
	item        NavItem
	label       *list.InteractiveLabel
	menu        *menu.ContextMenu
	fixMenuPos  bool
	showMenuBtn widget.Clickable

	padding   unit.Dp
	childList layout.List
	children  []*NavItemStyle
}

func (n *NavItemStyle) IsSelected() bool {
	return n.label.IsSelected()
}

func (n *NavItemStyle) Unselect() {
	n.label.Unselect()
}

func (n *NavItemStyle) Update(gtx C) bool {
	// handle naviitem events
	if n.label.Update(gtx) {
		n.drawer.OnItemSelected(gtx, n)
		return true
	}

	return false
}

func (n *NavItemStyle) layoutRoot(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := layout.Inset{Bottom: unit.Dp(2)}.Layout(gtx, func(gtx C) D {
		return n.label.Layout(gtx, th, func(gtx C, color color.NRGBA) D {
			return layout.UniformInset(n.padding).Layout(gtx, func(gtx C) D {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						if n.item.Icon() == nil {
							return layout.Dimensions{}
						}
						return layout.Inset{Right: unit.Dp(6)}.Layout(gtx, func(gtx C) D {
							iconColor := th.ContrastBg
							if n.label.IsSelected() {
								iconColor = th.ContrastFg
							}
							return misc.Icon{Icon: n.item.Icon(), Color: iconColor}.Layout(gtx, th)
						})
					}),
					layout.Flexed(1, func(gtx C) D {
						return layout.W.Layout(gtx, func(gtx C) D {
							return n.item.Layout(gtx, th, color)
						})
					}),

					layout.Rigid(func(gtx C) D {
						if n.menu == nil || !n.fixMenuPos {
							return D{}
						}

						return material.Clickable(gtx, &n.showMenuBtn, func(gtx C) D {
							alpha := 0xb6
							if n.showMenuBtn.Hovered() {
								alpha = 0xff
							}
							dims := misc.Icon{Icon: moreIcon, Color: misc.WithAlpha(color, uint8(alpha))}.Layout(gtx, th)
							// a tricky way to let contextual menu show up just near the button.
							n.menu.Layout(gtx, th)
							return dims
						})

					}),
				)
			})
		})
	})
	c := macro.Stop()
	defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
	c.Add(gtx.Ops)

	// if menu is not fixed position, let it follow the pointer.
	if n.menu != nil && !n.fixMenuPos {
		n.menu.Layout(gtx, th)
	}

	return dims
}

func (n *NavItemStyle) Layout(gtx C, th *theme.Theme) D {
	if n.label == nil {
		n.label = &list.InteractiveLabel{}
	}

	if n.padding <= 0 {
		n.padding = unit.Dp(1)
	}

	n.Update(gtx)

	itemChildren := n.item.Children()
	if len(n.item.Children()) <= 0 {
		return n.layoutRoot(gtx, th)
	}

	if len(n.children) != len(itemChildren) {
		n.children = n.children[:0]
		for _, child := range itemChildren {
			n.children = append(n.children, NewNavItem(child, n.drawer))
		}
	}

	n.childList.Axis = layout.Vertical

	return layout.Flex{
		Axis:      layout.Vertical,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return n.layoutRoot(gtx, th)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{
				Top:    unit.Dp(4),
				Bottom: unit.Dp(4),
				Left:   unit.Dp(10),
				Right:  unit.Dp(10),
			}.Layout(gtx, func(gtx C) D {
				return n.childList.Layout(gtx, len(n.children), func(gtx C, index int) D {
					return n.children[index].Layout(gtx, th)
				})
			})
		}),
	)

}

func NewNavItem(item NavItem, drawer *NavDrawer) *NavItemStyle {
	style := &NavItemStyle{
		item:       item,
		label:      &list.InteractiveLabel{},
		drawer:     drawer,
		fixMenuPos: false,
	}
	menuOpts, fixPos := item.ContextMenuOptions()
	if len(menuOpts) > 0 {
		style.menu = menu.NewContextMenu(menuOpts, fixPos)
		style.fixMenuPos = fixPos
	}

	return style
}

type simpleItemSection struct {
	item *NavItemStyle
}

type simpleNavItem struct {
	icon       *widget.Icon
	name       string
	targetView view.ViewID
}

func (item simpleNavItem) OnSelect(gtx C) view.Intent {
	return view.Intent{Target: item.targetView}
}

func (item simpleNavItem) Icon() *widget.Icon {
	return item.icon
}

func (item simpleNavItem) Layout(gtx C, th *theme.Theme, textColor color.NRGBA) D {
	label := material.Label(th.Theme, th.TextSize, item.name)
	label.Color = textColor
	return label.Layout(gtx)
}

func (item simpleNavItem) ContextMenuOptions() ([][]menu.MenuOption, bool) {
	return nil, false
}

func (item simpleNavItem) Children() []NavItem {
	return nil
}

func (ss simpleItemSection) Title() string {
	return ""
}

func (ss simpleItemSection) Layout(gtx C, th *theme.Theme) D {
	return ss.item.Layout(gtx, th)
}

func (ss simpleItemSection) Attach(d *NavDrawer) {
	d.AddSection(ss)
	ss.item.drawer = d
}

func SimpleItemSection(icon *widget.Icon, name string, targetView view.ViewID) NavSection {
	item := NewNavItem(simpleNavItem{icon: icon, name: name, targetView: targetView}, nil)
	return simpleItemSection{item: item}
}
