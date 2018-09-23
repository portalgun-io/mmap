package mmap

import (
	"io"
)

// Writer reads from and writes to a map. In addition to the methods of mmap.Reader,
// it also implements the following interfaces from the io standard package:
//
//     - Writer          (Write)
//     - WriterAt        (WriteAt)
//     - ByteWriter      (WriteByte)
//     - WriteSeeker     (Write, Seek)
//     - WriteCloser     (Write, Close)
//     - ReadWriter      (Read, Write)
//     - ReadWriteSeeker (Read, Write, Seek)
//     - ReadWriteCloser (Read, Write, Close)
type Writer struct {
	*Reader
}

// Writer returns a new Writer for the map.
func (m *Map) Writer() (*Writer, error) {
	m.Lock()
	defer m.Unlock()

	if m.data == nil {
		return nil, errors.New("mmap closed")
	}

	if !m.write {
		return nil, errors.New("mmap not opened for writing")
	}

	if len(m.direct) > 0 {
		return nil, errors.New("mmap has open direct access pointers")
	}

	id := m.id
	m.id++

	writer := &Writer{
		Reader: &Reader{
			Map: m,
			id:  id,
		},
	}

	m.writers[id] = writer

	return writer, nil
}

// Poke sets the byte at offset.
func (w *Writer) Poke(b byte, offset int) error {
	w.access.RLock()
	defer w.access.RUnlock()

	if w.closed {
		return errors.New("mmap writer closed").Set("name", w.name)
	}

	w.Lock()
	defer w.Unlock()

	if w.data == nil {
		return errors.New("mmap closed").Set("name", w.name)
	}

	if offset < 0 || len(w.data) <= offset {
		return errors.New("offset out of range").Set("name", w.name).
			Set("offset", offset).Set("map_size", len(w.data))
	}

	w.data[offset] = b

	if w.wsync {
		err := msync(w.data, true)
		if err != nil {
			return errors.Wrap(err, "sync error").Set("name", w.name)
		}
	}

	return nil
}

// Write writes len(b) bytes to the File.
// It returns the number of bytes written and an error, if any.
// Write returns io.ErrShortWrite when n < len(b).
func (w *Writer) Write(b []byte) (n int, err error) {
	w.access.Lock()
	defer w.access.Unlock()

	if w.closed {
		return 0, errors.New("mmap writer closed").Set("name", w.name)
	}

	w.Lock()
	defer w.Unlock()

	if w.data == nil {
		return 0, errors.New("mmap closed").Set("name", w.name)
	}

	if len(w.data) <= w.offset {
		return 0, io.EOF
	}

	if len(b) == 0 {
		return 0, nil
	}

	n = copy(w.data[w.offset:], b)
	w.offset += n

	if w.wsync {
		err := msync(w.data, true)
		if err != nil {
			return 0, errors.Wrap(err, "sync error").Set("name", w.name)
		}
	}

	if n < len(b) {
		return n, io.ErrShortWrite
	}

	return n, nil
}

// WriteAt writes len(b) bytes to the File starting at byte offset off.
// It returns the number of bytes written and an error, if any.
// WriteAt returns io.ErrShortWrite when n < len(b).
func (w *Writer) WriteAt(b []byte, offset int64) (n int, err error) {
	w.access.RLock()
	defer w.access.RUnlock()

	if w.closed {
		return 0, errors.New("mmap writer closed").Set("name", w.name)
	}

	w.Lock()
	defer w.Unlock()

	if w.data == nil {
		return 0, errors.New("mmap closed").Set("name", w.name)
	}

	if len(b) == 0 {
		return 0, nil
	}

	if offset < 0 || int64(len(w.data)) <= offset {
		return 0, errors.New("offset out of range").Set("name", w.name).
			Set("offset", offset).Set("map_size", len(w.data))
	}

	n = copy(w.data[offset:], b)

	if w.wsync {
		err := msync(w.data, true)
		if err != nil {
			return 0, errors.Wrap(err, "sync error").Set("name", w.name)
		}
	}

	if n < len(b) {
		return n, io.EOF
	}

	return n, nil
}

// WriteString is like Write, but writes the contents of string s rather than a slice of bytes.
func (w *Writer) WriteString(s string) (n int, err error) {
	w.access.Lock()
	defer w.access.Unlock()

	if w.closed {
		return 0, errors.New("mmap writer closed").Set("name", w.name)
	}

	w.Lock()
	defer w.Unlock()

	if w.data == nil {
		return 0, errors.New("mmap closed").Set("name", w.name)
	}

	if len(w.data) <= w.offset {
		return 0, io.EOF
	}

	if len(s) == 0 {
		return 0, nil
	}

	n = copy(w.data[w.offset:], s)
	w.offset += n

	if w.wsync {
		err := msync(w.data, true)
		if err != nil {
			return 0, errors.Wrap(err, "sync error").Set("name", w.name)
		}
	}

	if n < len(s) {
		return n, io.ErrShortWrite
	}

	return n, nil
}

// WriteByte a byte to the map.
// It returns io.EOF if the writer is at the end of the file.
func (w *Writer) WriteByte(b byte) error {
	w.access.Lock()
	defer w.access.Unlock()

	if w.closed {
		return errors.New("mmap writer closed").Set("name", w.name)
	}

	w.Lock()
	defer w.Unlock()

	if w.data == nil {
		return errors.New("mmap closed").Set("name", w.name)
	}

	if len(w.data) <= w.offset {
		return io.EOF
	}

	w.data[w.offset] = b
	w.offset++

	if w.wsync {
		err := msync(w.data, true)
		if err != nil {
			return errors.Wrap(err, "sync error").Set("name", w.name)
		}
	}

	return nil
}
