package ubqtlib

type Event struct {
	s string
	u *Srv
}

func (e *Event) ReadAt(p []byte, off int64) (int, error) {
	// Block on read until event occurs
	buf := <- e.u.event[e.s]
	n := copy(p, buf[off:])
	return n, nil
}

func (e *Event) Close() {
	// Remove from event map
	e.u.Lock()
	defer e.u.Unlock()
	delete(e.u.event, e.s)
}

func newEvent(u *Srv, s string) *Event {
	// Register to event map
	u.Lock()
	defer u.Unlock()
	u.event[s] = make(chan []byte)
	return &Event{u: u, s: s}
}

