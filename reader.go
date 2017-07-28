// +build darwin,amd64
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
func (r *Reader) Len() int {
	return len(r.data)
}

// PageCount returns the whole number of pages used by the file
// along with the number of extra bytes at the end of the file.
//
func (r *Reader) PageCount() (int, int) {
	return len(r.data) / pageSize, len(r.data) % pageSize
}

// ReadByteAt returns the byte in the mapped file at the offset
// specified.
//
// If the mmap is closed or the offset is out of range,
// the error will be non-nil.
//
func (r *Reader) ReadByteAt(off int64) (byte, error) {
	if r.data == nil {
		return 0, fmt.Errorf("mmap ReadAtByte: closed")
	}

	r.close.RLock()
	defer r.close.RUnlock()

	if off < 0 || int64(len(r.data)) < off {
		return 0, fmt.Errorf(
			"mmap ReadByteAt: offset %d out of range [0, %d)",
			off, len(r.data),
		)
	}
	return r.data[off], nil
}

// ReadAt reads len(p) bytes into p starting at offset off in the mapped file. It returns the number of bytes read (0 <= n <= len(p)) and any error encountered.
//
// It implements the io.ReaderAt interface.
//
func (r *Reader) ReadAt(p []byte, off int64) (int, error) {
	if r.data == nil {
		return 0, fmt.Errorf("mmap Read: closed")
	}

	r.close.RLock()
	defer r.close.RUnlock()

	if off < 0 || int64(len(r.data)) < off {
		return 0, fmt.Errorf(
			"mmap Read: offset %d out of range [0, %d)",
			off, len(r.data),
		)
	}

	return copy(p, r.data[off:]), nil
}

// Close unmaps the mmap from the underlying file.
//
func (r *Reader) Close() error {
	if r.data == nil {
		return nil
	}

	r.close.Lock()
	defer r.close.Unlock()

	data := r.data
	r.data = nil

	return syscall.Munmap(data)
}

// Closed returns whether the map has been closed.
//
func (r *Reader) Closed() bool {
	return r.data == nil
}
