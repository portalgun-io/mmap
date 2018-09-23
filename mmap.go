package mmap

import (
	"os"
	"sync"

	errorpkg "github.com/go-util/errors"
)

const (
	// Exactly one of O_RDONLY or O_RDWR must be specified.
	O_RDONLY int = os.O_RDONLY // open the file read-only.
	O_RDWR   int = os.O_RDWR   // open the file read-write.
	// The remaining values may be or'ed in to control behavior.
	O_CREATE int = os.O_CREATE // create a new file if none exists.
	O_EXCL   int = os.O_EXCL   // used with O_CREATE, file must not exist.
	O_SYNC   int = os.O_SYNC   // call Sync() after each write.
	O_TRUNC  int = os.O_TRUNC  // if possible, truncate file when opened.
)

func isSet(flags int, bit int) bool {
	return flags&bit == bit
}

var errors = errorpkg.NewOptions().Caller()

// Map represents a file on disk that has been mapped into memory.
type Map struct {
	sync.RWMutex
	name    string
	data    []byte
	write   bool
	wsync   bool
	id      int
	direct  map[uintptr]Direct
	readers map[int]*Reader
	writers map[int]*Writer
}

// Read opens a file as a read-only memory map.
func Read(name string) (*Map, error) {
	return Open(name, O_RDONLY, 0)
}

// Write opens a file as a writeable memory map. It will create the file if it doesn't exist with
// FileMode 0600.  It does not truncate the file, however if the file size is 0, it will resize
// the file to be size of a memory page as returned by os.Getpagesize(). If a different size is needed,
// call Resize after opening.
func Write(name string) (*Map, error) {
	return Open(name, O_RDWR|O_CREATE, 0600)
}

// Size returns the size of the map.
func (m *Map) Size() int {
	m.RLock()
	defer m.RUnlock()

	return len(m.data)
}

// Name returns the name of the backing file.
func (m *Map) Name() string {
	return m.name
}

// Writeable indicates if the map is writeable.
func (m *Map) Writeable() bool {
	return m.write
}

// WriteSync indicates if the map uses synchronous writes.
func (m *Map) WriteSync() bool {
	return m.wsync
}

// Closed indicates if the map is closed.
func (m *Map) Closed() bool {
	m.RLock()
	defer m.RUnlock()

	return m.data == nil
}
