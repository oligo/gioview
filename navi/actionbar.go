package navi

import (
	"time"

	"github/oligo/gioview/theme"
	"github/oligo/gioview/view"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	cmp "gioui.org/x/component"
)

const (
	actionAnimationDuration = time.Millisecond * 250
)

type ActionBar struct {
	actions     []view.ViewAction
	actionAnims []cmp.VisibilityAnimation
}

func (ab *ActionBar) SetActions(actions []view.ViewAction) {
	ab.actions = actions
	ab.actionAnims = make([]cmp.VisibilityAnimation, len(actions))
	for i := range ab.actionAnims {
		ab.actionAnims[i].Duration = actionAnimationDuration
	}
}

func (ab *ActionBar) Layout(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	gtx.Constraints.Min.Y = 0
	widthDp := float32(gtx.Constraints.Max.X) / gtx.Metric.PxPerDp

	// each item occupies 32 dp in size?
	visibleActionItems := int((widthDp / 32) - 1)
	if visibleActionItems < 0 {
		visibleActionItems = 0
	}

	visibleActionItems = min(visibleActionItems, len(ab.actions))
	var actions []layout.FlexChild
	for i := range ab.actions {
		action := ab.actions[i]
		anim := &ab.actionAnims[i]
		switch anim.State {
		case cmp.Visible:
			if i >= visibleActionItems {
				anim.Disappear(gtx.Now)
			}
		case cmp.Invisible:
			if i < visibleActionItems {
				anim.Appear(gtx.Now)
			}
		}
		actions = append(actions, layout.Rigid(func(gtx C) D {
			return ab.layoutAction(gtx, th, anim, action)
		}))
	}

	return layout.Flex{Alignment: layout.Middle}.Layout(gtx, actions...)
}

var actionButtonInset = layout.Inset{
	Top:    unit.Dp(2),
	Bottom: unit.Dp(2),
	Left:   unit.Dp(2),
}

func (ab *ActionBar) layoutAction(gtx layout.Context, th *theme.Theme, anim *cmp.VisibilityAnimation, action view.ViewAction) layout.Dimensions {
	if !anim.Visible() {
		return layout.Dimensions{}
	}
	animating := anim.Animating()
	var macro op.MacroOp
	if animating {
		macro = op.Record(gtx.Ops)
	}
	if !animating {
		return actionButtonInset.Layout(gtx, func(gtx C) D {
			return action.Layout(gtx, th)
		})
	}
	dims := actionButtonInset.Layout(gtx, func(gtx C) D {
		return action.Layout(gtx, th)
	})
	btnOp := macro.Stop()
	progress := anim.Revealed(gtx)
	dims.Size.X = int(progress * float32(dims.Size.X))

	// ensure this clip transformation stays local to this function
	defer clip.Rect{
		Max: dims.Size,
	}.Push(gtx.Ops).Pop()
	btnOp.Add(gtx.Ops)
	return dims
}
