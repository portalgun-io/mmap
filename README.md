# mmap

[![GoDoc](https://godoc.org/github.com/go-util/mmap?status.svg)](https://godoc.org/github.com/go-util/mmap)
[![Build Status](https://travis-ci.org/go-util/mmap.svg?branch=master)](https://travis-ci.org/go-util/mmap)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-util/mmap)](https://goreportcard.com/report/github.com/go-util/mmap)
[![Coverage](https://codecov.io/gh/go-util/mmap/branch/master/graph/badge.svg)](https://codecov.io/gh/go-util/mmap)

Package mmap provides an interface to memory mapped files.

Memory maps can be opened as read-only with Read or read-write with Write.
To specify additional flags and a file mode use Open.

There are two different ways to work with memory maps.
They cannot be used simultaneously.
Creating a Direct accessor will fail if any Readers or Writers are open.
Creating Reader or Writer will fail if any Direct accessors are open.

The first is through direct access via a Direct, which is a pointer to a byte slice.
This lets you write directly to the mapped memory, but you will need to manage access
between go routines.

The second is through Readers and Writers.
These implement the standard interfaces from the io package and can be treated like files
while still benefitting from the improved performance of memory mapping.

You can have multiple Readers and Writers.
The map will ensure that writes don't conflict with reads. That is, the underlying map
won't change during the middle of a read.

When the map is resized via Truncate, all open Direct, Reader, and Writer objects are closed.
