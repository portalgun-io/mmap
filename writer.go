package mmap

import (
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
)

type writemap struct {
	readmap
	write *sync.RWMutex
}

func NewWriter(path string) (MmapWriter, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("mmap: NewWriter %q %s", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("mmap: NewWriter %q %s", path, err)
	}

	size := info.Size()
	if size == 0 {
		size = int64(os.Getpagesize())
		if err := file.Truncate(size); err != nil {
			return nil, fmt.Errorf("mmap: NewWriter %q %s", path, err)
		}
	}

	switch {
	case size < 0:
		return nil, fmt.Errorf("mmap: NewWriter %q has negative size %v", path, size)
	case size == 0:
		return &writemap{}, nil
	case size != int64(int(size)):
		return nil, fmt.Errorf("mmap: NewWriter %q size is too large %v", path, size)
	}

	data, err := syscall.Mmap(int(file.Fd()), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("mmap: NewWriter %q %s", path, err)
	}

	return &writemap{
		readmap{
			data:  data,
			close: &sync.RWMutex{},
		},
		&sync.RWMutex{},
	}, nil
}

func (wm *writemap) WriteByteAt(value byte, off int64) error {
	if wm.data == nil {
		return fmt.Errorf("mmap WriteByteAt: closed")
	}

	wm.write.RLock()
	defer wm.write.RUnlock()

	if off < 0 || int64(len(wm.data)) < off {
		return fmt.Errorf("mmap WriteByteAt: offset %d out of range [0, %d)", off, len(wm.data))
	}
	wm.data[off] = value
	return nil
}

func (wm *writemap) WriteAt(p []byte, off int64) (int, error) {
	if wm.data == nil {
		return 0, fmt.Errorf("mmap WriteAt: closed")
	}

	if off < 0 || int64(len(wm.data)) < off {
		return 0, fmt.Errorf("mmap WriteAt: invalid WriteAt offset %d", off)
	}

	n := copy(wm.data[off:], p)
	if n < len(p) {
		return n, io.EOF
	}

	return n, nil
}

func (wm *writemap) Close() error {
	if wm.data == nil {
		return nil
	}

	wm.write.Lock()
	defer wm.write.Unlock()

	return wm.readmap.Close()
}
