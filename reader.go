package mmap

import (
	"fmt"
	"os"
	"syscall"
)

var pageSize int = os.Getpagesize()

// PageSize returns the result of os.Getpagesize()
//
func PageSize() int {
	return pageSize
}

// Len return the size of the memory mapped file.
//
func (rm *Reader) Len() int {
	return len(rm.data)
}

// PageCount returns the whole number of pages used by the file
// along with the number of extra bytes at the end of the file.
//
func (rm *Reader) PageCount() (int, int) {
	return len(rm.data) / pageSize, len(rm.data) % pageSize
}

// ReadByteAt returns the byte in the mapped file at the offset
// specified.
//
// If the mmap is closed or the offset is out of range,
// the error will be non-nil.
//
func (rm *Reader) ReadByteAt(off int64) (byte, error) {
	if rm.data == nil {
		return 0, fmt.Errorf("mmap ReadAtByte: closed")
	}

	rm.close.RLock()
	defer rm.close.RUnlock()

	if off < 0 || int64(len(rm.data)) < off {
		return 0, fmt.Errorf(
			"mmap ReadByteAt: offset %d out of range [0, %d)",
			off, len(rm.data),
		)
	}
	return rm.data[off], nil
}

// ReadAt reads len(p) bytes into p starting at offset off in the mapped file. It returns the number of bytes read (0 <= n <= len(p)) and any error encountered.
//
// It implements the io.ReaderAt interface.
//
func (rm *Reader) ReadAt(p []byte, off int64) (int, error) {
	if rm.data == nil {
		return 0, fmt.Errorf("mmap Read: closed")
	}

	rm.close.RLock()
	defer rm.close.RUnlock()

	if off < 0 || int64(len(rm.data)) < off {
		return 0, fmt.Errorf(
			"mmap Read: offset %d out of range [0, %d)",
			off, len(rm.data),
		)
	}

	return copy(p, rm.data[off:]), nil
}

// Close unmaps the mmap from the underlying file.
//
func (rm *Reader) Close() error {
	if rm.data == nil {
		return nil
	}

	rm.close.Lock()
	defer rm.close.Unlock()

	data := rm.data
	rm.data = nil

	return syscall.Munmap(data)
}

// Closed returns whether the map has been closed.
//
func (rm *Reader) Closed() bool {
	return rm.data == nil
}
