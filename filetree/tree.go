package filetree

import (
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

type FileTree struct {
	rootDir string
	root    *EntryNode
	// skip folders matching the prefix
	skipPatterns []string
}

type EntryNode struct {
	Name string
	Path string
	Kind NodeKind
	// parent must be of folder kind.
	Parent   *EntryNode
	Children []*EntryNode

	treeRef *FileTree
}

// Create a new file tree with a relative or absolute rootDir. Folders
// matching prefix in any of the skipPatterns will be skipped.
func NewFileTree(rootDir string, skipPatterns []string) *FileTree {
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		log.Fatalln(err)
	}

	return &FileTree{rootDir: rootDir, skipPatterns: skipPatterns}
}

// Load build the tree.
func (ft *FileTree) Load() error {
	var parent *EntryNode

	return filepath.Walk(ft.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			for _, prefix := range ft.skipPatterns {
				if strings.HasPrefix(info.Name(), prefix) {
					return filepath.SkipDir
				}
			}
		}

		entry := &EntryNode{
			treeRef: ft,
			Name:    info.Name(),
			Path:    filepath.Clean(path),
		}

		if info.IsDir() {
			entry.Kind = FolderNode
			if entry.Path == ft.rootDir {
				ft.root = entry
			}
		} else {
			entry.Kind = FileNode
		}

		// find the parent of the current entry:
		if p := findParent(parent, entry); p != nil {
			p.Children = append(p.Children, entry)
			entry.Parent = p
		} else {
			log.Println("no parent found!!!", entry.Name)
		}

		if entry.Kind == FolderNode {
			// update the current parent to this folder
			parent = entry
		}

		return nil
	})
}

// Print the entire tree in console.
func (ft *FileTree) Print() {
	ft.root.Print(0)
}

// what type of the file, e.g, pdf, png。 This is for
// file nodes only
func (n *EntryNode) FileType() string {
	return filepath.Ext(n.Name)
}

// Refresh reload child entries of the current entry node
func (n *EntryNode) Refresh() error {
	entries, err := os.ReadDir(n.Path)
	if err != nil {
		return err
	}

	entries = slices.DeleteFunc[[]fs.DirEntry, fs.DirEntry](entries, func(info fs.DirEntry) bool {
		if info.IsDir() {
			for _, prefix := range n.treeRef.skipPatterns {
				if strings.HasPrefix(info.Name(), prefix) {
					return true
				}
			}
		}

		return false 
	})
	

	entryMap := make(map[string]*EntryNode, len(n.Children))
	for _, entr := range n.Children {
		entryMap[entr.Name] = entr
	}

	// update existing or add new entries
	for _, entr := range entries {
		
		newKind := FileNode
		if entr.IsDir() {
			newKind = FolderNode
		}

		old, exist := entryMap[entr.Name()]
		if exist {
			if old.Kind != newKind {
				old.Kind = newKind
			}
			continue
		}

		n.Children = append(n.Children, &EntryNode{
			Name:   entr.Name(),
			Path:   filepath.Join(n.Path, entr.Name()),
			Kind:   newKind,
			Parent: n,
		})
	}

	// remove deleted entries
	newEntry := make(map[string]struct{}, len(entries))
	for _, entr := range entries {
		newEntry[entr.Name()] = struct{}{}
	}

	n.Children = slices.DeleteFunc(n.Children, func(en *EntryNode) bool {
		_, exists := newEntry[en.Name]
		return !exists
	})

	return nil
}

// for test purpose
func (n *EntryNode) Print(depth int) {
	if n.Kind == FolderNode {
		fmt.Printf("+--%s %s\n", strings.Repeat("-", depth), n.Path)

	} else {
		fmt.Printf("|%s \\--- %s\n", strings.Repeat(" ", depth), n.Path)
	}

	for _, child := range n.Children {
		child.Print(depth + 1)
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
