package mmap

import (
	"fmt"
	"os"
	"sync"
	"syscall"
)

type MmapReader struct {
	data  []byte
	close sync.RWMutex
}

type MmapWriter struct {
	MmapReader
	write sync.RWMutex
	path  string
}

func NewReader(path string) (*MmapReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("mmap NewReader: %q %s", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("mmap NewReader: %q %s", path, err)
	}

	mutex := sync.RWMutex{}

	size := info.Size()
	switch {
	case size < 0:
		return nil, fmt.Errorf("mmap NewReader: %q has negative size %v", path, size)
	case size == 0:
		return &MmapReader{[]byte{}, mutex}, nil
	case size != int64(int(size)):
		return nil, fmt.Errorf("mmap NewReader: %q size is too large %v", path, size)
	}

	data, err := syscall.Mmap(int(file.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("mmap NewReader: %q %s", path, err)
	}

	return &MmapReader{
		data:  data,
		close: mutex,
	}, nil
}

func NewWriter(path string) (*MmapWriter, error) {
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
		size = int64(pageSize)
		if err := file.Truncate(size); err != nil {
			return nil, fmt.Errorf("mmap: NewWriter %q %s", path, err)
		}
	}

	switch {
	case size < 0:
		return nil, fmt.Errorf("mmap: NewWriter %q has negative size %v", path, size)
	case size == 0:
		return &MmapWriter{}, nil
	case size != int64(int(size)):
		return nil, fmt.Errorf("mmap: NewWriter %q size is too large %v", path, size)
	}

	data, err := syscall.Mmap(int(file.Fd()), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("mmap: NewWriter %q %s", path, err)
	}

	return &MmapWriter{
		MmapReader{
			data:  data,
			close: sync.RWMutex{},
		},
		sync.RWMutex{},
		path,
	}, nil
}
