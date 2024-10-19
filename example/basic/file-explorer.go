package main

import (
	//"image"

	"github.com/oligo/gioview/filetree"
	"github.com/oligo/gioview/theme"
	"github.com/oligo/gioview/view"

	"gioui.org/layout"
	// "gioui.org/text"
)

var (
	ExplorerViewID = view.NewViewID("FileExplorerView")
)

type FileExplorerView struct {
	*view.BaseView
	fileExplorer *filetree.FileExplorer
}

func (vw *FileExplorerView) ID() view.ViewID {
	return ExplorerViewID
}

func (vw *FileExplorerView) Title() string {
	return "File Explorer"
}

func (vw *FileExplorerView) Layout(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	return vw.fileExplorer.Layout(gtx, th)
}

func (va *FileExplorerView) OnFinish() {
	va.BaseView.OnFinish()
	// Put your cleanup code here.
}

func NewFileExplorerView() view.View {
	v := &FileExplorerView{
		BaseView:     &view.BaseView{},
		fileExplorer: filetree.NewFileExplorer(),
	}

	return v
}
