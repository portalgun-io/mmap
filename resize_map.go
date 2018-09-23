package mmap

import (
	"os"
)

// Truncate resizes the backing file and the memory map to the requested size.
// Any open Direct, Readers, and Writers are closed.
func (m *Map) Truncate(size int64) error {
	if !m.write {
		return errors.New("cannot truncate read-only map").Set("name", m.name)
	}

	if size < 1 {
		return errors.New("size must be greater than zero").Set("name", m.name)
	}

	m.Lock()
	defer m.Unlock()

	if m.data == nil {
		return errors.New("mmap closed")
	}

	m.closeDirects()
	m.closeWriters()
	m.closeReaders()

	err := m.sync(true)
	if err != nil {
		return errors.Wrap(err, "could not sync before truncate").Set("name", m.name)
	}

	err = m.unmap()
	if err != nil {
		return errors.Wrap(err, "could not unmap map before truncate").Set("name", m.name)
	}

	file, err := os.OpenFile(m.name, O_RDWR, 0)
	if err != nil {
		return errors.Wrap(err, "error opening file for truncate").
			Set("name", m.name).Set("size", size)
	}
	defer file.Close()

	err = file.Truncate(int64(size))
	if err != nil {
		return errors.Wrap(err, "error truncating file").
			Set("name", m.name).Set("size", size)
	}

	data, err := mmap(file.Fd(), int(size), true)
	if err != nil {
		return errors.New("could not mmap file after truncate").Set("name", m.name)
	}

	m.data = data
	m.id = 0

	return nil
}
