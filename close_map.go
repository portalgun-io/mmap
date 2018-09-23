package mmap

// Lock the map before calling closeDirects
func (m *Map) closeDirects() {
	for addr, direct := range m.direct {
		*direct = nil
		delete(m.direct, addr)
	}
}

// Lock map before calling closeReaders
func (m *Map) closeReaders() {
	for _, reader := range m.readers {
		reader.close()
	}
}

// Lock map before calling closeWriters
func (m *Map) closeWriters() {
	for _, writer := range m.writers {
		writer.close()
	}
}

// Close closes the memory map and returns an error if any.
func (m *Map) Close() error {
	m.Lock()
	defer m.Unlock()

	if m.data == nil {
		return nil
	}

	m.closeDirects()
	m.closeWriters()
	m.closeReaders()

	err := m.unmap()
	if err != nil {
		return errors.Wrap(err, "unmap error during close").Set("name", m.name)
	}
	return nil
}

// Lock map and close all direct, writers, and readers before calling.
func (m *Map) unmap() error {
	data := m.data
	m.data = nil

	err := munmap(data)
	if err != nil {
		return errors.Wrap(err, "error unmapping memory").Set("name", m.name)
	}

	return nil
}
