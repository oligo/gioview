package filetree

import (
	"image"
	"image/color"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/dustin/go-humanize"
	"github.com/oligo/gioview/list"
	"github.com/oligo/gioview/misc"
	"github.com/oligo/gioview/theme"
	gvwidget "github.com/oligo/gioview/widget"
	"github.com/shirou/gopsutil/v4/disk"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type navAction uint8

const (
	noAction navAction = iota
	goBackwardAction
	goForwardAction
	refreshAction
	searchAction
)

type volume struct {
	label      string
	device     string
	mountPoint string
	fsType     string
	opts       []string
}

type favoritesList struct {
	dirs         []string
	labels       []*list.InteractiveLabel
	list         *widget.List
	lastSelected int
}

type locationList struct {
	volumes      []*volume
	labels       []*list.InteractiveLabel
	list         *widget.List
	lastSelected int
}

type entryItem struct {
	click    gesture.Click
	hovering bool
}

type history struct {
	nodes  []*EntryNode
	cursor int
}

type entryViewer struct {
	entryTree   *EntryNode
	pendingNext *EntryNode // to prevent list layout conflicts.
	list        *widget.List
	items       []*entryItem
	//panel
	panel   *entryPanel
	history *history
}

type entryPanel struct {
	forward     widget.Clickable
	backward    widget.Clickable
	refresh     widget.Clickable
	searchInput gvwidget.TextField
}

type FileExplorer struct {
	favorites *favoritesList
	locations *locationList
	viewer    *entryViewer
	history   *history

	resizer *component.Resize
}

var (
	diskIcon, _ = widget.NewIcon(icons.HardwareComputer)
	homeIcon, _ = widget.NewIcon(icons.ActionHome)

	//folderIcon, _     = widget.NewIcon(icons.FileFolder)
	//fileIcon, _       = widget.NewIcon(icons.ActionDescription)
	arrowForwardIcon, _  = widget.NewIcon(icons.NavigationArrowForward)
	arrowBackwardIcon, _ = widget.NewIcon(icons.NavigationArrowBack)
	refreshIcon, _       = widget.NewIcon(icons.NavigationRefresh)
	searchIcon, _        = widget.NewIcon(icons.ActionSearch)
)

// Favorite directories:
var (
	home, _ = os.UserHomeDir()
)

func (v volume) Name() string {
	if v.label != "" {
		return v.label
	}

	return filepath.Base(v.mountPoint)
}

func newEntryViewer(path string, history *history) *entryViewer {
	tree, err := NewFileTree(path, true)
	if err != nil {
		panic(err)
	}
	ev := &entryViewer{
		entryTree: tree,
		list: &widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		panel:   &entryPanel{},
		history: history,
	}

	ev.history.Push(ev.entryTree)

	return ev
}

func NewFileExplorer() *FileExplorer {
	return &FileExplorer{
		history: &history{},
		favorites: &favoritesList{
			dirs: []string{home},
			list: &widget.List{
				List: layout.List{
					Axis: layout.Vertical,
				},
			},
		},
		locations: &locationList{
			list: &widget.List{
				List: layout.List{
					Axis: layout.Vertical,
				},
			},
		},
		resizer: &component.Resize{Axis: layout.Horizontal, Ratio: 0.20},
	}
}

func (exp *FileExplorer) Update(gtx C) {
	if exp.favorites.update(gtx) {
		exp.viewer = newEntryViewer(exp.favorites.dirs[exp.favorites.lastSelected], exp.history)
		exp.locations.lastSelected = -1
	}

	if exp.locations.update(gtx) {
		exp.viewer = newEntryViewer(exp.locations.currentVol().mountPoint, exp.history)
		exp.favorites.lastSelected = -1
	}
}

func (exp *FileExplorer) Layout(gtx C, th *theme.Theme) D {
	exp.Update(gtx)

	return exp.resizer.Layout(gtx,
		func(gtx C) D {
			return exp.layoutPanel(gtx, th)
		},

		func(gtx C) D {
			return exp.layoutBody(gtx, th)
		},

		func(gtx C) D {
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
			divider := misc.Divider(layout.Vertical, unit.Dp(1))
			divider.Inset = layout.Inset{Left: unit.Dp(6), Right: unit.Dp(6)}
			return divider.Layout(gtx, th)
		},
	)
}

func (exp *FileExplorer) layoutPanel(gtx C, th *theme.Theme) D {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return exp.favorites.Layout(gtx, th)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
		layout.Rigid(func(gtx C) D {
			return exp.locations.Layout(gtx, th)
		}),
	)

}

