package ubqtlib

import (
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

func (e *Event) Close() {
	// Remove from event map
	e.u.Lock()
	defer e.u.Unlock()
	delete(e.u.event, e.client)
}

func newEvent(u *Srv, client string) *Event {
	// Register to event map
	u.Lock()
	u.event[client] = make(chan []byte)
	u.Unlock()
	return &Event{u: u, client: client}
}
