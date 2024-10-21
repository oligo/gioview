package explorer

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/oligo/gioview/misc"
	"github.com/oligo/gioview/theme"
	"github.com/oligo/gioview/view"
	gvwidget "github.com/oligo/gioview/widget"
)

type opKind uint8

const (
	openFileOp opKind = iota
	openFilesOp
	openFolderOp
	saveFileOp
)

var (
	fileChooserID = view.NewViewID("FileChooser")
)

type result struct {
	paths []string
	err   error
}

type FileChooser struct {
	vm         view.ViewManager
	resultChan chan result
}

type FileChooserDialog struct {
	*view.BaseView
	fileExplorer *FileExplorer
	resultChan   chan result
	op           opKind
}

type bottomPanel struct {
	op                 opKind
	saveFileInput      gvwidget.TextField
	addFolderInput     gvwidget.TextField
	addFolderBtn       widget.Clickable
	cancelAddFolderBtn widget.Clickable

	confirmBtn    widget.Clickable
	cancelBtn     widget.Clickable
	isAddingFoder bool
	confirmCb     func() error
	cancelCb      func()
	addFolderCb   func(folderName string) error
	err           error
}

func NewFileChooser(vm view.ViewManager) (*FileChooser, error) {
	err := vm.Register(fileChooserID, newFileChooserDialog)
	if err != nil {
		return nil, err
	}

	return &FileChooser{
		vm:         vm,
		resultChan: make(chan result),
	}, nil
}

// CreateFile opens the file chooser, and writes the given content into
// some file, which the use can choose the location. It's important to
// close the `io.WriteCloser`.
//
// It's a blocking call, you should call it on a separated goroutine.
func (fc *FileChooser) CreateFile(name string) (io.WriteCloser, error) {
	fc.show(saveFileOp, name)

	resp := <-fc.resultChan
	return os.Create(resp.paths[0])
}

// ChooseFile shows the file chooser, allowing the user to select a single file. It returns the
// file as a reader to user.
//
// This is a blocking call, you should call it in a seperated goroutine.
//
// Optionally, it's possible to set which file extensions is supported to
// be selected (such as `.jpg`, `.png`).
func (fc *FileChooser) ChooseFile(extensions ...string) (io.ReadCloser, error) {
	fc.show(openFileOp, "", extensions...)

	resp := <-fc.resultChan
	return os.Open(resp.paths[0])
}

// ChooseFile shows the file chooser, allowing the user to select multiple files. It returns the files as
// a list of reader to user.
// This is a blocking call, you should call it in a seperated goroutine.
//
// Optionally, it's possible to set which file extensions is supported to
// be selected (such as `.jpg`, `.png`).
func (fc *FileChooser) ChooseFiles(extensions ...string) ([]io.ReadCloser, error) {
	fc.show(openFilesOp, "", extensions...)

	resp := <-fc.resultChan
	readers := make([]io.ReadCloser, len(resp.paths))
	for idx, path := range resp.paths {
		d, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		readers[idx] = d
	}

	return readers, nil
}

// ChooseFolder shows the file chooser, allowing the user to select a single folder. It returns the folder
// path to user. This is a blocking call, you should call it in a seperated goroutine.
func (fc *FileChooser) ChooseFolder() (string, error) {
	fc.show(openFolderOp, "")

	resp := <-fc.resultChan
	return resp.paths[0], nil
}

func (fc *FileChooser) show(op opKind, filename string, extensions ...string) {
	params := map[string]interface{}{"resultChan": fc.resultChan, "op": op}
	if op == saveFileOp {
		params["filename"] = filename
	}

	params["filter"] = chooserFilter(op, extensions...)
	fc.vm.RequestSwitch(view.Intent{
		Target:      fileChooserID,
		ShowAsModal: true,
		Params:      params,
	})
}

func chooserFilter(op opKind, extensions ...string) EntryFilter {

	return func(info fs.FileInfo) bool {
		if op == openFolderOp {
			return info.IsDir()
		}
		// In other cases, folder should be kept.
		if info.IsDir() || len(extensions) <= 0 {
			return true
		}

		ext := filepath.Ext(info.Name())
		return slices.ContainsFunc(extensions, func(extension string) bool { return strings.EqualFold(extension, ext) })
	}
}

func (d *FileChooserDialog) ID() view.ViewID {
	return fileChooserID
}

func (vw *FileChooserDialog) Title() string {
	switch vw.op {
	case openFileOp:
		return "Open File"
	case openFilesOp:
		return "Open Files"
	case openFolderOp:
		return "Open Folder"
	case saveFileOp:
		return "Save File"
	}

	return "File Chooser"
}

