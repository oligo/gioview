package explorer

import (
	"errors"
	"fmt"
	"io"
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

// A filter is used to decide which files/folders are retained when
// buiding a EntryNode's children. Returning false will remove the current
// entry from from the children.
type EntryFilter func(info fs.FileInfo) bool

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

func searchFilter(query string) func(info fs.FileInfo) bool {
	return func(info fs.FileInfo) bool {
		return strings.Contains(info.Name(), query)
	}
}

func AggregatedFilters(filters ...EntryFilter) EntryFilter {
	if len(filters) <= 0 {
		return nil
	}

	return func(info fs.FileInfo) bool {
		for _, filter := range filters {
			if filter == nil {
				continue
			}

			if !filter(info) {
				return false
			}
		}

		return true
	}
}

// Create a new file tree with a relative or absolute rootDir. A filter is used
// to decide which files/folders are retained.
func NewFileTree(rootDir string) (*EntryNode, error) {
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		log.Fatalln(err)
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
	if !n.IsDir() {
		return nil
	}

	if n.children == nil {
		n.Refresh(hiddenFileFilter)
	}

	return n.children
}

// Add new file or folder.
func (n *EntryNode) AddChild(name string, kind NodeKind) error {
	if !n.IsDir() {
		return nil
	}

	if name == "" {
		return errors.New("empty file/folder name")
	}

	if n.exists(name) {
		return errors.New("duplicated file/folder name")
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

// Copy copies the file at nodePath to the current folder.
// Copy does not replace existing files/folders, instead it returns
// an error indicating that case.
func (n *EntryNode) Copy(nodePath string) error {
	if !n.IsDir() {
		return nil
	}

	if nodePath == "" || !entryExists(nodePath) {
		return errors.New("not a valid entry path")
	}

	if n.exists(filepath.Base(nodePath)) {
		return errors.New("duplicated file/folder name: " + nodePath)
	}

	nodeInfo, _ := os.Stat(nodePath)

	switch nodeInfo.Mode() & os.ModeType {
	case os.ModeDir:
		err := copyDirectory(nodePath, n.Path)
		if err != nil {
			return err
		}
	case os.ModeSymlink:
		if err := copySymLink(nodePath, filepath.Join(n.Path, filepath.Base(nodePath))); err != nil {
			return err
		}
	default:
		if err := copyFile(nodePath, filepath.Join(n.Path, filepath.Base(nodePath))); err != nil {
			return err
		}
	}

	return n.Refresh(hiddenFileFilter)
}

// Move moves the file at nodePath to the current folder.
// Move does not replace existing files/folders, instead it returns
// an error indicating that case.
func (n *EntryNode) Move(nodePath string) error {
	if !n.IsDir() {
		return nil
	}

	if nodePath == "" || !entryExists(nodePath) {
		return errors.New("not a valid entry path")
	}

	if n.exists(filepath.Base(nodePath)) {
		return errors.New("duplicated file/folder name")
	}

	err := os.Rename(nodePath, filepath.Join(n.Path, filepath.Base(nodePath)))
	if err != nil {
		return err
	}

	// if nodePath is a descendant of the root tree, refresh its parent to clean dirty nodes.
	parent := findNodeInTree(n, filepath.Dir(nodePath))
	if parent != nil && parent != n {
		log.Println("hhhh rereshed: ", parent.Path)
		parent.Refresh(hiddenFileFilter)
	}

	return n.Refresh(hiddenFileFilter)
}

func (n *EntryNode) exists(name string) bool {
	filename := filepath.Join(n.Path, name)

	return entryExists(filename)
}

// Update set a new name for the current file/folder.
func (n *EntryNode) UpdateName(newName string) error {
	if n.Parent == nil {
		return errors.New("cannot update name of root dir")
	}

	if n.Name() == newName || newName == "" {
		return nil
	}

	if n.Parent.exists(newName) {
		return errors.New("duplicated file/folder name")
	}

	newPath := filepath.Join(filepath.Dir(n.Path), newName)
	defer func() {
		n.Path = filepath.Clean(newPath)
		st, _ := os.Stat(n.Path)
		n.FileInfo = st
	}()

	return os.Rename(n.Path, newPath)
}

// Delete removes the current file/folders to the system Trash bin.
func (n *EntryNode) Delete() error {
	if n.Parent == nil {
		return errors.New("cannot update name of root dir")
	}

	err := throwToTrash(n.Path)
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
func (n *EntryNode) Refresh(filterFunc EntryFilter) error {
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

func entryExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// copyDirectory copies src dir to dest dir, preserving permissions and ownership.
func copyDirectory(src, dst string) error {
	subdir := filepath.Join(dst, filepath.Base(src))
	if err := createDir(subdir, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(subdir, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := copyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := copySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := copyFile(sourcePath, destPath); err != nil {
				return err
			}
		}

		err = chown(sourcePath, destPath, fileInfo)
		if err != nil {
			return err
		}

		isSymlink := fileInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, fileInfo.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile copies a src file to a dst file where src and dst are regular files.
func copyFile(src, dst string) error {
	srcStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

func createDir(dir string, perm os.FileMode) error {
	if entryExists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

// copySymLink copies a symbolic link from src to dst.
func copySymLink(src, dst string) error {
	link, err := os.Readlink(src)
	if err != nil {
		return err
	}
	return os.Symlink(link, dst)
}

// find longest common path for path1 and path2. path1 and path2 must be
// absolute paths.
func longestCommonPath(path1, path2 string) string {
	if path1 == path2 {
		return path1
	}

	lastSeq := -1
	idx := 0
	for i := 0; i < min(len(path1), len(path2)); i++ {
		if path1[i] != path2[i] {
			break
		}

		idx = i
		if path1[i] == os.PathSeparator {
			lastSeq = i
		}
	}

	if lastSeq < 0 {
		return ""
	}

	if idx > lastSeq && idx == min(len(path1), len(path2))-1 {
		if len(path1) > len(path2) && path1[idx+1] == os.PathSeparator {
			return path1[:idx+1]

		} else if len(path2) > len(path1) && path2[idx+1] == os.PathSeparator {
			return path2[:idx+1]
		}
	}

	// without the trailing seqarator.
	return path1[:lastSeq]

}

// find the node that has its Path equals path.
// The node is an artitary node of a tree.
func findNodeInTree(node *EntryNode, path string) *EntryNode {
	log.Println("running findNodeInTree: ", node.Path, path)

	if path == node.Path {
		return node
	}

	commonPath := longestCommonPath(node.Path, path)
	if commonPath == "" {
		return nil
	}

	commonNode := node
	for {
		if commonNode.Path == commonPath {
			break
		}

		commonNode = commonNode.Parent
		if commonNode == nil {
			return nil
		}
	}

	if commonNode.Path == path {
		return commonNode
	}

	// search decencents
	children := commonNode.children
LOOP:
	for len(children) > 0 {
		for _, child := range children {
			if child.Path == path {
				return child
			}
			if child.Path == node.Path {
				continue
			}
			if len(child.children) > 0 {
				children = child.children
				log.Println("loop: ", child.Path)

				goto LOOP
			}
		}
		break
	}

	return nil

}
