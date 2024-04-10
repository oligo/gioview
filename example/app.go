package main

import (
	"image/color"
	"os"

	"github.com/oligo/gioview/theme"
	"github.com/oligo/gioview/view"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	// "github.com/pkg/profile"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type UI struct {
	window *app.Window
	theme  *theme.Theme
	vm     *HomeView
}

func (ui *UI) Loop() error {
	ui.registerViews()

	var ops op.Ops
	for {
		e := ui.window.NextEvent()

		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			ui.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (ui *UI) layout(gtx C) D {
	return ui.vm.Layout(gtx, ui.theme)
}

func (ui *UI) registerViews() {
	vm := newHome(ui.window)
	// vm.Register(&view.EmptyView{})
	vm.Register(view.Provide(
		ExampleViewID,
		func() view.View { return NewExampleView() },
	))

	vm.Register(view.Provide(
		ExampleView2ID,
		func() view.View { return NewExampleView2() },
	))

	ui.vm = vm
}

func main() {
	//defer profile.Start(profile.MemProfile).Stop()

	go func() {
		w := app.NewWindow()
		th := theme.NewTheme(".", nil, false)
		th.TextSize = unit.Sp(12)
		th.Bg2 = color.NRGBA{R: 225, G: 225, B: 225, A: 255}

		ui := &UI{theme: th, window: w}
		err := ui.Loop()
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}()

	app.Main()

}
