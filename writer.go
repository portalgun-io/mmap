package mmap

import (
	"fmt"
	"io"
	"os"
)

// WriteByteAt writes a byte at an offset.
//
func (w *Writer) WriteByteAt(value byte, off int64) error {
	if w.data == nil {
		return fmt.Errorf("mmap WriteByteAt: closed")
	}

	w.write.RLock()
	defer w.write.RUnlock()

	if off < 0 || int64(len(w.data)) < off {
		return fmt.Errorf(
			"mmap WriteByteAt: offset %d out of range [0, %d)",
			off, len(w.data),
		)
	}
	w.data[off] = value
	return nil
}

// WriteAt writes len(p) bytes from p to the underlying data stream at the
// specified offset.
//
// It returns the number of bytes written from p (0 <= n <= len(p)) and any
// error encountered that caused the write to stop early.
//
// It implements the io.WriterAt interface.
//
func (w *Writer) WriteAt(p []byte, off int64) (int, error) {
	if w.data == nil {
		return 0, fmt.Errorf("mmap WriteAt: closed")
	}

	if off < 0 || int64(len(w.data)) < off {
		return 0, fmt.Errorf(
			"mmap WriteAt: offset %d out of range [0, %d)",
			off, len(w.data),
		)
	}

	n := copy(w.data[off:], p)
	if n < len(p) {
		return n, io.EOF
	}

	return n, nil
}

// Region returns a byte slice of the underlying memory mapped file.
//
// The returned byte slice starts at the offset for the length specified.
// Changes to the slice will be made to the underlying file when the
// memory map is flushed to disk.
//
func (w *Writer) Region(off int64, ln int64) ([]byte, error) {
	if w.data == nil {
		return nil, fmt.Errorf("mmap Region: closed")
	}

	if off < 0 || int64(len(w.data)) < off {
		return nil, fmt.Errorf(
			"mmap Region: offset %d out of range [0, %d)",
			off, len(w.data),
		)
	}

	if int64(len(w.data)) < off+ln {
		return nil, fmt.Errorf(
			"mmap Region: offset + length %d out of range [0, %d)",
			off+ln, len(w.data),
		)
	}

	return w.data[off : off+ln], nil
}

// Close unmaps the mmap from the underlying file.
//
func (w *Writer) Close() error {
	if w.data == nil {
		return nil
	}

	w.write.Lock()
	defer w.write.Unlock()

	return w.Reader.Close()
}

// AddPages extends the size of the underlying file by a give number of pages.
//
// Extra bytes that are not part of a whole page are not considered when
// increasing the size. If a file that is 4.5 pages long is extended by
// one page, then the file will be 5 pages long, not 5.5 pages.
//
func (w *Writer) AddPages(count int) error {
	if count <= 0 {
		return fmt.Errorf(
			"mmap AddPages: count must be greater than 0: %d",
			count,
		)
	}

	pages, _ := w.PageCount()

	if err := w.Close(); err != nil {
		return fmt.Errorf("mmap AddPages: %s", err)
	}

	w.write.Lock()
	defer w.write.Unlock()

	size := (int64(pages) + int64(count)) * int64(pageSize)

	if err := resize(w.path, size); err != nil {
		return fmt.Errorf("mmap AddPages: %s", err)
	}

	writer, err := NewWriter(w.path)
	if err != nil {
		return fmt.Errorf("mmap AddPages: %s", err)
	}

	w.data = writer.data

	return nil
}

func resize(path string, size int64) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	err = file.Truncate(size)
	if err != nil {
		return err
	}

	return file.Close()
}
