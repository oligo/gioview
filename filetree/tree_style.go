package filetree

import (
	"fmt"
	"image/color"
	"log"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/oligo/gioview/menu"
	"github.com/oligo/gioview/misc"
	"github.com/oligo/gioview/navi"
	"github.com/oligo/gioview/theme"
	"github.com/oligo/gioview/view"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

var (
	folderIcon, _     = widget.NewIcon(icons.FileFolder)
	folderOpenIcon, _ = widget.NewIcon(icons.FileFolderOpen)
	fileIcon, _       = widget.NewIcon(icons.ActionDescription)
)

var _ navi.NavSection = (*FileTreeNav)(nil)
var _ navi.NavItem = (*EntryNavItem)(nil)

type FileTreeNav struct {
	title  string
	tree   *FileTree
	drawer *navi.NavDrawer
	root   *navi.NavItemStyle
}

type EntryNavItem struct {
	entry          *EntryNode
	children       []navi.NavItem
	menuOptionFunc MenuOptionFunc
	expaned        bool
	needSync       bool

	// FolderIcon     *widget.Icon
	// FolderOpenIcon *widget.Icon
	// FileIcon       *widget.Icon
}

type MenuOptionFunc func(item *EntryNavItem) [][]menu.MenuOption

// Construct a FileTreeNav object that loads files and folders from rootDir. The skipFolders 
// parameter allows you to specify folder name prefixes to exclude from the navigation.
func NewFileTreeNav(title string, rootDir string, skipFolders []string) *FileTreeNav {
	tree := NewFileTree(rootDir, skipFolders)
	err := tree.Load()
	if err != nil {
		log.Println("load file tree failed", err)
		return nil
	}

	return &FileTreeNav{
		title: title,
		tree:  tree,
	}
}

func (tn *FileTreeNav) Attach(drawer *navi.NavDrawer) {
	tn.drawer = drawer
	tn.root = navi.NewNavItem(&EntryNavItem{entry: tn.tree.root}, tn.drawer)
}

func (tn *FileTreeNav) Title() string {
	return tn.title
}

func (tn *FileTreeNav) Layout(gtx C, th *theme.Theme) D {
	return tn.root.Layout(gtx, th)
}

func (eitem *EntryNavItem) Icon() *widget.Icon {
	if eitem.entry.Kind == FolderNode {
		if eitem.expaned {
			return folderOpenIcon
		}
		return folderIcon
	}

	return fileIcon
}

func (eitem *EntryNavItem) OnSelect(gtx C) view.Intent {
	eitem.expaned = !eitem.expaned
	if eitem.expaned {
		eitem.needSync = true
	}
	return view.Intent{}
}

func (eitem *EntryNavItem) Layout(gtx layout.Context, th *theme.Theme, textColor color.NRGBA) D {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			lb := material.Label(th.Theme, th.TextSize, eitem.entry.Name)
			lb.Color = textColor
			return lb.Layout(gtx)
		}),

		layout.Rigid(func(gtx C) D {
			if eitem.entry.Kind == FileNode {
				return D{}
			}

			return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx C) D {
				label := material.Label(th.Theme, th.TextSize, fmt.Sprintf("%d", len(eitem.entry.Children)))
				label.Color = misc.WithAlpha(textColor, 0xb6)
				label.TextSize = th.TextSize * 0.7
				return label.Layout(gtx)
			})
		}),
	)
}

func (eitem *EntryNavItem) SetMenuOptions(menuOptionFunc MenuOptionFunc) {
	eitem.menuOptionFunc = menuOptionFunc
}

func (eitem *EntryNavItem) ContextMenuOptions() ([][]menu.MenuOption, bool) {

	if eitem.menuOptionFunc != nil {
		return eitem.menuOptionFunc(eitem), false
	}

	return nil, false
}

func (eitem *EntryNavItem) Children() []navi.NavItem {
	if eitem.entry.Kind == FileNode {
		return nil
	}

	if !eitem.expaned {
		return nil
	}

	if eitem.children == nil {
		eitem.buildChildren()
	}

	if eitem.needSync {
		eitem.entry.Refresh()
		eitem.buildChildren()
		eitem.needSync = false
	}

	return eitem.children
}

func (eitem *EntryNavItem) buildChildren() {
	eitem.children = eitem.children[:0]
	for _, c := range eitem.entry.Children {
		eitem.children = append(eitem.children, &EntryNavItem{
			entry:    c,
			expaned:  false,
			needSync: false,
		})
	}
}
