package mmap

import (
	"fmt"
	"os"
	"sync"
	"syscall"
)

type readmap struct {
	data  []byte
	close *sync.RWMutex
}

func NewReader(path string) (MmapReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("mmap NewReader: %q %s", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("mmap NewReader: %q %s", path, err)
	}

	mutex := &sync.RWMutex{}

	size := info.Size()
	switch {
	case size < 0:
		return nil, fmt.Errorf("mmap NewReader: %q has negative size %v", path, size)
	case size == 0:
		return &readmap{[]byte{}, mutex}, nil
	case size != int64(int(size)):
		return nil, fmt.Errorf("mmap NewReader: %q size is too large %v", path, size)
	}

	data, err := syscall.Mmap(int(file.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("mmap NewReader: %q %s", path, err)
	}

	return &readmap{
		data:  data,
		close: mutex,
	}, nil
}

func (rm *readmap) Len() int {
	return len(rm.data)
}

func (rm *readmap) PageCount() (int, int) {
	return len(rm.data) / pageSize, len(rm.data) % pageSize
}

func (rm *readmap) ReadByteAt(off int64) (byte, error) {
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

func (rm *readmap) ReadAt(p []byte, off int64) (int, error) {
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

func (rm *readmap) Close() error {
	if rm.data == nil {
		return nil
	}

	rm.close.Lock()
	defer rm.close.Unlock()

	data := rm.data
	rm.data = nil

	return syscall.Munmap(data)
}

func (rm *readmap) Closed() bool {
	return rm.data == nil
}
