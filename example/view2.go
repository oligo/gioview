package main

import (
	//"image"

	"image/color"
	"regexp"

	"github.com/oligo/gioview/editor"
	"github.com/oligo/gioview/page"
	"github.com/oligo/gioview/theme"
	"github.com/oligo/gioview/view"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

var (
	ExampleView2ID = view.NewViewID("Example2")
)

type ExampleView2 struct {
	*view.BaseView
	page.PageStyle
	ed           *editor.Editor
	patternInput component.TextField
}

func (vw *ExampleView2) ID() view.ViewID {
	return ExampleView2ID
}

func (vw *ExampleView2) Title() string {
	return "Editor Example"
}

func (vw *ExampleView2) Layout(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	return vw.PageStyle.Layout(gtx, th, func(gtx C) D {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		}.Layout(gtx,

			layout.Rigid(func(gtx C) D {
				return material.Label(th.Theme, th.TextSize, "Editor with text highlighting example").Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

			layout.Rigid(func(gtx C) D {
				vw.patternInput.Alignment = text.Middle
				return vw.patternInput.Layout(gtx, th.Theme, "Regex of substring to be hightlighted")
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				editorConf := &editor.EditorConf{
					Shaper:          th.Shaper,
					TextColor:       th.Fg,
					Bg:              th.Bg,
					SelectionColor:  th.ContrastBg,
					TypeFace:        "Go, Helvetica, Arial, sans-serif",
					TextSize:        th.TextSize,
					LineHeightScale: 1.6,
					ColorScheme:     "default",
				}

				vw.ed.UpdateTextStyles(stylingText(vw.ed.Text(), vw.patternInput.Text()))

				return editor.NewEditor(vw.ed, editorConf, "type to input...").Layout(gtx)

			}),
		)
	})
}

func (va *ExampleView2) OnFinish() {
	va.BaseView.OnFinish()
	// Put your cleanup code here.
}

func NewExampleView2() *ExampleView2 {
	v := &ExampleView2{
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
			Color:      colorToOp(color.NRGBA{R: 255, A: 200}),
			Background: colorToOp(color.NRGBA{R: 215, G: 215, B: 215, A: 100}),
		})
	}

	return styles
}

var sampleText = `
Gio-view is a third-party toolkit that simplifies building user interfaces (UIs) 
for desktop applications written with the Gio library in Go. It provides pre-built components and widgets,
saving you time and effort compared to creating everything from scratch. Gio-view offers a more user-friendly 
experience for developers new to Gio.
`
