package main

import (
	"os"

	"github/oligo/gioview/theme"
	"github/oligo/gioview/view"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
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

	ui.vm = vm
}

func main() {
	//defer profile.Start(profile.MemProfile).Stop()

	go func() {
		w := app.NewWindow()
		th := theme.NewTheme(".", true)
		ui := &UI{theme: th, window: w}
		err := ui.Loop()
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}()

	app.Main()

}
