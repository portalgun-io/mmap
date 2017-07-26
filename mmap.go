package mmap

import (
	"io"
)

type MmapReader interface {
	Len() int
	ReadByteAt(off int64) (byte, error)
	io.ReaderAt
	io.Closer
	Closed() bool
}

type MmapWriter interface {
	MmapReader
	Sync() error
	WriteByteAt(value byte, off int64) error
	io.WriterAt
}
