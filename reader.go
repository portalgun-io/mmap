package mmap

import (
	"io"
	"sync"
)

const (
	SEEK_SET int = 0 // seek relative to the origin of the file
	SEEK_CUR int = 1 // seek relative to the current offset
	SEEK_END int = 2 // seek relative to the end
)

// Reader reads from a map. It implements the following interfaces from the
// io standard package:
//
// - Reader (Read)
// - ReaderAt (ReadAt)
// - ByteReader (ReadByte)
// - Seeker (Seek)
// - Closer (Close)
// - ReadCloser (Read, Close)
// - ReadSeeker (Read, Seek)
type Reader struct {
	*Map
	access sync.RWMutex
	closed bool
	id     int
	offset int
}

// Reader returns a new Reader for the map.
func (m *Map) Reader() (*Reader, error) {
	m.Lock()
	defer m.Unlock()

	if m.data == nil {
		return nil, errors.New("mmap closed")
	}

	if len(m.direct) > 0 {
		return nil, errors.New("mmap has open direct access pointers")
	}

	id := m.id
	m.id++

	reader := &Reader{
		Map: m,
		id:  id,
	}

	m.readers[id] = reader

	return reader, nil
}

// Peek returns the value of the byte at offset.
func (r *Reader) Peek(offset int) (byte, error) {
	r.access.RLock()
	defer r.access.RUnlock()

	if r.closed {
		return 0, errors.New("mmap reader closed").Set("name", r.name)
	}

	r.RLock()
	defer r.RUnlock()

	if r.data == nil {
		return 0, errors.New("mmap closed").Set("name", r.name)
	}

	if offset < 0 || len(r.data) <= offset {
		return 0, errors.New("offset out of range").Set("name", r.name).
			Set("offset", offset).Set("map_size", len(r.data))
	}

	return r.data[offset], nil
}

// Read reads up to len(b) bytes from the map Reader. It returns the number of bytes read
// and any error encountered. At end of map, Read returns 0, io.EOF.
func (r *Reader) Read(b []byte) (n int, err error) {
	r.access.Lock()
	defer r.access.Unlock()

	if r.closed {
		return 0, errors.New("mmap reader closed").Set("name", r.name)
	}

	r.RLock()
	defer r.RUnlock()

	if r.data == nil {
		return 0, errors.New("mmap closed").Set("name", r.name)
	}

	if len(r.data) <= r.offset {
		return 0, io.EOF
	}

	if len(b) == 0 {
		return 0, nil
	}

	n = copy(b, r.data[r.offset:])
	r.offset += n

	return n, nil
}

// ReadAt reads len(b) bytes from the Map starting at offset.
// It returns the number of bytes read and the error, if any.
// ReadAt always returns io.EOF when n < len(b).
func (r *Reader) ReadAt(b []byte, offset int64) (n int, err error) {
	r.access.RLock()
	defer r.access.RUnlock()

	if r.closed {
		return 0, errors.New("mmap reader closed").Set("name", r.name)
	}

	r.RLock()
	defer r.RUnlock()

	if r.data == nil {
		return 0, errors.New("mmap closed").Set("name", r.name)
	}

	if len(b) == 0 {
		return 0, nil
	}

	if offset < 0 || int64(len(r.data)) <= offset {
		return 0, errors.New("offset out of range").Set("name", r.name).
			Set("offset", offset).Set("map_size", len(r.data))
	}

	n = copy(b, r.data[offset:])
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}

// ReadByte reads and returns the next byte from the map.
// It returns io.EOF if the Reader is at the end of the file.
func (r *Reader) ReadByte() (byte, error) {
	r.access.Lock()
	defer r.access.Unlock()

	if r.closed {
		return 0, errors.New("mmap reader closed").Set("name", r.name)
	}

	r.RLock()
	defer r.RUnlock()

	if r.data == nil {
		return 0, errors.New("mmap closed").Set("name", r.name)
	}

	if len(r.data) <= r.offset {
		return 0, io.EOF
	}

	b := r.data[r.offset]
	r.offset++

	return b, nil
}

// Seek sets the offset for the next Read or Write on Reader to offset, interpreted
// according to whence: 0 means relative to the origin of the file, 1 means relative
// to the current offset, and 2 means relative to the end. It returns the new offset
// and an error, if any.
func (r *Reader) Seek(offset int64, whence int) (ret int64, err error) {
	r.access.Lock()
	defer r.access.Unlock()

	if r.closed {
		return 0, errors.New("mmap reader closed").Set("name", r.name)
	}

	r.Lock()
	defer r.Unlock()

	if r.data == nil {
		return 0, errors.New("mmap closed").Set("name", r.name)
	}

	var pos int64

	switch whence {
	case SEEK_SET:
		pos = offset
	case SEEK_CUR:
		pos = int64(r.offset) + offset
	case SEEK_END:
		pos = int64(len(r.data)) + offset
	default:
		return 0, errors.New("invalid whence").Set("name", r.name).Set("whence", whence)
	}

	if pos < 0 || int64(len(r.data)) <= pos {
		return 0, errors.New("invalid position").Set("name", r.name).
			Set("offset", offset).Set("whence", whence).
			Set("map_size", len(r.data)).Set("position", pos)
	}

	if pos != int64(int(pos)) {
		return 0, errors.New("position is invalid for architecture").Set("name", r.name).
			Set("offset", offset).Set("whence", whence).
			Set("map_size", len(r.data)).Set("position", pos)
	}

	r.offset = int(pos)

	return pos, nil
}

// Close closes the Reader.
func (r *Reader) Close() error {
	r.Lock()
	defer r.Unlock()

	r.close()

	return nil
}

// Lock map before calling
func (r *Reader) close() {
	r.access.Lock()
	defer r.access.Unlock()

	r.closed = true
	delete(r.readers, r.id)
}
