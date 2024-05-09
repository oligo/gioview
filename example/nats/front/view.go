package main

import (
	//"image"

	gioimg "github.com/oligo/gioview/image"
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
	tabView *tabview.TabView
	img     *gioimg.ImageSource
	link    *view.Link
}

func (vw *ExampleView) ID() view.ViewID {
	return ExampleViewID
}

func (vw *ExampleView) Title() string {
	return "Tabviews & Image"
}

func (vw *ExampleView) Layout(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	vw.Update(gtx)
	return vw.PageStyle.Layout(gtx, th, func(gtx C) D {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

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
				return vw.layoutTabViews(gtx, th)
			}),

			layout.Rigid(func(gtx C) D {
				return view.NewLink(vw.link, "label").Layout(gtx, th)
			}),
		)
	})
}

func (va *ExampleView) layoutTabViews(gtx C, th *theme.Theme) D {
	if va.tabView == nil {
		va.tabView = tabview.NewTabView(
			layout.Inset{
				Left:   unit.Dp(12),
				Right:  unit.Dp(12),
				Top:    unit.Dp(8),
				Bottom: unit.Dp(8),
			},

			tabview.NewTabItem("Tab 1", func(gtx C, th *theme.Theme) D {
				return va.layoutTab(gtx, th, "Tab one")
			}),

			tabview.NewTabItem("Tab 2", func(gtx C, th *theme.Theme) D {
				return va.layoutTab(gtx, th, "Tab two")
			}),

			tabview.NewTabItem("Tab 3", func(gtx C, th *theme.Theme) D {
				return va.layoutTab(gtx, th, "Tab three")
			}),

			tabview.NewTabItem("Tab 4", func(gtx C, th *theme.Theme) D {
				return va.layoutTab(gtx, th, "Tab four")
			}),

			tabview.NewTabItem("Tab 5", func(gtx C, th *theme.Theme) D {
				return va.layoutTab(gtx, th, "Tab five")
			}),
		)
	}

	va.tabView.Axis = layout.Horizontal
	return va.tabView.Layout(gtx, th)
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

func NewExampleView() view.View {
	return &ExampleView{
		BaseView: &view.BaseView{},
		link:     &view.Link{Src: EditorExampleViewID, Title: "This is a great link"},
	}
}

func loadImg() *gioimg.ImageSource {
	return gioimg.ImageFromFile("https://cdn.pixabay.com/photo/2013/04/04/12/34/mountains-100367_1280.jpg")
}