func (exp *FileExplorer) layoutBody(gtx C, th *theme.Theme) D {
	if exp.viewer == nil {
		return material.Label(th.Theme, th.TextSize, "No Files/Folders here.").Layout(gtx)
	}

	return exp.viewer.Layout(gtx, th)
}

func (fav *favoritesList) update(gtx C) bool {
	selected := false
	lastSelected := fav.lastSelected
	for idx, label := range fav.labels {
		if label.Update(gtx) {
			selected = true
			fav.lastSelected = idx
			continue
		}
	}

	if selected {
		if lastSelected >= 0 && lastSelected != fav.lastSelected {
			fav.labels[lastSelected].Unselect()
		}
	} else if fav.lastSelected < 0 {
		for _, label := range fav.labels {
			label.Unselect()
		}
	}

	return selected
}

func (fav *favoritesList) Layout(gtx C, th *theme.Theme) D {
	fav.update(gtx)

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return material.Caption(th.Theme, "Favorites").Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx C) D {
			return material.List(th.Theme, fav.list).Layout(gtx, len(fav.dirs), func(gtx C, index int) D {
				if len(fav.labels) < index+1 {
					fav.labels = append(fav.labels, &list.InteractiveLabel{})
				}

				lb := fav.labels[index]

				return lb.Layout(gtx, th, func(gtx C, textColor color.NRGBA) D {
					return layout.Inset{
						Left:   unit.Dp(12),
						Right:  unit.Dp(12),
						Top:    unit.Dp(4),
						Bottom: unit.Dp(4),
					}.Layout(gtx, func(gtx C) D {
						return layout.Flex{
							Axis:      layout.Horizontal,
							Alignment: layout.Middle,
						}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								iconColor := th.ContrastBg
								if lb.IsSelected() {
									iconColor = th.ContrastFg
								}
								return misc.Icon{Icon: homeIcon, Color: iconColor, Size: unit.Dp(th.TextSize * 1.3)}.Layout(gtx, th)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
							layout.Rigid(func(gtx C) D {
								return material.Label(th.Theme, th.TextSize, filepath.Base(fav.dirs[index])).Layout(gtx)
							}),
						)
					})

				})
			})
		}),
	)

}

func (loc *locationList) Layout(gtx C, th *theme.Theme) D {
	loc.update(gtx)

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return material.Caption(th.Theme, "Locations").Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx C) D {
			return material.List(th.Theme, loc.list).Layout(gtx, len(loc.volumes), func(gtx C, index int) D {
				if len(loc.labels) < index+1 {
					loc.labels = append(loc.labels, &list.InteractiveLabel{})
				}

				lb := loc.labels[index]

				return lb.Layout(gtx, th, func(gtx C, textColor color.NRGBA) D {
					return layout.Inset{
						Left:   unit.Dp(12),
						Right:  unit.Dp(12),
						Top:    unit.Dp(4),
						Bottom: unit.Dp(4),
					}.Layout(gtx, func(gtx C) D {
						return layout.Flex{
							Axis:      layout.Horizontal,
							Alignment: layout.Middle,
						}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								iconColor := th.ContrastBg
								if lb.IsSelected() {
									iconColor = th.ContrastFg
								}
								return misc.Icon{Icon: diskIcon, Color: iconColor, Size: unit.Dp(th.TextSize * 1.3)}.Layout(gtx, th)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
							layout.Rigid(func(gtx C) D {
								return material.Label(th.Theme, th.TextSize, loc.volumes[index].Name()).Layout(gtx)
							}),
						)
					})

				})
			})
		}),
	)

}

func (loc *locationList) update(gtx C) bool {
	if loc.volumes == nil {
		partitions, err := disk.Partitions(false)
		if err != nil {
			panic(err)
		}

		loc.volumes = loc.volumes[:0]
		for _, p := range partitions {
			if slices.Contains(p.Opts, "nobrowse") {
				continue
			}

			vol := &volume{
				device:     p.Device,
				mountPoint: p.Mountpoint,
				fsType:     p.Fstype,
				opts:       p.Opts,
			}
			vol.label, _ = disk.Label(p.Device)
			loc.volumes = append(loc.volumes, vol)
		}
	}

	selected := false
	lastSelected := loc.lastSelected
	for idx, label := range loc.labels {
		if label.Update(gtx) {
			selected = true
			loc.lastSelected = idx
			continue
		}
	}

	if selected {
		if lastSelected >= 0 && lastSelected != loc.lastSelected {
			loc.labels[lastSelected].Unselect()
		}
	} else if loc.lastSelected < 0 {
		for _, label := range loc.labels {
			label.Unselect()
		}
	}

	return selected
}

