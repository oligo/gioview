package explorer

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"io"
	"log"
	"slices"
	"strings"
	"unsafe"

	"gioui.org/gesture"
	"gioui.org/io/clipboard"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/oligo/gioview/menu"
	"github.com/oligo/gioview/misc"
	"github.com/oligo/gioview/navi"
	"github.com/oligo/gioview/theme"
	gv "github.com/oligo/gioview/widget"
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

var _ navi.NavItem = (*EntryNavItem)(nil)

type EntryNavItem struct {
	state          *EntryNode
	click          gesture.Click
	menuOptionFunc MenuOptionFunc
	onSelectFunc   OnSelectFunc

	parent   navi.NavItem
	children []navi.NavItem
	label    *gv.Editable
	expaned  bool
	needSync bool
	isCut    bool
}

type MenuOptionFunc func(gtx C, item *EntryNavItem) [][]menu.MenuOption
type OnSelectFunc func(item *EntryNode)

// Construct a file tree object that loads files and folders from rootDir.
// `menuOptionFunc` is used to define the operations allowed by context menu(use right click to active it).
// `onSelectFunc` defines what action to take when a navigable item is clicked (files or folders).
func NewEntryNavItem(rootDir string, menuOptionFunc MenuOptionFunc, onSelectFunc OnSelectFunc) (*EntryNavItem, error) {
	tree, err := NewFileTree(rootDir)
	if err != nil {
		return nil, err
	}

	return &EntryNavItem{
		parent:         nil,
		state:          tree,
		menuOptionFunc: menuOptionFunc,
		onSelectFunc:   onSelectFunc,
		expaned:        true,
	}, nil

}

func (eitem *EntryNavItem) icon() *widget.Icon {
	if eitem.state.Kind() == FolderNode {
		if eitem.expaned {
			return folderOpenIcon
		}
		return folderIcon
	}

	return fileIcon
}

func (eitem *EntryNavItem) OnSelect() {
	eitem.expaned = !eitem.expaned
	if eitem.expaned {
		eitem.needSync = true

		for _, child := range eitem.children {
			child := child.(*EntryNavItem)
			if child.isCut {
				child.isCut = false
			}
		}

	}

	if eitem.state.Kind() == FileNode && eitem.onSelectFunc != nil {
		eitem.onSelectFunc(eitem.state)
	}
}

func (eitem *EntryNavItem) Layout(gtx layout.Context, th *theme.Theme, textColor color.NRGBA) D {
	eitem.Update(gtx)

	macro := op.Record(gtx.Ops)
	dims := eitem.layout(gtx, th, textColor)
	call := macro.Stop()

	defer pointer.PassOp{}.Push(gtx.Ops).Pop()
	defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
	if eitem.isCut {
		defer paint.PushOpacity(gtx.Ops, 0.6).Pop()
	}
	event.Op(gtx.Ops, eitem)
	call.Add(gtx.Ops)

	return dims
}

func (eitem *EntryNavItem) layout(gtx layout.Context, th *theme.Theme, textColor color.NRGBA) D {
	if eitem.label == nil {
		eitem.label = gv.EditableLabel(eitem.state.Name(), func(text string) {
			err := eitem.state.UpdateName(text)
			if err != nil {
				log.Println("err: ", err)
			}
		})
	}

	eitem.label.Color = textColor
	eitem.label.TextSize = th.TextSize

	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if eitem.icon() == nil {
				return layout.Dimensions{}
			}
			return layout.Inset{Right: unit.Dp(6)}.Layout(gtx, func(gtx C) D {
				iconColor := th.ContrastBg
				return misc.Icon{Icon: eitem.icon(), Color: iconColor, Size: unit.Dp(th.TextSize)}.Layout(gtx, th)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return eitem.label.Layout(gtx, th)
		}),
	)

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

	if eitem.children == nil || eitem.needSync {
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

func (eitem *EntryNavItem) Refresh() {
	eitem.expaned = true
	eitem.needSync = true
}

// StartEditing inits and focused on the editor to accept user input.
func (eitem *EntryNavItem) StartEditing(gtx C) {
	eitem.label.SetEditing(true)
}

// Create file or subfolder under the current folder.
// File or subfolder is inserted at the beginning of the children.
func (eitem *EntryNavItem) CreateChild(gtx C, kind NodeKind, postAction func(node *EntryNode)) error {
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
		return err
	}

	childNode := eitem.state.Children()[0]

	child := &EntryNavItem{
		parent:         eitem,
		state:          childNode,
		menuOptionFunc: eitem.menuOptionFunc,
		onSelectFunc:   eitem.onSelectFunc,
		expaned:        false,
		needSync:       false,
	}

	child.label = gv.EditableLabel(childNode.Name(), func(text string) {
		err := childNode.UpdateName(text)
		if err != nil {
			log.Println("update name err: ", err)
		}
		if postAction != nil {
			postAction(childNode)
		}
	})

	eitem.children = slices.Insert[[]navi.NavItem, navi.NavItem](eitem.children, 0, child)
	// focus the child input
	child.StartEditing(gtx)

	return nil
}

