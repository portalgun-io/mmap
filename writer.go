package mmap

import (
	"fmt"
	"io"
	"os"
)

func (wm *MmapWriter) WriteByteAt(value byte, off int64) error {
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

func (wm *MmapWriter) WriteAt(p []byte, off int64) (int, error) {
	if wm.data == nil {
		return 0, fmt.Errorf("mmap WriteAt: closed")
	}

	if off < 0 || int64(len(wm.data)) < off {
		return 0, fmt.Errorf("mmap WriteAt: offset %d out of range [0, %d)", off, len(wm.data))
	}

	n := copy(wm.data[off:], p)
	if n < len(p) {
		return n, io.EOF
	}

	return n, nil
}

func (wm *MmapWriter) Region(off int64, ln int64) ([]byte, error) {
	if wm.data == nil {
		return nil, fmt.Errorf("mmap Region: closed")
	}

	if off < 0 || int64(len(wm.data)) < off {
		return nil, fmt.Errorf("mmap Region: offset %d out of range [0, %d)", off, len(wm.data))
	}

	if int64(len(wm.data)) < off+ln {
		return nil, fmt.Errorf("mmap Region: offset + length %d out of range [0, %d)", off+ln, len(wm.data))
	}

	return wm.data[off : off+ln], nil
}

func (wm *MmapWriter) Close() error {
	if wm.data == nil {
		return nil
	}

	wm.write.Lock()
	defer wm.write.Unlock()

	return wm.MmapReader.Close()
}

func (wm *MmapWriter) AddPages(count int) error {
	if count <= 0 {
		return fmt.Errorf("mmap AddPages: count must be greater than 0: %d", count)
	}

	pages, _ := wm.PageCount()

	if err := wm.Close(); err != nil {
		return fmt.Errorf("mmap AddPages: %s", err)
	}

	wm.write.Lock()
	defer wm.write.Unlock()

	size := (int64(pages) + int64(count)) * int64(pageSize)

	if err := resize(wm.path, size); err != nil {
		return fmt.Errorf("mmap AddPages: %s", err)
	}

	writer, err := NewWriter(wm.path)
	if err != nil {
		return fmt.Errorf("mmap AddPages: %s", err)
	}

	wm.data = writer.data

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
