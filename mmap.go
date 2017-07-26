package mmap

import (
	"io"
	"os"
)

type MmapReader interface {
	Len() int
	PageCount() (int, int)
	ReadByteAt(off int64) (byte, error)
	io.ReaderAt
	io.Closer
	Closed() bool
}

type MmapWriter interface {
	MmapReader
	WriteByteAt(value byte, off int64) error
	io.WriterAt
	Sync() error
	AddPages(count int) error
}

var pageSize int = os.Getpagesize()

func PageSize() int {
	return pageSize
}
