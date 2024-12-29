package main

import (
	"image/color"
	"slices"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/oligo/gioview/explorer"
	"github.com/oligo/gioview/menu"
	"github.com/oligo/gioview/misc"
	"github.com/oligo/gioview/navi"
	"github.com/oligo/gioview/theme"
	"github.com/oligo/gioview/view"
)

type NavSection interface {
	Title() string
	Layout(gtx C, th *theme.Theme) D
}

type NavDrawer struct {
	vm           view.ViewManager
	selectedItem *navi.NavTree
	listItems    []NavSection
	listState    *widget.List

	// used to set inset of each section.
	SectionInset layout.Inset
}

type NaviDrawerStyle struct {
	*NavDrawer
	Inset layout.Inset
	Bg    color.NRGBA
	Width unit.Dp
}

type FileTreeNav struct {
	title string
	root  *navi.NavTree
}

type simpleItemSection struct {
	item *navi.NavTree
}

type simpleNavItem struct {
	icon *widget.Icon
	name string
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

func (nv *NavDrawer) InsertAt(index int, item NavSection) {
	nv.listItems = slices.Insert(nv.listItems, index, []NavSection{item}...)
}

func (nv *NavDrawer) RemoveSection(index int) {
	nv.listItems = slices.Delete(nv.listItems, index, index)
}

func (nv *NavDrawer) Layout(gtx C, th *theme.Theme) D {
	if nv.SectionInset == (layout.Inset{}) {
		nv.SectionInset = layout.Inset{
			Bottom: unit.Dp(5),
		}
	}

	list := material.List(th.Theme, nv.listState)
	list.AnchorStrategy = material.Overlay
	return list.Layout(gtx, len(nv.listItems), func(gtx C, index int) D {
		item := nv.listItems[index]
		dims := nv.SectionInset.Layout(gtx, func(gtx C) D {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					if item.Title() == "" {
						return layout.Dimensions{}
					}

					return layout.Inset{
						Bottom: unit.Dp(1),
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

func (nv *NavDrawer) OnItemSelected(item *navi.NavTree) {
	if item != nv.selectedItem {
		if nv.selectedItem != nil {
			nv.selectedItem.Unselect()
		}
		nv.selectedItem = item
	}

	nv.vm.Invalidate()
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

func (item simpleNavItem) Layout(gtx C, th *theme.Theme, textColor color.NRGBA) D {
	label := material.Label(th.Theme, th.TextSize, item.name)
	label.Color = textColor
	return label.Layout(gtx)
}

func (item simpleNavItem) ContextMenuOptions(gtx C) ([][]menu.MenuOption, bool) {
	return nil, false
}

func (item simpleNavItem) Children() ([]navi.NavItem, bool) {
	return nil, false
}

func (ss simpleItemSection) Title() string {
	return ""
}

func (ss simpleItemSection) Layout(gtx C, th *theme.Theme) D {
	return ss.item.Layout(gtx, th)
}

func SimpleItemSection(icon *widget.Icon, name string, onSelect func(item *navi.NavTree)) NavSection {
	item := navi.NewNavItem(simpleNavItem{icon: icon, name: name}, onSelect)
	item.VerticalPadding = unit.Dp(4)
	return simpleItemSection{item: item}
}

// Construct a FileTreeNav object that loads files and folders from rootDir. The skipFolders
// parameter allows you to specify folder name prefixes to exclude from the navigation.
func NewFileTreeNav(title string, navRoot *explorer.EntryNavItem, onClick func(item *navi.NavTree)) *FileTreeNav {
	tree := &FileTreeNav{
		title: title,
		root:  navi.NewNavItem(navRoot, onClick),
	}
	tree.root.Indention = unit.Dp(8)
	tree.root.VerticalPadding = unit.Dp(4)
	return tree
}

func (tn *FileTreeNav) Title() string {
	return tn.title
}

func (tn *FileTreeNav) Layout(gtx C, th *theme.Theme) D {
	return tn.root.Layout(gtx, th)
}