func (vw *FileChooserDialog) OnNavTo(intent view.Intent) error {
	rc, ok := intent.Params["resultChan"]
	if !ok {
		return errors.New("missing mandatory params")
	}

	opVal, ok := intent.Params["op"]
	if !ok {
		return errors.New("missing mandatory params")
	}

	op := opVal.(opKind)
	vw.resultChan = rc.(chan result)
	vw.fileExplorer.bottomPanel.op = op
	vw.op = op
	if op == saveFileOp {
		param := intent.Params["filename"]
		vw.fileExplorer.bottomPanel.saveFileInput.SetText(param.(string))
	}

	vw.fileExplorer.bottomPanel.addFolderInput.SetText("untitled folder")

	vw.fileExplorer.bottomPanel.cancelCb = func() { vw.OnFinish() }
	vw.fileExplorer.bottomPanel.confirmCb = func() error {
		currentPath := vw.fileExplorer.viewer.entryTree.Path

		switch op {
		case saveFileOp:
			filename := strings.TrimSpace(vw.fileExplorer.bottomPanel.saveFileInput.Text())
			if filename == "" {
				return errors.New("empty filename")
			}

			vw.resultChan <- result{paths: []string{filepath.Join(currentPath, filename)}}
		case openFileOp, openFilesOp, openFolderOp:
			paths := make([]string, 0)
			for item := range vw.fileExplorer.viewer.selectedItems {
				paths = append(paths, item.node.Path)
			}
			vw.resultChan <- result{paths: paths}
		}

		vw.OnFinish()
		return nil
	}

	vw.fileExplorer.bottomPanel.addFolderCb = func(folderName string) error {
		return vw.fileExplorer.viewer.entryTree.AddChild(folderName, FolderNode)
	}

	if f, ok := intent.Params["filter"]; ok {
		vw.fileExplorer.entryFilter = f.(EntryFilter)
	}

	return nil
}

func (vw *FileChooserDialog) Layout(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	return vw.fileExplorer.Layout(gtx, th)
}

func newFileChooserDialog() view.View {
	return &FileChooserDialog{
		BaseView:     &view.BaseView{},
		fileExplorer: newFileExplorer(),
	}
}

func (p *bottomPanel) Layout(gtx C, th *theme.Theme) D {
	if p.addFolderBtn.Clicked(gtx) {
		if p.isAddingFoder {
			p.err = p.addFolderCb(p.addFolderInput.Text())
			if p.err == nil {
				p.isAddingFoder = false
			}
		} else {
			p.isAddingFoder = true
		}
	}
	if p.cancelAddFolderBtn.Clicked(gtx) {
		p.addFolderInput.Clear()
		p.err = nil
		p.isAddingFoder = false
	}

	if p.isAddingFoder && p.err != nil && p.addFolderInput.Changed() {
		p.err = nil
	}

	if p.cancelBtn.Clicked(gtx) {
		p.cancelCb()
	}
	if p.confirmBtn.Clicked(gtx) {
		p.err = p.confirmCb()
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			switch p.op {
			case saveFileOp:
				return p.layoutSaveFileForm(gtx, th)
			}

			return D{}
		}),
		layout.Rigid(func(gtx C) D {
			if p.op != saveFileOp {
				return D{}
			}
			return layout.Spacer{Height: unit.Dp(16)}.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			if p.err == nil {
				return D{}
			}
			return misc.LayoutErrorLabel(gtx, th, p.err)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
				Spacing:   layout.SpaceBetween,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx C) D {
					return p.layoutAddFolderForm(gtx, th)
				}),

				layout.Rigid(layout.Spacer{Width: unit.Dp(32)}.Layout),

				layout.Rigid(func(gtx C) D {
					return layout.Flex{
						Axis:      layout.Horizontal,
						Alignment: layout.Middle,
						Spacing:   layout.SpaceStart,
					}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							btn := material.Button(th.Theme, &p.cancelBtn, "Cancel")
							btn.Inset = layout.UniformInset(unit.Dp(6))
							btn.Background = th.Bg
							btn.Color = th.Fg
							return btn.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),

						layout.Rigid(func(gtx C) D {
							label := "Open"
							if p.op == saveFileOp {
								label = "Save"
							}
							btn := material.Button(th.Theme, &p.confirmBtn, label)
							btn.Inset = layout.UniformInset(unit.Dp(6))
							return btn.Layout(gtx)
						}),
					)
				}),
			)
		}),
	)
}

func (p *bottomPanel) layoutSaveFileForm(gtx C, th *theme.Theme) D {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			lb := material.Label(th.Theme, th.TextSize, "Save as:")
			lb.Color = misc.WithAlpha(th.Fg, 0xb6)
			return lb.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx C) D {
			p.saveFileInput.SingleLine = true
			p.saveFileInput.LabelOption = gvwidget.LabelOption{Alignment: gvwidget.Hidden}
			p.saveFileInput.Padding = unit.Dp(6)
			return p.saveFileInput.Layout(gtx, th, "")
		}),
	)
}

func (p *bottomPanel) layoutAddFolderForm(gtx C, th *theme.Theme) D {
	btn := "New Folder"
	if p.isAddingFoder {
		btn = "Save"
	}

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if !p.isAddingFoder {
				return D{}
			}

			gtx.Constraints.Max.X = gtx.Dp(unit.Dp(250))
			p.addFolderInput.SingleLine = true
			p.addFolderInput.LabelOption = gvwidget.LabelOption{Alignment: gvwidget.Hidden}
			p.addFolderInput.Padding = unit.Dp(6)
			return p.addFolderInput.Layout(gtx, th, "")
		}),
		layout.Rigid(func(gtx C) D {
			if !p.isAddingFoder {
				return D{}
			}
			return layout.Spacer{Width: unit.Dp(8)}.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			btn := material.Button(th.Theme, &p.addFolderBtn, btn)
			btn.Inset = layout.UniformInset(unit.Dp(6))
			return btn.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
		layout.Rigid(func(gtx C) D {
			if !p.isAddingFoder {
				return D{}
			}
			btn := material.Button(th.Theme, &p.cancelAddFolderBtn, "Cancel")
			btn.Inset = layout.UniformInset(unit.Dp(6))
			btn.Background = th.Bg
			btn.Color = th.Fg
			return btn.Layout(gtx)
		}),
	)
}
