//go:build !windows
// +build !windows

package explorer

import (
	"fmt"
	"io/fs"
	"os"
	"syscall"
)

func chown(sourcePath, destPath string, srcInfo fs.FileInfo) error {
	stat, ok := srcInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
	}
	if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
		return err
	}

	return nil
}
