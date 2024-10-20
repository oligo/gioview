package main

import (
	//"image"

	"os"
	"strings"

	"github.com/oligo/gioview/theme"
	"github.com/oligo/gioview/view"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	// "gioui.org/text"
)

var (
	ExplorerViewID = view.NewViewID("FileExplorerView")
)

type FileExplorerView struct {
	*view.BaseView
	openFileBtn   widget.Clickable
	openFilesBtn  widget.Clickable
	saveFileBtn   widget.Clickable
	openFolderBtn widget.Clickable

	msg1 string
	msg2 string
	msg3 string
	msg4 string
}

func (vw *FileExplorerView) ID() view.ViewID {
	return ExplorerViewID
}

func (vw *FileExplorerView) Title() string {
	return "File Explorer"
}

func (vw *FileExplorerView) Layout(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	if vw.openFileBtn.Clicked(gtx) {
		go func() {
			reader, _ := fileChooser.ChooseFile()
			defer reader.Close()
			file := reader.(*os.File)
			vw.msg1 = file.Name()
		}()
	}

	if vw.openFilesBtn.Clicked(gtx) {
		go func() {
			readers, _ := fileChooser.ChooseFiles()
			for _, reader := range readers {
				defer reader.Close()
				file := reader.(*os.File)
				vw.msg2 += file.Name() + "\n"
			}
			vw.msg2 = strings.TrimSpace(vw.msg2)
		}()
	}

	if vw.openFolderBtn.Clicked(gtx) {
		go func() {
			vw.msg3, _ = fileChooser.ChooseFolder()
		}()
	}

	if vw.saveFileBtn.Clicked(gtx) {
		go func() {
			writer, _ := fileChooser.CreateFile("abcdefg.txt")
			defer writer.Close()
			file := writer.(*os.File)
			file.WriteString("A test message written to a file. File will be replaced if exists")
			vw.msg4 = file.Name()
		}()
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return material.Button(th.Theme, &vw.openFileBtn, "Open a file").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return material.Label(th.Theme, th.TextSize, vw.msg1).Layout(gtx)
		}),

		layout.Rigid(func(gtx C) D {
			return material.Button(th.Theme, &vw.openFilesBtn, "Open multiple files").Layout(gtx)
		}),

		layout.Rigid(func(gtx C) D {
			return material.Label(th.Theme, th.TextSize, vw.msg2).Layout(gtx)
		}),

		layout.Rigid(func(gtx C) D {
			return material.Button(th.Theme, &vw.openFolderBtn, "Open a folder").Layout(gtx)

		}),
		layout.Rigid(func(gtx C) D {
			return material.Label(th.Theme, th.TextSize, vw.msg3).Layout(gtx)

		}),

		layout.Rigid(func(gtx C) D {
			return material.Button(th.Theme, &vw.saveFileBtn, "Save a file named abcdefg.txt").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return material.Label(th.Theme, th.TextSize, "File saved: "+vw.msg4).Layout(gtx)

		}),
	)
}

func (va *FileExplorerView) OnFinish() {
	va.BaseView.OnFinish()
	// Put your cleanup code here.
}

func NewFileExplorerView() view.View {
	v := &FileExplorerView{
		BaseView: &view.BaseView{},
	}

	return v
}
