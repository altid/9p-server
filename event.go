package ubqtlib

import (
	"os"
	"time"
)

type Event struct {
	data   []byte
	client string
	u *Srv
}

func (e *Event) ReadAt(p []byte, off int64) (int, error) {
	// Block on initial read until event occurs
	if off == 0 {
		e.data = <-e.u.event[e.client]
	}
	n := copy(p, e.data[off:])
	return n, nil
}

func (e *Event) WriteAt(p []byte, off int64) (int, error) {
	return len(p), nil
}

func (e *Event) Close() {
	// Remove from event map
	e.u.Lock()
	defer e.u.Unlock()
	delete(e.u.event, e.client)
}

func (e *Event) IsDir() bool { return false }
func (e *Event) ModTime() time.Time { return time.Now() }
func (e *Event) Mode() os.FileMode { return 0444 }
func (e *Event) Name() string { return "event" }
func (e *Event) Size() int64 { return int64(len(e.data)) }
func (e *Event) Sys() interface{} { return nil }
	
func newEvent(u *Srv, client string) *Event {
	// Register to event map
	u.Lock()
	u.event[client] = make(chan []byte)
	u.Unlock()
	return &Event{u: u, client: client}
}
