package explorer

import (
	"gioui.org/widget/material"
	"github.com/oligo/gioview/menu"
	"github.com/oligo/gioview/theme"
	"github.com/oligo/gioview/view"
)

// Default operation for file tree nodes.
// Support file/folder copy, cut, paste, rename, delete and new file/folder creation.
func DefaultFileMenuOptions(vm view.ViewManager) MenuOptionFunc {
	return func(gtx C, item *EntryNavItem) [][]menu.MenuOption {

		common := [][]menu.MenuOption{
			{
				// copy & paste files or folders
				{
					OnClicked: func() error {
						//TODO
						return nil
					},

					Layout: func(gtx C, th *theme.Theme) D {
						return material.Label(th.Theme, th.TextSize, "Copy").Layout(gtx)
					},
				},

				{
					OnClicked: func() error {
						// TODO
						return nil
					},

					Layout: func(gtx C, th *theme.Theme) D {
						return material.Label(th.Theme, th.TextSize, "Cut").Layout(gtx)
					},
				},

				{
					OnClicked: func() error {
						// TODO
						return nil
					},

					Layout: func(gtx C, th *theme.Theme) D {
						return material.Label(th.Theme, th.TextSize, "Paste").Layout(gtx)
					},
				},
			},

			{
				{
					OnClicked: func() error {
						item.StartEditing(gtx)
						return nil
					},

					Layout: func(gtx C, th *theme.Theme) D {
						return material.Label(th.Theme, th.TextSize, "Rename").Layout(gtx)
					},
				},

				{
					OnClicked: func() error {
						return item.Remove()
					},

					Layout: func(gtx C, th *theme.Theme) D {
						return material.Label(th.Theme, th.TextSize, "Delete").Layout(gtx)
					},
				},
			},
		}

		if item.Kind() == FolderNode {
			// create subfolder, files, remove files, rename files
			dirOptions := []menu.MenuOption{
				// create new file in current folder
				{
					OnClicked: func() error {
						return item.CreateChild(gtx, FileNode)
					},

					Layout: func(gtx C, th *theme.Theme) D {
						return material.Label(th.Theme, th.TextSize, "New File").Layout(gtx)
					},
				},

				// create subfolder
				{
					OnClicked: func() error {
						return item.CreateChild(gtx, FolderNode)
					},

					Layout: func(gtx C, th *theme.Theme) D {
						return material.Label(th.Theme, th.TextSize, "New Folder").Layout(gtx)
					},
				},
			}

			dirOptions = append(dirOptions, common[0]...)
			common[0] = dirOptions
		}

		return common
	}
}
