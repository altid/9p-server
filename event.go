package ubqtlib

import (
	"os"
	"time"
)

type event struct {
	client string
	wait   chan []byte
	u      *Srv
}

// Read in a loop, writing to b as we get more data
func (e *event) Read(p []byte) (int, error) {
	c := <-e.wait
	n := copy(p, c)
	return n, nil
}

func (e *event) Close() error {
	e.u.Lock()
	delete(e.u.event, e.client)
	defer e.u.Unlock()
	return nil
}

func (e *event) Size() int64        { return 1 }
func (e *event) Name() string       { return "event" }
func (e *event) ModTime() time.Time { return time.Now() }
func (e *event) Mode() os.FileMode  { return 0400 }
func (e *event) IsDir() bool        { return false }
func (e *event) Sys() interface{}   { return nil }

func (u *Srv) readEvent(user string) *event {
	u.Lock()
	u.event[user] = make(chan []byte)
	defer u.Unlock()
	return &event{u: u, client: user, wait: u.event[user]}
}
