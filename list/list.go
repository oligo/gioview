package list

import (
	"image/color"
	"sync"

	"github/oligo/gioview/theme"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type ListContent interface {
	SetSelected(selected bool)
	// text should be rendered using contentColor
	Layout(gtx layout.Context, th *theme.Theme, contentColor color.NRGBA) layout.Dimensions
}

type SelectableList struct {
	listState *widget.List
	listItems []*ListItem
	// items to be put into list items. They should be update during event handing, because list widget
	// does not handle concurrent list modification.
	newItems []*ListItem

	selectedIndex int
	updateLock    sync.Mutex
}

func NewSelectableList(contents []ListContent) *SelectableList {
	sl := &SelectableList{
		listState: &widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		selectedIndex: -1,
	}

	sl.populateItems(contents)

	return sl
}

func (sl *SelectableList) populateItems(contents []ListContent) {
	if contents == nil {
		return
	}

	listItems := make([]*ListItem, len(contents))
	for i, content := range contents {
		listItems[i] = &ListItem{
			content: content,
		}
	}

	sl.newItems = listItems
}

func (sl *SelectableList) refresh() {
	sl.listItems = sl.newItems
	sl.newItems = nil
}

func (sl *SelectableList) Layout(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	sl.Update(gtx)

	return material.List(th.Theme, sl.listState).Layout(gtx, len(sl.listItems), func(gtx C, index int) D {
		// should not happen?
		if index >= len(sl.listItems) {
			return layout.Dimensions{}
		}
		return sl.listItems[index].Layout(gtx, th)
	})
}

func (sl *SelectableList) UpdateListItems(contents []ListContent) {
	sl.updateLock.Lock()
	defer sl.updateLock.Unlock()
	sl.populateItems(contents)
}

// return whether current selected item is clicked.
func (sl *SelectableList) Update(gtx C) bool {
	if sl.newItems != nil {
		sl.refresh()
	}

	var clicked bool
	for idx := range sl.listItems {
		item := sl.listItems[idx]
		clicked = item.Update(gtx)
		if clicked && item.isSelected() {
			// selectionChanged = sl.changeSelected(idx)
			sl.changeSelected(idx)
			break
		}
	}

	return clicked
}

func (sl *SelectableList) SelectionChanged(gtx C) bool {
	return sl.Update(gtx)
}

func (sl *SelectableList) changeSelected(index int) bool {
	// reset
	if index < 0 {
		if sl.selectedIndex >= 0 {
			sl.listItems[sl.selectedIndex].setSelected(false)
		}

		sl.selectedIndex = -1
		return true
	}

	if sl.selectedIndex < 0 && index >= 0 {
		sl.selectedIndex = index
		sl.listItems[sl.selectedIndex].setSelected(true)
		return true
	}

	if index == sl.selectedIndex && sl.listItems[index].isSelected() {
		return false
	}

	// when deleting the last item, the last selected index is larger than the max index of the new list items
	if sl.selectedIndex < len(sl.listItems) {
		sl.listItems[sl.selectedIndex].setSelected(false)
	}
	sl.selectedIndex = index
	return true
}

func (sl *SelectableList) SelectedItem() *ListItem {
	if sl.selectedIndex < 0 {
		return nil
	}

	if len(sl.listItems) > 0 {
		return sl.listItems[sl.selectedIndex]
	}

	return nil
}

type ListItem struct {
	label   InteractiveLabel
	content ListContent
}

func (li *ListItem) GetContent() ListContent {
	return li.content
}

func (li *ListItem) Update(gtx layout.Context) bool {
	clicked := li.label.Update(gtx)
	if li.label.IsSelected() {
		li.content.SetSelected(true)
	} else {
		li.content.SetSelected(false)
	}
	return clicked
}

func (li *ListItem) setSelected(selected bool) {
	li.label.isSelected = selected
	li.content.SetSelected(selected)
}

func (li *ListItem) isSelected() bool {
	return li.label.IsSelected()
}

func (li *ListItem) Layout(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	li.Update(gtx)
	return layout.Inset{
		Top:    unit.Dp(4),
		Bottom: unit.Dp(4),
		Left:   unit.Dp(2),
		Right:  unit.Dp(2),
	}.Layout(gtx, func(gtx C) D {
		return li.label.Layout(gtx, th, func(gtx C, contentColor color.NRGBA) D {
			return li.layoutContent(gtx, th, contentColor)
		})
	})
}

func (li *ListItem) layoutContent(gtx layout.Context, th *theme.Theme, contentColor color.NRGBA) layout.Dimensions {
	return li.content.Layout(gtx, th, contentColor)
}
