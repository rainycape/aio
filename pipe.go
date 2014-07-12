// +build !windows

package aio

import (
	"os"
	"syscall"
)

type PipeFlag int

const (
	ReadNonBlock PipeFlag = 1 << iota
	WriteNonBlock
)

// Pipe returns a connected pair of Files; reads from
// r return bytes written to w. It returns the files and an error, if any.
// Optionally, r or w might be set to non-blocking mode using the appropriate
// flags. To obtain a blocking pipe just pass 0 as the flag.
func Pipe(flag PipeFlag) (r *os.File, w *os.File, err error) {
	var p [2]int

	syscall.ForkLock.RLock()
	if err := syscall.Pipe(p[:]); err != nil {
		syscall.ForkLock.RUnlock()
		return nil, nil, os.NewSyscallError("pipe", err)
	}
	syscall.CloseOnExec(p[0])
	syscall.CloseOnExec(p[1])
	if flag&ReadNonBlock != 0 {
		syscall.SetNonblock(p[0], true)
	}
	if flag&WriteNonBlock != 0 {
		syscall.SetNonblock(p[1], true)
	}
	syscall.ForkLock.RUnlock()

	return os.NewFile(uintptr(p[0]), "|0"), os.NewFile(uintptr(p[1]), "|1"), nil
}
