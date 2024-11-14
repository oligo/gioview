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

	"github.com/oligo/gioview/list"
	"github.com/oligo/gioview/menu"
	"github.com/oligo/gioview/misc"
	"github.com/oligo/gioview/theme"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

var (
	moreIcon, _ = widget.NewIcon(icons.NavigationMoreHoriz)
)

var NavItemPadding = layout.Inset{
	Left:   unit.Dp(4),
	Right:  unit.Dp(4),
	Top:    unit.Dp(1),
	Bottom: unit.Dp(1),
}

type NavItem interface {
	Layout(gtx layout.Context, th *theme.Theme, textColor color.NRGBA) D
	// when there's menu options, a context menu should be attached to this navItem.
	// The returned boolean value suggest the position of the popup menu should be at
	// fixed position or not. NavTree should place a clickable icon to guide user interactions.
	ContextMenuOptions(gtx layout.Context) ([][]menu.MenuOption, bool)
	Children() []NavItem
}

type NavTree struct {
	item        NavItem
	label       *list.InteractiveLabel
	menu        *menu.ContextMenu
	fixMenuPos  bool
	showMenuBtn widget.Clickable

	childList layout.List
	children  []*NavTree
	OnClicked func(item *NavTree)
}

func (n *NavTree) IsSelected() bool {
	return n.label.IsSelected()
}

func (n *NavTree) Unselect() {
	n.label.Unselect()
}

func (n *NavTree) Update(gtx C) bool {
	if n.menu == nil {
		menuOpts, fixPos := n.item.ContextMenuOptions(gtx)
		if len(menuOpts) > 0 {
			n.menu = menu.NewContextMenu(menuOpts, fixPos)
			n.menu.PositionHint = layout.N
			n.fixMenuPos = fixPos
		}
	}

	// handle naviitem events
	if n.label.Update(gtx) && n.OnClicked != nil {
		n.OnClicked(n)
		return true
	}

	return false
}

func (n *NavTree) layoutRoot(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := layout.Inset{Bottom: unit.Dp(2)}.Layout(gtx, func(gtx C) D {
		return n.label.Layout(gtx, th, func(gtx C, color color.NRGBA) D {
			return NavItemPadding.Layout(gtx, func(gtx C) D {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
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

func (n *NavTree) Layout(gtx C, th *theme.Theme) D {
	if n.label == nil {
		n.label = &list.InteractiveLabel{}
	}

	n.Update(gtx)

	itemChildren := n.item.Children()
	if len(n.item.Children()) <= 0 {
		return n.layoutRoot(gtx, th)
	}

	if len(n.children) != len(itemChildren) {
		n.children = n.children[:0]
		for _, child := range itemChildren {
			n.children = append(n.children, NewNavItem(child, n.OnClicked))
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

func NewNavItem(item NavItem, onClicked func(item *NavTree)) *NavTree {
	style := &NavTree{
		item:       item,
		label:      &list.InteractiveLabel{},
		fixMenuPos: false,
		OnClicked:  onClicked,
	}

	return style
}
