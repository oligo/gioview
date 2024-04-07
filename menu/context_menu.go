package menu

import (
	"image"
	"log"

	"github/oligo/gioview/misc"
	"github/oligo/gioview/theme"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type ContextMenu struct {
	contextArea ContextArea
	optionList  widget.List
	// Inset applied around the rendered contents of the state's Options field.
	inset         layout.Inset
	options       [][]MenuOption
	optionStates  []*widget.Clickable
	focusedOption int

	menuItems []layout.Widget
}

type MenuOption struct {
	Layout    func(gtx C, th *theme.Theme) D
	OnClicked func() error
}

func NewContextMenu(options [][]MenuOption, absPosition bool) *ContextMenu {
	m := &ContextMenu{
		inset: layout.Inset{
			Top:    unit.Dp(8),
			Bottom: unit.Dp(8),
		},
		optionList: widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		options: options,
	}

	if absPosition {
		m.contextArea.AbsolutePosition = true
		m.contextArea.Activation = pointer.ButtonPrimary
	}

	return m
}

func (m *ContextMenu) buildMenus(th *theme.Theme) []layout.Widget {
	if len(m.options) <= 0 || len(m.optionStates) > 0 {
		return nil
	}

	optionCnt := 0
	for _, opts := range m.options {
		optionCnt += len(opts)
	}

	if len(m.optionStates) != optionCnt {
		m.optionStates = m.optionStates[:0]
		for _, group := range m.options {
			for _ = range group {
				m.optionStates = append(m.optionStates, &widget.Clickable{})
			}
		}
	}

	menuItems := make([]layout.Widget, 0)

	idx := 0
	for i, group := range m.options {
		if i != 0 {
			menuItems = append(menuItems, func(gtx C) D {
				return layout.Inset{
					// list scrollbar on the right side has width of 10px or 20px in HiDP system ,
					Left:   unit.Dp(10),
					Bottom: unit.Dp(4),
				}.Layout(gtx, func(gtx C) D {
					return misc.Divider(layout.Horizontal, unit.Dp(1)).Layout(gtx, th)
				})
			})
		}

		for _, opt := range group {
			// closure captured opt
			opt := opt
			state := m.optionStates[idx]
			idx++
			menuItems = append(menuItems, func(gtx C) D {
				if state.Clicked(gtx) {
					m.contextArea.Dismiss()
					opt.OnClicked()
				}
				return m.layoutOption(gtx, th, state, func(gtx C) D {
					return opt.Layout(gtx, th)
				})
			})
		}
	}

	return menuItems
}

func (m *ContextMenu) Layout(gtx C, th *theme.Theme) D {
	m.Update(gtx)

	macro := op.Record(gtx.Ops)
	gtx.Constraints.Min = gtx.Constraints.Max
	dims := m.contextArea.Layout(gtx, func(gtx C) D {
		gtx.Constraints.Min = image.Point{}
		//gtx.Constraints.Max.Y = gtx.Dp(420)
		return m.layoutOptions(gtx, th)
	})
	menuOps := macro.Stop()

	// Important!!! Otherwise widget below the ContextMenu will not receive pointer events.
	defer pointer.PassOp{}.Push(gtx.Ops).Pop()
	defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
	menuOps.Add(gtx.Ops)
	event.Op(gtx.Ops, m)

	return dims
}

// layoutOptions renders the menu option list.
func (m *ContextMenu) layoutOptions(gtx C, th *theme.Theme) D {
	var fakeOps op.Ops
	originalOps := gtx.Ops
	gtx.Ops = &fakeOps
	maxWidth := 0

	if len(m.menuItems) <= 0 {
		m.menuItems = m.buildMenus(th)
	}

	for _, w := range m.menuItems {
		dims := w(gtx)
		if dims.Size.X > maxWidth {
			maxWidth = dims.Size.X
		}
	}
	gtx.Ops = originalOps

	return component.Surface(th.Theme).Layout(gtx, func(gtx C) D {
		macro := op.Record(gtx.Ops)
		dims := m.inset.Layout(gtx, func(gtx C) D {
			return material.List(th.Theme, &m.optionList).Layout(gtx, len(m.menuItems), func(gtx C, index int) D {
				gtx.Constraints.Min.X = maxWidth
				gtx.Constraints.Max.X = maxWidth
				return m.menuItems[index](gtx)
			})
		})
		call := macro.Stop()
		defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
		paint.ColorOp{Color: th.NaviBgColor}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		call.Add(gtx.Ops)

		return dims
	})

}

func (m *ContextMenu) Update(gtx C) {
	m.contextArea.PositionHint = layout.N
	if m.contextArea.Activated() {
		// let the menu be focused!
		gtx.Execute(key.FocusCmd{Tag: m})
		m.focusedOption = -1
	}

	if m.contextArea.Dismissed() {
		m.optionList.List.ScrollTo(0)
	}

	for {
		e, ok := gtx.Event(
			key.Filter{Focus: m, Name: key.NameUpArrow},
			key.Filter{Focus: m, Name: key.NameDownArrow},
			key.Filter{Focus: m, Name: key.NameLeftArrow},
			key.Filter{Focus: m, Name: key.NameRightArrow},
			key.Filter{Focus: m, Name: key.NameEnter},
			key.Filter{Focus: m, Name: key.NameReturn},
		)

		if !ok {
			break
		}

		switch e := e.(type) {
		case key.Event:
			log.Println("menu received key event", e)

			if e.Name == key.NameDownArrow {
				log.Println("down arrow pressed")

				m.focusedOption++
				if m.focusedOption >= len(m.menuItems) {
					m.focusedOption = 0
				}
				gtx.Execute(key.FocusCmd{Tag: m.optionStates[m.focusedOption]})
				log.Println("down arrow pressed")
			}
			if e.Name == key.NameUpArrow {
				log.Println("up arrow pressed")

				m.focusedOption--
				if m.focusedOption < 0 {
					m.focusedOption = len(m.menuItems) - 1
				}
				gtx.Execute(key.FocusCmd{Tag: m.optionStates[m.focusedOption]})
				log.Println("up arrow pressed")

			}
			if e.Name == key.NameEnter || e.Name == key.NameReturn {
				log.Println("enter or return key pressed")

				if m.focusedOption >= 0 && gtx.Focused(m.menuItems[m.focusedOption]) {
					// simulate a mouse click
					m.optionStates[m.focusedOption].Click()
				}
			}
		}
	}
}

func (m *ContextMenu) layoutOption(gtx C, th *theme.Theme, state *widget.Clickable, w layout.Widget) D {
	return layout.Inset{
		// list scrollbar on the right side has width of 10px or 20px in HiDP system ,
		Left:   unit.Dp(10),
		Bottom: unit.Dp(4),
	}.Layout(gtx, func(gtx C) D {
		return material.Clickable(gtx, state, func(gtx C) D {
			return layout.Inset{
				Left:   unit.Dp(20),
				Right:  unit.Dp(20),
				Top:    unit.Dp(2),
				Bottom: unit.Dp(2),
			}.Layout(gtx, func(gtx C) D {
				return w(gtx)
			})
		})
	})
}
