package filetree

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
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
	fs.FileInfo
	// parent must be of folder kind.
	Parent   *EntryNode
	children []*EntryNode
}

var isWindows = runtime.GOOS == "windows"

func hiddenFileFilter(info fs.FileInfo) bool {
	name := info.Name()
	if isWindows {
		return !strings.HasPrefix(name, "$") && !strings.HasSuffix(name, ".sys")
	} else {
		return !strings.HasPrefix(name, ".")
	}
}

// Create a new file tree with a relative or absolute rootDir. Folders
// matching prefix in any of the skipPatterns will be skipped.
func NewFileTree(rootDir string, lazyLoad bool) (*EntryNode, error) {
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		log.Fatalln(err)
	}

	if !lazyLoad {
		return loadTree(rootDir)
	}

	st, err := os.Stat(rootDir)
	if err != nil {
		return nil, err
	}
	root := &EntryNode{
		Path:     rootDir,
		Parent:   nil,
		FileInfo: st,
	}

	return root, nil
}

// Load build the tree.
func loadTree(rootDir string) (*EntryNode, error) {
	var root *EntryNode

	// current parent during walk.
	var parent *EntryNode

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("file/folder skipped: ", path)
			return filepath.SkipDir
		}

		// if info.IsDir() {
		// 	for _, prefix := range skipPatterns {
		// 		if strings.HasPrefix(info.Name(), prefix) || strings.HasPrefix(path, prefix) {
		// 			return filepath.SkipDir
		// 		}
		// 	}
		// }

		entry := &EntryNode{
			Path:     filepath.Clean(path),
			FileInfo: info,
		}

		if info.IsDir() {
			if entry.Path == rootDir {
				root = entry
			}
		}

		// find the parent of the current entry:
		if p := findParent(parent, entry); p != nil {
			p.children = append(p.children, entry)
			entry.Parent = p
		}

		if entry.IsDir() {
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

func (n *EntryNode) Kind() NodeKind {
	if n.IsDir() {
		return FolderNode
	}

	return FileNode
}

func (n *EntryNode) Children() []*EntryNode {
	if n.children == nil {
		n.Refresh(hiddenFileFilter)
	}

	return n.children
}

// Add new file or folder.
func (n *EntryNode) AddChild(name string, kind NodeKind) error {
	if name == "" {
		return errors.New("empty file/folder name")
	}

	if err := n.checkDuplicate(name); err != nil {
		return err
	}

	path := filepath.Join(n.Path, name)
	if kind == FileNode {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		file.Close()
	} else if kind == FolderNode {
		if err := os.Mkdir(path, 0755); err != nil {
			return err
		}
	}

	st, _ := os.Stat(path)
	child := &EntryNode{
		Path:     filepath.Clean(path),
		Parent:   n,
		FileInfo: st,
	}

	// insert at the beginning of the children.
	n.children = slices.Insert(n.children, 0, child)
	return nil
}

func (n *EntryNode) checkDuplicate(name string) error {
	filename := filepath.Join(n.Path, name)
	_, err := os.Stat(filename)

	return err
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
		n.Path = filepath.Clean(newPath)
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

	n.Parent.children = slices.DeleteFunc(n.Parent.children, func(en *EntryNode) bool {
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
func (n *EntryNode) Refresh(filterFunc func(entry fs.FileInfo) bool) error {
	if !n.IsDir() {
		return nil
	}

	n.children = n.children[:0]

	err := filepath.Walk(n.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("file/folder skipped: ", path, err)
			return filepath.SkipDir
		}

		if path == n.Path {
			return nil
		}

		// only direct children dir is walked.
		if filepath.Dir(path) != n.Path {
			return filepath.SkipDir
		}

		if filterFunc != nil && filterFunc(info) {
			entry := &EntryNode{
				Path:     filepath.Clean(path),
				FileInfo: info,
				Parent:   n,
			}

			n.children = append(n.children, entry)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// for test purpose
func (n *EntryNode) printTree(depth int) {
	if !n.IsDir() {
		fmt.Printf("+--%s %s\n", strings.Repeat("-", depth), n.Path)

	} else {
		fmt.Printf("|%s \\--- %s\n", strings.Repeat(" ", depth), n.Path)
	}

	for _, child := range n.Children() {
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
