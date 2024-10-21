package explorer

import (
	"errors"
	"image/color"
	"log"
	"slices"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
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
	title string
	root  *navi.NavItemStyle
}

type EntryNavItem struct {
	state          *EntryNode
	menuOptionFunc MenuOptionFunc
	onSelectFunc   OnSelectFunc

	parent     navi.NavItem
	children   []navi.NavItem
	nameEditor *widget.Editor
	expaned    bool
	needSync   bool
	isEditing  bool

	// FolderIcon     *widget.Icon
	// FolderOpenIcon *widget.Icon
	// FileIcon       *widget.Icon
}

type MenuOptionFunc func(gtx C, item *EntryNavItem) [][]menu.MenuOption
type OnSelectFunc func(gtx C, item *EntryNode) view.Intent

// Construct a FileTreeNav object that loads files and folders from rootDir. The skipFolders
// parameter allows you to specify folder name prefixes to exclude from the navigation.
func NewFileTreeNav(drawer *navi.NavDrawer, title string, navRoot *EntryNavItem) *FileTreeNav {
	return &FileTreeNav{
		title: title,
		root:  navi.NewNavItem(navRoot, drawer),
	}
}

func (tn *FileTreeNav) Attach(drawer *navi.NavDrawer) {
	// NOOP
}

func (tn *FileTreeNav) Title() string {
	return tn.title
}

func (tn *FileTreeNav) Layout(gtx C, th *theme.Theme) D {
	return tn.root.Layout(gtx, th)
}

// Construct a file tree object that loads files and folders from rootDir.
// `skipFolders` allows you to specify folder name prefixes to exclude from the navigation.
// `menuOptionFunc` is used to define the operations allowed by context menu(use right click to active it).
// `onSelectFunc` defines what action to take when a navigable item is clicked (files or folders).
func NewEntryNavItem(rootDir string, menuOptionFunc MenuOptionFunc, onSelectFunc OnSelectFunc) *EntryNavItem {
	tree, err := NewFileTree(rootDir)
	if err != nil {
		log.Fatal(err)
	}

	//tree.Print()

	if err != nil {
		log.Println("load file tree failed", err)
		return nil
	}

	return &EntryNavItem{
		parent:         nil,
		state:          tree,
		menuOptionFunc: menuOptionFunc,
		onSelectFunc:   onSelectFunc,
		expaned:        true,
	}

}

