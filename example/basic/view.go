package main

import (
	//"image"

	gioimg "github.com/oligo/gioview/image"
	"github.com/oligo/gioview/misc"
	"github.com/oligo/gioview/page"
	"github.com/oligo/gioview/tabview"
	"github.com/oligo/gioview/theme"
	"github.com/oligo/gioview/view"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var (
	ExampleViewID = view.NewViewID("Example")
)

type ExampleView struct {
	*view.BaseView
	page.PageStyle
	vm                view.ViewManager
	horizontalTabView *tabview.TabView
	verticalTabView   *tabview.TabView
	img               *gioimg.ImageSource
	showDialogBtn     widget.Clickable
}

func (vw *ExampleView) ID() view.ViewID {
	return ExampleViewID
}

func (vw *ExampleView) Title() string {
	return "Tabviews & Image"
}

func (vw *ExampleView) Layout(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	vw.Update(gtx)
	vw.Padding = unit.Dp(30)
	return vw.PageStyle.Layout(gtx, th, func(gtx C) D {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

			layout.Rigid(func(gtx C) D {
				return material.H6(th.Theme, "1. Loading image from local filesystem or from network").Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),

			layout.Rigid(func(gtx layout.Context) layout.Dimensions {

				if vw.img == nil {
					vw.img = loadImg()
				}

				//sz := 480
				//gtx.Constraints = layout.Exact(image.Pt(sz, sz))
				// gtx.Constraints.Max.X = 500
				// gtx.Constraints.Min = gtx.Constraints.Max
				img := gioimg.NewImage(vw.img)
				img.Radius = unit.Dp(12)
				img.Fit = widget.Unscaled
				img.Position = layout.N
				return img.Layout(gtx)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(25)}.Layout),

			layout.Rigid(func(gtx C) D {
				return material.H6(th.Theme, "2. Horizontal tab view").Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),

			layout.Rigid(func(gtx C) D {
				if vw.horizontalTabView == nil {
					vw.horizontalTabView = tabview.NewTabView(layout.Horizontal, vw.buildTabItems()...)
				}
				return vw.horizontalTabView.Layout(gtx, th)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(40)}.Layout),

			layout.Rigid(func(gtx C) D {
				return misc.Divider(layout.Horizontal, unit.Dp(0.5)).Layout(gtx, th)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),

			layout.Rigid(func(gtx C) D {
				return material.H6(th.Theme, "3. Vertical tab view").Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),

			layout.Rigid(func(gtx C) D {
				if vw.verticalTabView == nil {
					vw.verticalTabView = tabview.NewTabView(layout.Vertical, vw.buildTabItems()...)
				}
				return vw.verticalTabView.Layout(gtx, th)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
			layout.Rigid(func(gtx C) D {
				return material.H6(th.Theme, "4. Click to open a modal view").Layout(gtx)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			layout.Rigid(func(gtx C) D {
				if vw.showDialogBtn.Clicked(gtx) {
					err := vw.vm.RequestSwitch(view.Intent{Target: EditorExampleViewID, ShowAsModal: true})
					if err != nil {
						panic(err)
					}
				}
				return material.Button(th.Theme, &vw.showDialogBtn, "Click me to open a modal view").Layout(gtx)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(30)}.Layout),
		)
	})
}

func (vw *ExampleView) buildTabItems() []*tabview.TabItem {
	inset := layout.Inset{
		Left:   unit.Dp(12),
		Right:  unit.Dp(12),
		Top:    unit.Dp(8),
		Bottom: unit.Dp(8),
	}

	var tabItems []*tabview.TabItem
	tabItems = append(tabItems, tabview.SimpleTabItem(inset, "Tab 1", func(gtx C, th *theme.Theme) D {
		return vw.layoutTab(gtx, th, "Tab one")
	}))

	tabItems = append(tabItems, tabview.SimpleTabItem(inset, "A long tab name", func(gtx C, th *theme.Theme) D {
		return vw.layoutTab(gtx, th, "Tab two")
	}))

	tabItems = append(tabItems, tabview.SimpleTabItem(inset, "Tab 3", func(gtx C, th *theme.Theme) D {
		return vw.layoutTab(gtx, th, "Tab three")
	}))

	tabItems = append(tabItems, tabview.SimpleTabItem(inset, "Tab 4", func(gtx C, th *theme.Theme) D {
		return vw.layoutTab(gtx, th, "Tab four")
	}))

	tabItems = append(tabItems, tabview.SimpleTabItem(inset, "Tab 5", func(gtx C, th *theme.Theme) D {
		return vw.layoutTab(gtx, th, "Tab five")
	}))

	return tabItems

}

func (va *ExampleView) layoutTab(gtx C, th *theme.Theme, content string) D {
	return layout.Center.Layout(gtx, func(gtx C) D {
		label := material.Label(th.Theme, th.TextSize*0.9, content)
		label.Font.Typeface = font.Typeface("Go Mono")
		return label.Layout(gtx)
	})
}

func (va *ExampleView) Update(gtx layout.Context) {

}

func (va *ExampleView) OnFinish() {
	va.BaseView.OnFinish()
	// Put your cleanup code here.
}

func NewExampleView(vm view.ViewManager) view.View {
	return &ExampleView{
		BaseView: &view.BaseView{},
		vm:       vm,
	}
}

func loadImg() *gioimg.ImageSource {
	return gioimg.ImageFromFile("https://cdn.pixabay.com/photo/2013/04/04/12/34/mountains-100367_1280.jpg")
}
