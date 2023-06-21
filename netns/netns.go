//go:build linux && amd64

package netns

import (
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
	"syscall"
)

const (
	nr_pidfd_open = 434
	nr_setns      = 308
)

type NetNS struct {
	l    sync.Mutex
	orig *os.File
	fd   uintptr
}

func Open(pid int) (*NetNS, error) {
	orig, err := os.Open("/proc/self/ns/net")
	if err != nil {
		return nil, err
	}

	fd, _, errno := syscall.Syscall(nr_pidfd_open, uintptr(pid), 0, 0)

	if errno != 0 {
		return nil, errno
	}

	return &NetNS{orig: orig, fd: fd}, nil
}

func (n *NetNS) Close() error {
	n.l.Lock()
	defer n.l.Unlock()

	err := errors.Join(
		n.orig.Close(),
		syscall.Close(int(n.fd)),
	)

	n.fd = 0
	n.orig = nil

	return err
}

func (n *NetNS) Exec(fn func()) error {
	n.l.Lock()
	defer n.l.Unlock()

	if n.orig == nil || n.orig.Fd() < 1 {
		return errors.New("original namespace missing")
	}

	if n.fd < 1 {
		return errors.New("target namespace missing")
	}

	// lock ourselves to one thread, as only one thread is moved to the network namespace
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// switch to the target namespace
	_, _, errno := syscall.Syscall(nr_setns, uintptr(n.fd), syscall.CLONE_NEWNET, 0)
	if errno != 0 {
		return errno
	}
	var errs []error

	defer func() {
		_, _, errno = syscall.Syscall(nr_setns, n.orig.Fd(), syscall.CLONE_NEWNET, 0)
		if errno != 0 {
			errs = append(errs, errno)
		}
	}()

	defer func() {
		r := recover()
		if r != nil {
			errs = append(errs, fmt.Errorf("caught panic: %v", r))
		}
	}()

	fn()

	return errors.Join(errs...)
}

func (n *NetNS) Listen(network, address string) (listener net.Listener, err error) {
	n.Exec(func() {
		listener, err = net.Listen(network, address)
	})
	return
}

func (n *NetNS) Interfaces() (i []net.Interface, e error) {
	n.Exec(func() {
		i, e = net.Interfaces()
	})
	return
}
