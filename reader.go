package mmap

import (
	"fmt"
	"os"
	"syscall"
)

var pageSize int = os.Getpagesize()

func PageSize() int {
	return pageSize
}

func (rm *MmapReader) Len() int {
	return len(rm.data)
}

func (rm *MmapReader) PageCount() (int, int) {
	return len(rm.data) / pageSize, len(rm.data) % pageSize
}

func (rm *MmapReader) ReadByteAt(off int64) (byte, error) {
	if rm.data == nil {
		return 0, fmt.Errorf("mmap ReadAtByte: closed")
	}

	rm.close.RLock()
	defer rm.close.RUnlock()

	if off < 0 || int64(len(rm.data)) < off {
		return 0, fmt.Errorf("mmap ReadByteAt: offset %d out of range [0, %d)", off, len(rm.data))
	}
	return rm.data[off], nil
}

func (rm *MmapReader) ReadAt(p []byte, off int64) (int, error) {
	if rm.data == nil {
		return 0, fmt.Errorf("mmap Read: closed")
	}

	rm.close.RLock()
	defer rm.close.RUnlock()

	if off < 0 || int64(len(rm.data)) < off {
		return 0, fmt.Errorf("mmap Read: offset %d out of range [0, %d)", off, len(rm.data))
	}

	return copy(p, rm.data[off:]), nil
}

func (rm *MmapReader) Close() error {
	if rm.data == nil {
		return nil
	}

	rm.close.Lock()
	defer rm.close.Unlock()

	data := rm.data
	rm.data = nil

	return syscall.Munmap(data)
}

func (rm *MmapReader) Closed() bool {
	return rm.data == nil
}
