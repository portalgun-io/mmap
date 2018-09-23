package mmap

// Sync flushes all changes to the map out to the backing file.
func (m *Map) Sync(wait bool) error {
	if !m.write {
		return errors.New("cannot sync read-only map").Set("name", m.name)
	}

	m.Lock()
	defer m.Unlock()

	err := m.sync(wait)
	if err != nil {
		return errors.Wrap(err, "error syncing map").Set("name", m.name)
	}

	return nil
}

// sync assumes the map is locked and writeable.
func (m *Map) sync(wait bool) error {
	if m.data == nil {
		return errors.New("mmap closed")
	}
	return msync(m.data, wait)
}
