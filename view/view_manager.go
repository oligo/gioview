package view

import (
	"fmt"
	"net/url"

	"gioui.org/layout"
	"looz.ws/gioview/theme"
)

type ViewID string

type Intent struct {
	Target      ViewID
	Params      map[string]interface{}
	Referer     url.URL
	ShowAsModal bool
	// indicates the provider to create a new view instance and show up
	// in a new tab
	RequireNew bool
}

func (i Intent) Location() url.URL {
	return BuildURL(i.Target, i.Params)
}

type ViewAction struct {
	Layout func(gtx C, th *theme.Theme) D
}

type View interface {
	Actions() []ViewAction
	Layout(gtx layout.Context, th *theme.Theme) layout.Dimensions
	OnNavTo(intent Intent) error
	ID() ViewID
	Location() url.URL
	Title() string
	// set the view to finished state and do some cleanup ops.
	OnFinish()
	Finished() bool
}

// ViewProvider provides View. This enables us to write dynamic views.
type ViewProvider interface {
	// ID returns ViewID of the View that provider provides.
	ID() ViewID
	Provide(intent Intent) View
}

// A multi tab view manager. Each view pushed to the manager have a
// managed history stack.
type ViewManager interface {
	// Register is used to register all views before all the view rendering happens.
	// Use provider to enable us to use dynamically constructed views.
	Register(provider ViewProvider) error
	// RegisterWithProvider(provider ViewProvider) error

	// try to swith the current view to the requested view. If referer of the intent equals to
	// the current viewID of the current tab, the requested view should be routed and pushed to
	// to the existing viewstack(current tab). Otherwise a new viewstack for the intent is created(a new tab)
	// if there's no duplicate active view (first views of the stacks).
	RequestSwitch(intent Intent) error

	// OpenedViews return the views on top of the stack of each tab.
	OpenedViews() []View
	// Close the current tab and move backwards to the previous one if there's any.
	CloseTab(idx int)
	SwitchTab(idx int)
	// CurrentView returns the top most view of the current tab.
	CurrentView() View
	// current tab index
	CurrentViewIndex() int
	// Navigate back to the last view if there's any and pop out the current view.
	// It returns the view that is to be rendered.
	NavBack() View
	// Check is there are any naviBack-able views in the current stack or not. This should not
	// count for the current view.
	HasPrev() bool
	// return the next view that is intened to be shown in the modal layer. It returns nil if
	// there's no shownAsModal intent request.
	NextModalView() *ModalView
	// finish the last modal view handling.
	FinishModalView()
	// refresh the window
	Invalidate()

	//Reset resets internal states of the VM
	Reset()
}

func BuildURL(target ViewID, params map[string]interface{}) url.URL {
	var urlParams = make(url.Values)
	for k, v := range params {
		urlParams.Add(k, fmt.Sprintf("%v", v))
	}

	return url.URL{
		Scheme:   "gioview",
		Host:     "local",
		Path:     string(target),
		RawQuery: urlParams.Encode(),
	}
}
