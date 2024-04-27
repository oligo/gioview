package filetree

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type NodeKind uint8

const (
	FileNode NodeKind = iota
	FolderNode
)

type EntryNode struct {
	Path string
	Kind NodeKind
	// parent must be of folder kind.
	Parent   *EntryNode
	Children []*EntryNode

	// skip folders matching the prefix
	skipPatterns []string
}

// Create a new file tree with a relative or absolute rootDir. Folders
// matching prefix in any of the skipPatterns will be skipped.
func NewFileTree(rootDir string, skipPatterns []string) (*EntryNode, error) {
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		log.Fatalln(err)
	}

	return loadTree(rootDir, skipPatterns)
}

// Load build the tree.
func loadTree(rootDir string, skipPatterns []string) (*EntryNode, error) {
	var root *EntryNode

	// current parent during walk.
	var parent *EntryNode

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			for _, prefix := range skipPatterns {
				if strings.HasPrefix(info.Name(), prefix) {
					return filepath.SkipDir
				}
			}
		}

		entry := &EntryNode{
			Path: filepath.Clean(path),
		}

		if info.IsDir() {
			entry.Kind = FolderNode
			entry.skipPatterns = skipPatterns
			if entry.Path == rootDir {
				root = entry
			}
		} else {
			entry.Kind = FileNode
		}

		// find the parent of the current entry:
		if p := findParent(parent, entry); p != nil {
			p.Children = append(p.Children, entry)
			entry.Parent = p
		}

		if entry.Kind == FolderNode {
			// update the current parent to this folder
			parent = entry
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return root, nil
}

func (n *EntryNode) Name() string {
	return filepath.Base(n.Path)
}

// Add new file or folder.
func (n *EntryNode) AddChild(name string, kind NodeKind) error {
	if name == "" {
		return errors.New("empty file/folder name")
	}

	if err := n.checkDuplicate(name); err != nil {
		return err
	}

	child := &EntryNode{
		Path:         filepath.Join(n.Path, name),
		Kind:         kind,
		Parent:       n,
		skipPatterns: n.skipPatterns,
	}

	if kind == FileNode {
		file, err := os.Create(child.Path)
		if err != nil {
			return err
		}
		file.Close()
	} else if kind == FolderNode {
		if err := os.Mkdir(child.Path, 0755); err != nil {
			return err
		}
	}

	// insert at the beginning of the children.
	n.Children = slices.Insert(n.Children, 0, child)
	return nil
}

func (n *EntryNode) checkDuplicate(name string) error {
	for _, sibling := range n.Children {
		if sibling.Name() == name {
			return errors.New("duplicate file/folder name found")
		}
	}

	return nil
}

// Update set a new name for the current file/folder.
func (n *EntryNode) UpdateName(newName string) error {
	if n.Parent == nil {
		return errors.New("cannot update name of root dir")
	}

	if n.Name() == newName || newName == "" {
		return nil
	}

	if err := n.Parent.checkDuplicate(newName); err != nil {
		return err
	}

	newPath := filepath.Join(filepath.Dir(n.Path), newName)
	defer func() {
		n.Path = newPath
	}()

	return os.Rename(n.Path, newPath)
}

// Delete removes the current file/folders. If onlyEmptyDir is set,
// Delete stops removing non-empty dir if n is a folder node and returns an error.
// TODO: only empty dir is allowed to be removed for now. May add support for
// removing to recyle bin.
func (n *EntryNode) Delete(onlyEmptyDir bool) error {
	if n.Parent == nil {
		return errors.New("cannot update name of root dir")
	}

	// if onlyEmptyDir {
	// 	return os.Remove(n.Path)
	// } else {
	// 	return os.RemoveAll(n.Path)
	// }

	err := os.Remove(n.Path)
	if err != nil {
		return err
	}

	n.Parent.Children = slices.DeleteFunc(n.Parent.Children, func(en *EntryNode) bool {
		return en.Path == n.Path
	})
	return nil
}

// Print the entire tree in console.
func (n *EntryNode) Print() {
	n.printTree(0)
}

// what type of the file, e.g, pdf, png。 This is for
// file nodes only
func (n *EntryNode) FileType() string {
	return filepath.Ext(n.Name())
}

// Refresh reload child entries of the current entry node
func (n *EntryNode) Refresh() error {
	if n.Kind == FileNode {
		return nil
	}

	entries, err := os.ReadDir(n.Path)
	if err != nil {
		return err
	}

	entries = slices.DeleteFunc(entries, func(info fs.DirEntry) bool {
		if info.IsDir() {
			for _, prefix := range n.skipPatterns {
				if strings.HasPrefix(info.Name(), prefix) {
					return true
				}
			}
		}

		return false
	})

	n.Children = n.Children[:0]
	for _, entr := range entries {
		kind := FileNode
		if entr.IsDir() {
			kind = FolderNode
		}

		n.Children = append(n.Children, &EntryNode{
			Path:         filepath.Join(n.Path, entr.Name()),
			Kind:         kind,
			Parent:       n,
			skipPatterns: n.skipPatterns,
		})
	}

	return nil
}

// for test purpose
func (n *EntryNode) printTree(depth int) {
	if n.Kind == FolderNode {
		fmt.Printf("+--%s %s\n", strings.Repeat("-", depth), n.Path)

	} else {
		fmt.Printf("|%s \\--- %s\n", strings.Repeat(" ", depth), n.Path)
	}

	for _, child := range n.Children {
		child.printTree(depth + 1)
	}
}

func findParent(root *EntryNode, child *EntryNode) *EntryNode {
	if root == nil {
		return nil
	}

	if filepath.Dir(child.Path) == root.Path {
		return root
	}

	// trace back to parent's parent
	grandparent := root.Parent
	if grandparent == nil {
		return nil
	}

	return findParent(grandparent, child)
}
