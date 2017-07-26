package mmap

import (
	"fmt"
	"reflect"
	"syscall"
	"unsafe"
)

const syncFlags = uintptr(syscall.MS_SYNC | syscall.MS_INVALIDATE)

// Sync calls Msync to flush buffered writes to disk.
//
func (wm *MmapWriter) Sync() error {
	// Sync is in its own file to limit use of unsafe standard library
	wm.write.Lock()
	defer wm.write.Unlock()

	header := *(*reflect.SliceHeader)(unsafe.Pointer(&wm.data))

	_, _, err := syscall.Syscall(
		syscall.SYS_MSYNC, uintptr(header.Data),
		uintptr(header.Len), syncFlags,
	)
	if err != 0 {
		return fmt.Errorf("mmap Sync: %v", err)
	}

	return nil
}
