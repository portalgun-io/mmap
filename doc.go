// Package mmap provides access to memory mapped files.
//
// There are two primary types. The Reader is read-only and
// the Writer is read-write.
//
// An example of using a Writer:
//
//     package main
//
//     import (
//         "github.com/go-util/mmap"
//         "log"
//     )
//
//     func main() {
//         mm, err := mmap.NewWriter("mydata.dat")
//         if err != nil {
//             log.Fatal(err)
//         }
//         defer mm.Close()
//
//         data, err := mm.Region(0, int64(mm.Len()))
//         if err != nil {
//             log.Fatal(err)
//         }
//
//         sum := 0
//         for _, value := range data {
//             sum += int(value)
//         }
//
//         log.Printf("Sum Before: %d\n", sum)
//
//         for i := range data {
//             data[i] = byte((sum + i) % 256)
//         }
//
//         sum = 0
//         for _, value := range data {
//             sum += int(value)
//         }
//
//         log.Printf("Sum After: %d\n", sum)
//     }
//
// An example of using a Reader:
//
//     package main
//
//     import (
//         "github.com/go-util/mmap"
//         "log"
//     )
//
//     func reader() {
//         mm, err := mmap.NewReader("mydata.dat")
//         if err != nil {
//             log.Fatal(err)
//         }
//         defer mm.Close()
//
//         data := make([]byte, mm.Len())
//
//         _, err = mm.Reader(data, 0)
//         if err != nil {
//             log.Fatal(err)
//         }
//
//         sum := 0
//         for _, value := range data {
//             sum += int(value)
//         }
//
//         log.Printf("Sum: %d\n", sum)
//     }
//
package mmap
