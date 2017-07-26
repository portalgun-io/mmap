package mmap

import (
	"fmt"
	"os"
	"sync"
	"syscall"
)

// MmapReader represents a read-only memory mapped file.
//
type MmapReader struct {
	data  []byte
	close sync.RWMutex
}

// MmapWriter represents a read/write memory mapped file.
//
// It includes the methods supported by a MmapWriter.
//
type MmapWriter struct {
	MmapReader
	write sync.RWMutex
	path  string
}

// NewReader takes a file path and returns a MmapReader and an error.
//
// It uses os.Open to open the file and then file.Stat to get information
// about the file. Errors from those calls will be returned.
//
// The file handle does not have to be kept open. The kernel keeps the
// relationship between the mmap and the file on disk until the mmap is
// unmapped.
//
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
		return nil, fmt.Errorf(
			"mmap NewReader: %q has negative size %v",
			path, size,
		)
	case size == 0:
		return &MmapReader{[]byte{}, mutex}, nil
	case size != int64(int(size)):
		return nil, fmt.Errorf(
			"mmap NewReader: %q size is too large %v",
			path, size,
		)
	}

	data, err := syscall.Mmap(
		int(file.Fd()), 0, int(size),
		syscall.PROT_READ, syscall.MAP_SHARED,
	)
	if err != nil {
		return nil, fmt.Errorf("mmap NewReader: %q %s", path, err)
	}

	return &MmapReader{
		data:  data,
		close: mutex,
	}, nil
}

// NewWriter takes a file path and returns a MmapWriter and an error.
//
// It uses os.OpenFile to open the file and then file.Stat to get information
// about the file. The OpenFile options are similar to os.Create except that
// it doesn't truncate the file. Errors from those calls will be returned.
//
// If the file doesn't exist, the file is resized to the result of
// os.Getpagesize(), which is typically 4KB.
//
// The file handle does not have to be kept open. The kernel keeps the
// relationship between the mmap and the file on disk until the mmap is
// unmapped.
//
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
