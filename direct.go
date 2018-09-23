package mmap

import (
	"unsafe"
)

// Direct is a pointer to a byte slice that directly accesses the memory map.
// Multiple Direct pointers can be created, but only if there are no open
// Readers or Writers on the Map.
//
// As a pointer, the value will have to be dereferenced when used.
//
//     b, _ := memmap.Direct()
//     n := len(*b)
//     for i := range *b {
//             *(b)[i] = i % 256
//     }
//
// Warning: Do not copy or assign the Direct to another variable. The copy
// won't be released when the original is and will become invalid if the
// map is closed or resized.
//
//     b, _ := memmap.Direct()
//     x := *b // Do not do this!
//
type Direct *[]byte

// Direct creates a Direct slice to the entire memory map.
func (m *Map) Direct() (Direct, error) {
	m.Lock()
	defer m.Unlock()

	if m.data == nil {
		return nil, errors.New("mmap closed").Set("name", m.name)
	}

	if len(m.readers) > 0 || len(m.writers) > 0 {
		return nil, errors.New("mmap has open readers and/or writers").Set("name", m.name)
	}

	direct := m.data

	addr := uintptr(unsafe.Pointer(&direct))

	m.direct[addr] = &direct

	return &direct, nil
}

// DirectAt creates a Direct slice to a region of the memory map specified
// by offset and size.
func (m *Map) DirectAt(offset int, size int) (Direct, error) {
	m.Lock()
	defer m.Unlock()

	if m.data == nil {
		return nil, errors.New("mmap closed").Set("name", m.name)
	}

	if len(m.readers) > 0 || len(m.writers) > 0 {
		return nil, errors.New("mmap has open readers and/or writers").Set("name", m.name)
	}

	if offset < 0 || len(m.data) <= offset {
		return nil, errors.New("offset out of range").Set("name", m.name).
			Set("offset", offset).Set("mmap_size", len(m.data))
	}

	if size < 1 {
		return nil, errors.New("size must be greater than zero").Set("name", m.name).
			Set("size", size)
	}

	end := offset + size
	if end > len(m.data) {
		return nil, errors.New("size extends past end of map").Set("name", m.name).
			Set("offset", offset).Set("size", size).
			Set("end", end).Set("mmap_size", len(m.data))
	}

	direct := m.data[offset:end]

	addr := uintptr(unsafe.Pointer(&direct))

	m.direct[addr] = &direct

	return &direct, nil
}

// Free releases a Direct slice
func (m *Map) Free(direct Direct) error {
	if *direct == nil {
		return nil
	}

	m.Lock()
	defer m.Unlock()

	if len(m.direct) == 0 {
		return nil
	}

	addr := uintptr(unsafe.Pointer(direct))

	if _, ok := m.direct[addr]; !ok {
		return errors.New("invalid direct value").Set("name", m.name)
	}

	*direct = nil
	delete(m.direct, addr)

	return nil
}
