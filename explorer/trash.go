//go:build !windows
// +build !windows

package explorer

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

func throwToTrash(path string) error {
	trashDir, err := getTrashFolder()
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	_, err = os.Stat(absPath)
	if err != nil {
		return err
	}

	trashPath := filepath.Join(trashDir, filepath.Base(path))
	err = os.Rename(absPath, trashPath)
	if err != nil {
		return err
	}

	return nil
}

func getTrashFolder() (string, error) {
	switch runtime.GOOS {
	case "darwin", "ios":
		return appleTrashDir()
	default:
		return unixTrashDir()
	}
}

// According to Freedesktop.org specifications, the "home trash" directory
// is at $XDG_DATA_HOME/Trash, and $XDG_DATA_HOME in turn defaults to $HOME/.local/share.
// Refs: https://specifications.freedesktop.org/basedir-spec/latest/
func unixTrashDir() (string, error) {
	if xdgHome := os.Getenv("XDG_DATA_HOME"); xdgHome != "" {
		trashDir := filepath.Join(xdgHome, "Trash")
		err := checkDirExists(trashDir)
		if err != nil {
			return "", err
		}
		return trashDir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	trashDir := filepath.Join(home, ".local/share/Trash")
	err = checkDirExists(trashDir)
	if err != nil {
		return "", err
	}

	return trashDir, nil
}

func appleTrashDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	trashDir := filepath.Join(dir, ".Trash")
	err = checkDirExists(trashDir)
	if err != nil {
		return "", err
	}

	return trashDir, nil
}

func checkDirExists(dir string) error {
	stat, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(dir, os.ModeDir)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	if !stat.IsDir() {
		return errors.New("Trash dir is not a folder")
	}

	return nil

}
