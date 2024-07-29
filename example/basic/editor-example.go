package main

import (
	//"image"

	"image/color"
	"regexp"

	"github.com/oligo/gioview/editor"
	"github.com/oligo/gioview/theme"
	"github.com/oligo/gioview/view"
	"github.com/oligo/gioview/widget"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"

	// "gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

var (
	EditorExampleViewID = view.NewViewID("EditorExampleView")
)

type EditorExample struct {
	*view.BaseView
	ed           *editor.Editor
	patternInput widget.TextField
}

func (vw *EditorExample) ID() view.ViewID {
	return EditorExampleViewID
}

func (vw *EditorExample) Title() string {
	return "Editor Example"
}

func (vw *EditorExample) Layout(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Vertical,
		Alignment: layout.Middle,
	}.Layout(gtx,

		layout.Rigid(func(gtx C) D {
			return material.Label(th.Theme, th.TextSize, "Editor with text highlighting example").Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

		layout.Rigid(func(gtx C) D {
			vw.patternInput.Padding = unit.Dp(8)
			vw.patternInput.HelperText = "Illustrating colored text painting in text editor."
			vw.patternInput.MaxChars = 64
			return vw.patternInput.Layout(gtx, th, "Regex of substring hightlighted")
		}),

		layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

		layout.Rigid(func(gtx C) D {
			editorConf := &editor.EditorConf{
				Shaper:          th.Shaper,
				TextColor:       th.Fg,
				Bg:              th.Bg,
				SelectionColor:  th.ContrastBg,
				TypeFace:        "Go, Helvetica, Arial, sans-serif",
				TextSize:        th.TextSize,
				LineHeightScale: 1.6,
				ColorScheme:     "default",
				ShowLineNum:     true,
			}

			vw.ed.UpdateTextStyles(stylingText(vw.ed.Text(), vw.patternInput.Text()))

			return layout.Inset{
				Left:  unit.Dp(10),
				Right: unit.Dp(10),
			}.Layout(gtx, func(gtx C) D {
				return editor.NewEditor(vw.ed, editorConf, "type to input...").Layout(gtx)
			})
		}),
	)

}

func (va *EditorExample) OnFinish() {
	va.BaseView.OnFinish()
	// Put your cleanup code here.
}

func NewEditorExample() view.View {
	v := &EditorExample{
		BaseView: &view.BaseView{},
		ed:       &editor.Editor{},
	}

	v.ed.SetText(sampleText, false)
	return v
}

func colorToOp(textColor color.NRGBA) op.CallOp {
	ops := new(op.Ops)

	m := op.Record(ops)
	paint.ColorOp{Color: textColor}.Add(ops)
	return m.Stop()
}

func stylingText(text string, pattern string) []*editor.TextStyle {
	var styles []*editor.TextStyle

	re := regexp.MustCompile(pattern)
	matches := re.FindAllIndex([]byte(text), -1)
	for _, match := range matches {
		styles = append(styles, &editor.TextStyle{
			Start:      match[0],
			End:        match[1],
			Color:      colorToOp(color.NRGBA{R: 255, A: 212}),
			Background: colorToOp(color.NRGBA{R: 215, G: 215, B: 215, A: 100}),
		})
	}

	return styles
}

var sampleText = `
Gio-view is a third-party toolkit that simplifies building user interfaces (UIs) for desktop applications written with the Gio library in Go.
It provides pre-built components and widgets, saving you time and effort compared to creating everything from scratch. Gio-view offers a 
more user-friendly experience for developers new to Gio.
`
