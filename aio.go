// Package aio implements asynchronous style IO for Go.
package aio

import (
	"errors"

	"sync"
	"time"
)

var (
	errAlreadyRunning = errors.New("already running")
	errNotRunning     = errors.New("not running")
)

type Flags int

const (
	In Flags = 1 << iota
	Out
	OneShot
)

type Event struct {
	File  interface{}
	Flags Flags
	Data  interface{}
}

type Handler func(ev *Event)

type fileData struct {
	file    interface{}
	data    interface{}
	handler Handler
}

type event struct {
	fd    int
	flags Flags
}

type Set struct {
	mu     sync.RWMutex
	poller poller
	data   map[int]*fileData
	stop   chan struct{}
}

func New() (*Set, error) {
	p, err := newPoller()
	if err != nil {
		return nil, err
	}
	return &Set{poller: p, data: make(map[int]*fileData)}, nil
}

func (s *Set) Add(file interface{}, flags Flags, data interface{}, handler Handler) error {
	fd, err := Fd(file)
	if err != nil {
		return err
	}
	if err := s.poller.Add(fd, flags); err != nil {
		return nil
	}
	s.mu.Lock()
	s.data[fd] = &fileData{
		file:    file,
		data:    data,
		handler: handler,
	}
	s.mu.Unlock()
	return nil
}

func (s *Set) Delete(file interface{}) error {
	fd, err := Fd(file)
	if err != nil {
		return err
	}
	if err := s.poller.Delete(fd); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.data, fd)
	s.mu.Unlock()
	return nil
}

func (s *Set) Wait(duration time.Duration) ([]*Event, error) {
	ev, err := s.poller.Wait(duration)
	if err != nil {
		return nil, err
	}
	events := make([]*Event, len(ev))
	for ii, v := range ev {
		data := s.data[v.fd]
		e := &Event{
			File:  data.file,
			Flags: v.flags,
			Data:  data.data,
		}
		events[ii] = e
		if data.handler != nil {
			data.handler(e)
		}
	}
	return events, nil
}

func (s *Set) Run() error {
	s.mu.Lock()
	if s.stop != nil {
		s.mu.Unlock()
		return errAlreadyRunning
	}
	s.stop = make(chan struct{}, 1)
	s.mu.Unlock()
loop:
	for {
		select {
		case <-s.stop:
			break loop
		default:
			break
		}
		s.Wait(100 * time.Millisecond)
	}
	return nil
}

func (s *Set) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stop == nil {
		return errNotRunning
	}
	s.stop <- struct{}{}
	return nil
}
