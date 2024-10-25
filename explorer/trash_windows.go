package explorer

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// The code in this file are mostly credited to the authors of
// https://github.com/Kei-K23/trashbox/blob/main/trashbox_windows.go.

const (
	FO_DELETE          = 3    // File operation: delete
	FOF_ALLOWUNDO      = 0x40 // Allow to move to Recycle Bin
	FOF_NOCONFIRMATION = 0x10 // No confirmation dialog
)

// SHFILEOPSTRUCT represents the structure used in SHFileOperationW.
type SHFILEOPSTRUCT struct {
	WFunc             uint32
	PFrom             *uint16
	PTo               *uint16
	FFlags            uint16
	AnyOps            bool
	HNameMap          uintptr
	LpszProgressTitle *uint16
}

var (
	shell32              = syscall.NewLazyDLL("shell32.dll")
	procSHFileOperationW = shell32.NewProc("SHFileOperationW")
)

func shFileOperation(op *SHFILEOPSTRUCT) error {
	ret, _, _ := procSHFileOperationW.Call(uintptr(unsafe.Pointer(op)))
	if ret != 0 {
		return fmt.Errorf("failed to move file to Recycle Bin, error code: %d", ret)
	}
	return nil
}

// throwToTrash moves the specified file or directory to the Windows Recycle Bin.
//
// This function takes the path of a file or directory as an argument,
// converts it to an absolute path, and then moves it to the Windows
// Recycle Bin using the Shell API. If the provided path does not
// exist or cannot be accessed, an error will be returned.
//
// The function uses the SHFileOperationW function from the Windows
// Shell API to perform the move operation. It sets the appropriate
// flags to allow undo and suppress confirmation dialogs. If the
// operation is successful, the file or directory will no longer exist
// at the original path and will be relocated to the Recycle Bin for
// potential recovery.
//
// Parameters:
//   - path: The path of the file or directory to be moved to the
//     Recycle Bin.
//
// Returns:
//   - error: Returns nil on success. If an error occurs during the
//     process (e.g., if the file does not exist or the move fails),
//     an error will be returned explaining the reason for failure,
//     including any relevant error codes from the Windows API.
func throwToTrash(path string) error {
	// Get the absolute file path of delete file
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	wPathPtr, err := windows.UTF16PtrFromString(absPath + "\x00")
	if err != nil {
		return err
	}

	op := &SHFILEOPSTRUCT{
		WFunc:  FO_DELETE,
		PFrom:  wPathPtr,
		FFlags: FOF_ALLOWUNDO | FOF_NOCONFIRMATION,
	}

	return shFileOperation(op)
}