func (loc *locationList) currentVol() *volume {
	return loc.volumes[loc.lastSelected]
}

func searchFilter(query string) func(info fs.FileInfo) bool {
	return func(info fs.FileInfo) bool {
		if strings.Contains(info.Name(), query) {
			return hiddenFileFilter(info)
		}

		return false
	}
}

func (ev *entryViewer) Update(gtx C) {
	lastTree := ev.entryTree
	if ev.pendingNext != nil {
		ev.entryTree = ev.pendingNext
		ev.history.Push(ev.entryTree)
		ev.pendingNext = nil
	}

	action := ev.panel.Update(gtx)
	switch action {
	case goBackwardAction:
		ev.entryTree = ev.history.Backward()
	case goForwardAction:
		ev.entryTree = ev.history.Forward()
	case refreshAction:
		ev.entryTree.Refresh(hiddenFileFilter)
	case searchAction:
		ev.entryTree.Refresh(searchFilter(strings.TrimSpace(ev.panel.searchInput.Text())))
	default:
		// pass
	}

	// reset if node tree changed
	if lastTree != ev.entryTree || len(ev.entryTree.Children()) != len(lastTree.Children()) {
		ev.list.Position.First = 0
		ev.items = ev.items[:0]
	}
}

func (ev *entryViewer) Layout(gtx C, th *theme.Theme) D {
	ev.Update(gtx)

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Inset{
				Top:    unit.Dp(2),
				Bottom: unit.Dp(6),
			}.Layout(gtx, func(gtx C) D {
				return ev.panel.Layout(gtx, th, ev.entryTree)
			})
		}),
		layout.Rigid(func(gtx C) D {
			return misc.Divider(layout.Horizontal, unit.Dp(1)).Layout(gtx, th)
		}),

		layout.Rigid(func(gtx C) D {
			return ev.layoutEntries(gtx, th)
		}),
	)
}

func (ev *entryViewer) layoutEntries(gtx C, th *theme.Theme) D {
	return material.List(th.Theme, ev.list).Layout(gtx, len(ev.entryTree.Children()), func(gtx C, index int) D {
		return layout.Inset{
			Left:  unit.Dp(4),
			Right: unit.Dp(4),
			Top:   unit.Dp(4),
		}.Layout(gtx, func(gtx C) D {
			entry := ev.entryTree.Children()[index]
			if len(ev.items) < index+1 {
				ev.items = append(ev.items, &entryItem{})
			}
			item := ev.items[index]
			// A folder is double clicked, open it in the explorer.
			if item.Update(gtx) && entry.IsDir() {
				ev.pendingNext = entry
			}

			return item.Layout(gtx, th, entry)
		})
	})
}

// entry viewer panel
func (ep *entryPanel) Layout(gtx C, th *theme.Theme, entry *EntryNode) D {
	ep.Update(gtx)

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return misc.IconButton(th, arrowBackwardIcon, &ep.backward, "Go backward").Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx C) D {
			return misc.IconButton(th, arrowForwardIcon, &ep.forward, "Go forward").Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),

		layout.Flexed(1, func(gtx C) D {
			return material.Label(th.Theme, th.TextSize, entry.Name()).Layout(gtx)
		}),

		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),

		layout.Rigid(func(gtx C) D {
			return misc.IconButton(th, refreshIcon, &ep.refresh, "Refresh the current folder").Layout(gtx)

		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),

		layout.Rigid(func(gtx C) D {
			gtx.Constraints.Max.X = gtx.Dp(unit.Dp(230))
			// gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return layout.Inset{
				Left:  unit.Dp(4),
				Right: unit.Dp(4),
			}.Layout(gtx, func(gtx C) D {
				ep.searchInput.SingleLine = true
				ep.searchInput.LabelOption = gvwidget.LabelOption{Alignment: gvwidget.Hidden}
				ep.searchInput.Padding = unit.Dp(4)
				ep.searchInput.Leading = func(gtx C) D {
					return misc.Icon{Icon: searchIcon, Size: unit.Dp(18), Color: misc.WithAlpha(th.Fg, 0xb0)}.Layout(gtx, th)
				}
				return ep.searchInput.Layout(gtx, th, "Search")
			})
		}),
	)
}