func (eitem *EntryNavItem) Icon() *widget.Icon {
	if eitem.state.Kind() == FolderNode {
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

	// move focus to the nav item clicked.
	if !gtx.Focused(eitem.nameEditor) {
		gtx.Execute(key.FocusCmd{Tag: eitem})
	}

	if eitem.state.Kind() == FileNode && eitem.onSelectFunc != nil {
		return eitem.onSelectFunc(gtx, eitem.state)
	}

	return view.Intent{}

}

func (eitem *EntryNavItem) Layout(gtx layout.Context, th *theme.Theme, textColor color.NRGBA) D {
	eitem.update(gtx)

	if eitem.isEditing {
		return eitem.layoutEditArea(gtx, th)
	}

	lb := material.Label(th.Theme, th.TextSize, eitem.state.Name())
	lb.Color = textColor
	return lb.Layout(gtx)
}

func (eitem *EntryNavItem) layoutEditArea(gtx C, th *theme.Theme) D {
	macro := op.Record(gtx.Ops)
	dims := material.Editor(th.Theme, eitem.nameEditor, "").Layout(gtx)
	call := macro.Stop()

	rect := clip.Rect{Max: dims.Size}

	paint.FillShape(gtx.Ops, misc.WithAlpha(th.Fg, 0x60),
		clip.Stroke{
			Path:  rect.Path(),
			Width: float32(gtx.Dp(1)),
		}.Op(),
	)

	call.Add(gtx.Ops)
	return dims
}

func (eitem *EntryNavItem) IsDir() bool {
	return eitem.state.IsDir()
}

func (eitem *EntryNavItem) SetMenuOptions(menuOptionFunc MenuOptionFunc) {
	eitem.menuOptionFunc = menuOptionFunc
}

func (eitem *EntryNavItem) ContextMenuOptions(gtx C) ([][]menu.MenuOption, bool) {

	if eitem.menuOptionFunc != nil {
		return eitem.menuOptionFunc(gtx, eitem), false
	}

	return nil, false
}

func (eitem *EntryNavItem) Children() []navi.NavItem {
	if eitem.state.Kind() == FileNode {
		return nil
	}

	if !eitem.expaned {
		return nil
	}

	if eitem.children == nil {
		eitem.buildChildren(true)
	}

	if eitem.needSync {
		eitem.buildChildren(true)
		eitem.needSync = false
	}

	return eitem.children
}

func (eitem *EntryNavItem) buildChildren(sync bool) {
	eitem.children = eitem.children[:0]
	if sync {
		err := eitem.state.Refresh(hiddenFileFilter)
		if err != nil {
			log.Println(err)
		}
	}
	for _, c := range eitem.state.Children() {
		eitem.children = append(eitem.children, &EntryNavItem{
			parent:         eitem,
			state:          c,
			menuOptionFunc: eitem.menuOptionFunc,
			onSelectFunc:   eitem.onSelectFunc,
			expaned:        false,
			needSync:       false,
		})
	}
}

// StartEditing inits and focused on the editor to accept user input.
func (eitem *EntryNavItem) StartEditing(gtx C) {
	if eitem.nameEditor == nil {
		eitem.nameEditor = &widget.Editor{
			SingleLine: true,
			MaxLen:     256, // windows has max filename length 256
			Submit:     true,
		}
	}

	eitem.nameEditor.SetText(eitem.state.Name())
	gtx.Execute(key.FocusCmd{Tag: eitem.nameEditor})
	eitem.nameEditor.SetCaret(0, len(eitem.nameEditor.Text()))
	eitem.isEditing = true
}

// IsEditing is used to indicate the current node is being edited.
func (eitem *EntryNavItem) IsEditing(gtx C) bool {
	return eitem.nameEditor != nil && eitem.isEditing
}

// update process eitem edit events. name is updated when one of the following situation
// occurred:
// 1. user pressed Enter key after input
// 2. current node's editor lost focus but text changed.
func (eitem *EntryNavItem) update(gtx C) {
	if eitem.nameEditor == nil {
		return
	}

	saveName := false
	//check if user pressed Enter key after input
	for {
		event, ok := eitem.nameEditor.Update(gtx)
		if !ok {
			break
		}

		switch event.(type) {
		case widget.SubmitEvent:
			// enter pressed
			saveName = true
		}
	}

	//  or editor lost focus but name is changed
	if eitem.IsEditing(gtx) && !gtx.Focused(eitem.nameEditor) && eitem.state.Name() != eitem.nameEditor.Text() {
		saveName = true
	}

	if saveName {
		err := eitem.state.UpdateName(eitem.nameEditor.Text())
		if err != nil {
			log.Println("err: ", err)
		}
		eitem.isEditing = false
	}
}

// Create file or subfolder under the current folder.
// File or subfolder is inserted at the beginning of the children.
func (eitem *EntryNavItem) CreateChild(gtx C, kind NodeKind) error {
	if eitem.state.Kind() == FileNode {
		return nil
	}

	var err error
	if kind == FileNode {
		err = eitem.state.AddChild("new file", FileNode)
	} else {
		err = eitem.state.AddChild("new folder", FolderNode)
	}

	if err != nil {
		// TODO: use modal to show the error if user provided one.
		log.Println(err)
		return err
	}

	//eitem.StartEditing(gtx)

	child := &EntryNavItem{
		parent:         eitem,
		state:          eitem.state.Children()[0],
		menuOptionFunc: eitem.menuOptionFunc,
		onSelectFunc:   eitem.onSelectFunc,
		expaned:        false,
		needSync:       false,
	}

	eitem.children = slices.Insert[[]navi.NavItem, navi.NavItem](eitem.children, 0, child)
	// focus the child input
	child.StartEditing(gtx)

	return nil
}

func (eitem *EntryNavItem) Remove() error {
	if eitem.parent == nil {
		return errors.New("cannot remove root dir/file")
	}

	err := eitem.state.Delete(true)
	if err != nil {
		return err
	}

	(eitem.parent).(*EntryNavItem).needSync = true
	return nil
}

// File or folder name of this node
func (eitem *EntryNavItem) Name() string {
	return eitem.state.Name()
}

// File or folder path of this node
func (eitem *EntryNavItem) Path() string {
	return eitem.state.Path
}

// EntryNode kind of this node
func (eitem *EntryNavItem) Kind() NodeKind {
	return eitem.state.Kind()
}
