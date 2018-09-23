// +build darwin linux
package mmap

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

var (
	errEAGAIN error = unix.EAGAIN
	errEINVAL error = unix.EINVAL
	errENOENT error = unix.ENOENT
)

func errnoErr(e unix.Errno) error {
	switch e {
	case 0:
		return nil
	case unix.EAGAIN:
		return errEAGAIN
	case unix.EINVAL:
		return errEINVAL
	case unix.ENOENT:
		return errENOENT
	default:
		return e
	}
}

func mmap(fd uintptr, size int, write bool) ([]byte, error) {
	if size <= 0 {
		return nil, errEINVAL
	}

	var (
		prot int = unix.PROT_READ
		flag int = unix.MAP_SHARED
	)

	if write {
		prot = prot | unix.PROT_WRITE
	}

	r0, _, e1 := unix.Syscall6(unix.SYS_MMAP, 0, uintptr(size), uintptr(prot), uintptr(flag), fd, 0)
	if e1 != 0 {
		return nil, errnoErr(e1)
	}

	addr := uintptr(r0)

	slice := struct {
		addr uintptr
		len  int
		cap  int
	}{addr, size, size}

	data := *(*[]byte)(unsafe.Pointer(&slice))

	return data, nil
}

func munmap(data []byte) error {
	if len(data) == 0 || len(data) != cap(data) {
		return errEINVAL
	}

	addr := uintptr(unsafe.Pointer(&data[0]))
	size := uintptr(len(data))

	_, _, e1 := unix.Syscall(unix.SYS_MUNMAP, addr, size, 0)
	if e1 != 0 {
		return errnoErr(e1)
	}

	return nil
}

func msync(data []byte, wait bool) error {
	flags := unix.MS_INVALIDATE
	if wait {
		flags = flags | unix.MS_SYNC
	} else {
		flags = flags | unix.MS_ASYNC
	}
	return unix.Msync(data, flags)
}
