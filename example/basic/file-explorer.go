package main

import (
	//"image"

	"image"
	"io"
	"os"
	"strings"

	"github.com/oligo/gioview/explorer"
	"github.com/oligo/gioview/theme"
	"github.com/oligo/gioview/view"

	"gioui.org/io/event"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/unit"
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

	msg1       string
	msg2       string
	msg3       string
	msg4       string
	dropedFile string
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
			reader, _ := fileChooser.ChooseFile(".jpg", ".png")
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

		layout.Rigid(func(gtx C) D {
			return layout.Center.Layout(gtx, func(gtx C) D {
				return vw.layoutDropArea(gtx, th)
			})
		}),
	)
}

func (vw *FileExplorerView) layoutDropArea(gtx C, th *theme.Theme) D {
	// Check for the received data.
	for {
		ev, ok := gtx.Event(transfer.TargetFilter{Target: vw, Type: explorer.EntryMIME})
		if !ok {
			break
		}
		switch e := ev.(type) {
		case transfer.DataEvent:
			data := e.Open()
			defer data.Close()
			content, _ := io.ReadAll(data)
			vw.dropedFile = string(content)
		}
	}

	gtx.Constraints = layout.Exact(image.Point{X: 500, Y: 300})

	defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, vw)

	widget.Border{
		Color: th.Fg,
		Width: unit.Dp(2),
	}.Layout(gtx, func(gtx C) D {
		return layout.UniformInset(unit.Dp(24)).Layout(gtx, func(gtx C) D {
			msg := "Drag file from the file tree to here"
			if vw.dropedFile != "" {
				msg = "File Received: " + vw.dropedFile
			}
			return material.Label(th.Theme, th.TextSize, msg).Layout(gtx)
		})
	})

	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func (vw *FileExplorerView) OnFinish() {
	vw.BaseView.OnFinish()
	// Put your cleanup code here.
}

func NewFileExplorerView() view.View {
	v := &FileExplorerView{
		BaseView: &view.BaseView{},
	}

	return v
}
