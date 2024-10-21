package widget

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	wg "gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/oligo/gioview/theme"
)

// Editable is an editable label that layouts an editor in responds to clicking.
type Editable struct {
	Text      string
	TextSize  unit.Sp
	OnChanged func(text string)

	editor *wg.Editor
	// clickable gesture.Click
	hovering bool
	editing  bool
}

func (e *Editable) SetEditing(editing bool) {
	e.editing = editing
}

func (e *Editable) Update(gtx C) {
	e.editor.SingleLine = true
	e.editor.Submit = true

	for {
		event, ok := gtx.Event(
			key.Filter{Focus: e.editor, Name: key.NameEscape},
			pointer.Filter{Target: e, Kinds: pointer.Enter | pointer.Leave},
		)
		if !ok {
			break
		}

		switch event := event.(type) {
		case key.Event:
			if event.Name == key.NameEscape {
				e.editing = false
				e.editor.SetText(e.Text)
			}
		case pointer.Event:
			switch event.Kind {
			case pointer.Enter:
				e.hovering = true
			case pointer.Leave:
				e.hovering = false
			case pointer.Cancel:
				e.hovering = false
			}
		}
	}

	if e.hovering {
		// As an indicator.
		pointer.CursorText.Add(gtx.Ops)
	} else {
		pointer.CursorDefault.Add(gtx.Ops)
	}

	// handle editor events:
	if ev, ok := e.editor.Update(gtx); ok {
		if _, ok := ev.(wg.SubmitEvent); ok {
			e.editing = false
			e.Text = e.editor.Text()
			if e.OnChanged != nil {
				e.OnChanged(e.Text)
			}
		}
	}

}

func (e *Editable) Layout(gtx C, th *theme.Theme) D {
	textSize := e.TextSize
	if textSize <= 0 {
		textSize = th.TextSize
	}

	if e.editing {
		return wg.Border{
			Color:        th.ContrastBg,
			Width:        unit.Dp(1),
			CornerRadius: unit.Dp(4),
		}.Layout(gtx, func(gtx C) D {
			return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx C) D {
				editor := material.Editor(th.Theme, e.editor, "")
				editor.TextSize = textSize
				return editor.Layout(gtx)
			})
		})
	}

	macro := op.Record(gtx.Ops)
	dims := material.Label(th.Theme, textSize, e.Text).Layout(gtx)
	callOp := macro.Stop()

	defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, e)
	callOp.Add(gtx.Ops)

	return dims
}
