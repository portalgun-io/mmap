package mmap

import (
	"os"
)

// Open opens a file as a memory map using the given flags. It does not support the O_WRONLY or
// O_APPEND flags. It will create the file if it doesn't exist. If the file is empty or
// O_TRUNC is specified, the file will be resized to the size of a memory page as
// returned by os.Getpagesize() and the bytes will be zeroed out.
func Open(name string, flags int, mode os.FileMode) (*Map, error) {
	switch {
	case isSet(flags, os.O_WRONLY):
		return nil, errors.New("Map does not support O_WRONLY flag").
			Set("name", name).Set("flags", flags)
	case isSet(flags, os.O_APPEND):
		return nil, errors.New("Map does not support O_APPEND flag").
			Set("name", name).Set("flags", flags)
	}

	var (
		write = isSet(flags, O_RDWR)
		creat = isSet(flags, O_CREATE)
		excl  = isSet(flags, O_EXCL)
		wsync = isSet(flags, O_SYNC)
		trunc = isSet(flags, O_TRUNC)
	)

	switch {
	case creat && !write:
		return nil, errors.New("O_CREATE requires O_RDWR flag").
			Set("name", name).Set("flags", flags)
	case excl && !creat:
		return nil, errors.New("O_EXCL requires O_CREATE flag").
			Set("name", name).Set("flags", flags)
	case wsync && !write:
		return nil, errors.New("O_SYNC requires O_RDWR flag").
			Set("name", name).Set("flags", flags)
	case trunc && !write:
		return nil, errors.New("O_TRUNC requires O_RDWR flag").
			Set("name", name).Set("flags", flags)
	}

	file, err := os.OpenFile(name, flags, mode)
	if err != nil {
		err = errors.Wrap(err, "could not open file").Set("name", name)
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		err = errors.Wrap(err, "could not stat file").Set("name", name)
		return nil, err
	}

	if info.Size() < 1 && !write {
		return nil, errors.New("cannot mmap empty file").Set("name", name)
	}

	if write && (info.Size() < 1 || trunc) {
		size := int64(os.Getpagesize())
		err = file.Truncate(size)
		if err != nil {
			if trunc {
				return nil, errors.New("could not truncate file").Set("name", name)
			} else {
				return nil, errors.New("could not resize new or empty file").Set("name", name)
			}
		}

		info, err = file.Stat()
		if err != nil {
			return nil, errors.Wrap(err, "could not stat after resize").Set("name", name)
		}

		if info.Size() != size {
			return nil, errors.New("incorrect size of resized file").
				Set("name", name).Set("requested_size", size).Set("actual_size", info.Size())
		}
	}

	size := int(info.Size())
	if info.Size() != int64(size) {
		return nil, errors.New("file too large for architecture").Set("name", name).Set("size", info.Size())
	}

	data, err := mmap(file.Fd(), size, write)
	if err != nil {
		return nil, errors.New("could not mmap file").Set("name", name)
	}

	if trunc {
		for i := range data {
			data[i] = 0
		}
		err := msync(data, true)
		if err != nil {
			return nil, errors.Wrap(err, "error syncing mmap after truncate")
		}
	}

	return &Map{
		name:    name,
		data:    data,
		write:   write,
		wsync:   wsync,
		direct:  make(map[uintptr]Direct),
		readers: make(map[int]*Reader),
		writers: make(map[int]*Writer),
	}, nil
}