func (eitem *EntryNavItem) Remove() error {
	if eitem.parent == nil {
		return errors.New("cannot remove root dir/file")
	}

	err := eitem.state.Delete()
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

// read data from clipboard.
func (eitem *EntryNavItem) OnPaste(data string, removeOld bool, src *EntryNavItem) error {
	// when paste destination is a normal file node, use its parent dir to ease the CUT/COPY operations.
	dest := eitem
	if !eitem.IsDir() && eitem.parent != nil {
		dest = eitem.parent.(*EntryNavItem)
	}

	pathes := strings.Split(string(data), "\n")
	if removeOld {
		for _, p := range pathes {
			err := dest.state.Move(p)
			if err != nil {
				return err
			}

			if src != nil && src.parent != nil {
				parent := src.parent.(*EntryNavItem)
				parent.children = slices.DeleteFunc(parent.children, func(chd navi.NavItem) bool {
					entry := chd.(*EntryNavItem)
					return entry.Path() == p
				})
			}
		}
	} else {
		for _, p := range pathes {
			err := dest.state.Copy(p)
			if err != nil {
				return err
			}
		}
	}

	dest.needSync = true
	dest.expaned = true
	return nil
}

func (eitem *EntryNavItem) OnCopyOrCut(gtx C, isCut bool) {
	gtx.Execute(clipboard.WriteCmd{Type: mimeText, Data: io.NopCloser(asPayload(eitem, eitem.Path(), isCut))})
	eitem.isCut = isCut
}

func (eitem *EntryNavItem) Update(gtx C) error {
	for {
		ke, ok := gtx.Event(
			// focus conflicts with editable. so subscribe editable's key events here.
			pointer.Filter{Target: eitem, Kinds: pointer.Press | pointer.Release},
			key.Filter{Focus: eitem.label, Name: "C", Required: key.ModShortcut},
			key.Filter{Focus: eitem.label, Name: "V", Required: key.ModShortcut},
			key.Filter{Focus: eitem.label, Name: "X", Required: key.ModShortcut},
			transfer.TargetFilter{Target: eitem, Type: mimeOctStream},
			transfer.TargetFilter{Target: eitem, Type: mimeText},
		)

		if !ok {
			break
		}

		switch event := ke.(type) {
		case key.Event:
			if !event.Modifiers.Contain(key.ModShortcut) {
				break
			}

			switch event.Name {
			// Initiate a paste operation, by requesting the clipboard contents; other
			// half is in DataEvent.
			case "V":
				gtx.Execute(clipboard.ReadCmd{Tag: eitem})

			// Copy or Cut selection -- ignored if nothing selected.
			case "C", "X":
				eitem.OnCopyOrCut(gtx, event.Name == "X")
			}

		case transfer.DataEvent:
			// read the clipboard content:
			reader := event.Open()
			defer reader.Close()
			content, err := io.ReadAll(reader)
			if err != nil {
				return err
			}

			defer gtx.Execute(op.InvalidateCmd{})

			//FIXME: clipboard data might be invalid file path.
			if event.Type == mimeText {
				p, err := toPayload(content)
				if err == nil {
					if err := eitem.OnPaste(p.Data, p.IsCut, p.GetSrc()); err != nil {
						return err
					}
				} else {
					if err := eitem.OnPaste(string(content), false, nil); err != nil {
						return err
					}
				}
			}
		case pointer.Event:
			if event.Buttons.Contain(pointer.ButtonPrimary) && event.Kind == pointer.Press {
				eitem.OnSelect()
			}

		}
	}

	return nil
}

const (
	mimeText      = "application/text"
	mimeOctStream = "application/octet-stream"
)

type payload struct {
	IsCut bool    `json:"isCut"`
	Data  string  `json:"data"`
	Src   uintptr `json:"src"`
}

func (p *payload) GetSrc() *EntryNavItem {
	return (*EntryNavItem)(unsafe.Pointer(p.Src))
}

func asPayload(src *EntryNavItem, data string, isCut bool) io.Reader {
	p := payload{Data: data, IsCut: isCut, Src: uintptr(unsafe.Pointer(src))}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(p)
	if err != nil {
		panic(err)
	}

	return strings.NewReader(buf.String())
}

func toPayload(buf []byte) (*payload, error) {
	p := payload{}
	err := json.Unmarshal(buf, &p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