func (ep *entryPanel) Update(gtx C) navAction {
	var action navAction
	if ep.backward.Clicked(gtx) {
		action = goBackwardAction
	}
	if ep.forward.Clicked(gtx) {
		action = goForwardAction
	}
	if ep.refresh.Clicked(gtx) {
		action = refreshAction
	}
	if ep.searchInput.Changed() {
		action = searchAction
	}

	return action
}

// entryItem
func (ei *entryItem) layout(gtx C, th *theme.Theme, entry *EntryNode) D {
	return layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(gtx,
		layout.Flexed(0.6, func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					entryIcon := fileIcon
					if entry.FileInfo.IsDir() {
						entryIcon = folderIcon
					}
					return misc.Icon{Icon: entryIcon, Color: th.ContrastBg, Size: unit.Dp(th.TextSize)}.Layout(gtx, th)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(2)}.Layout),
				layout.Rigid(func(gtx C) D {
					return material.Label(th.Theme, th.TextSize, entry.Name()).Layout(gtx)
				}),
			)
		}),
		layout.Flexed(0.2, func(gtx C) D {
			humanizedSize := "--"

			if !entry.FileInfo.IsDir() {
				humanizedSize = humanize.Bytes(uint64(entry.FileInfo.Size()))
			}
			lb := material.Label(th.Theme, th.TextSize, humanizedSize)
			lb.Color = misc.WithAlpha(th.Fg, 0xb6)
			return lb.Layout(gtx)
		}),
		layout.Flexed(0.2, func(gtx C) D {
			lb := material.Label(th.Theme, th.TextSize, humanize.Time(entry.FileInfo.ModTime()))
			lb.Color = misc.WithAlpha(th.Fg, 0xb6)
			return lb.Layout(gtx)
		}),
	)
}

// Update entryItem states and report whether the item is double clicked.
func (ei *entryItem) Update(gtx C) bool {
	for {
		event, ok := gtx.Event(
			pointer.Filter{Target: ei, Kinds: pointer.Enter | pointer.Leave},
		)
		if !ok {
			break
		}

		switch event := event.(type) {
		case pointer.Event:
			switch event.Kind {
			case pointer.Enter:
				ei.hovering = true
			case pointer.Leave:
				ei.hovering = false
			case pointer.Cancel:
				ei.hovering = false
			}
		}
	}

	var selected bool
	for {
		e, ok := ei.click.Update(gtx.Source)
		if !ok {
			break
		}
		if e.Kind == gesture.KindClick {
			if e.Source == pointer.Mouse && e.NumClicks == 2 {
				selected = true
			} else if e.Source == pointer.Touch && e.NumClicks == 1 {
				selected = true
			}

		}
	}

	return selected
}

func (ei *entryItem) layoutBackground(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	if !ei.hovering {
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}

	rr := gtx.Dp(unit.Dp(4))
	rect := clip.RRect{
		Rect: image.Rectangle{
			Max: image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Min.Y},
		},
		NE: rr,
		SE: rr,
		NW: rr,
		SW: rr,
	}
	paint.FillShape(gtx.Ops, misc.WithAlpha(th.Palette.Fg, th.HoverAlpha), rect.Op(gtx.Ops))
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

func (ei *entryItem) Layout(gtx C, th *theme.Theme, entry *EntryNode) D {
	ei.Update(gtx)

	macro := op.Record(gtx.Ops)
	dims := layout.Background{}.Layout(gtx,
		func(gtx C) D { return ei.layoutBackground(gtx, th) },
		func(gtx C) D { return ei.layout(gtx, th, entry) },
	)

	itemOps := macro.Stop()

	defer pointer.PassOp{}.Push(gtx.Ops).Pop()
	defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()

	ei.click.Add(gtx.Ops)
	event.Op(gtx.Ops, ei)
	itemOps.Add(gtx.Ops)

	return dims
}

func (h *history) Len() int {
	return len(h.nodes)
}

// Push should be called after folder is open.
func (h *history) Push(entry *EntryNode) {
	init := h.nodes == nil

	// remove items behind the cursor and insert the newly pushed entry.
	if h.cursor < len(h.nodes)-1 {
		h.nodes = h.nodes[:h.cursor+1]
	}
	h.nodes = append(h.nodes, entry)
	if !init {
		h.cursor++
	}
}

func (h *history) Forward() *EntryNode {
	h.cursor++
	if h.cursor >= len(h.nodes)-1 {
		h.cursor = len(h.nodes) - 1
	}

	return h.nodes[h.cursor]
}

func (h *history) Backward() *EntryNode {
	h.cursor--
	if h.cursor <= 0 {
		h.cursor = 0
	}

	return h.nodes[h.cursor]
}
