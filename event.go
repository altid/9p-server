package ubqtlib

import (
	"os"
	"time"
)

type event struct {
	client string
	mtime  time.Time
	wait   chan []byte
	done   chan struct{}
	off    int64
}

// Read in a loop, writing to b as we get more data
func (e *event) Read(b []byte) (n int, err error) {
	for {
		select {
		case msg := <-e.wait:
			copy(b, msg)
		}
	}
	return 0, nil
}

func (e *event) Close() error {
	close(e.done)
	return nil
}

func (e *event) Size() int64        { return e.off }
func (e *event) Name() string       { return "event" }
func (e *event) ModTime() time.Time { return e.mtime }
func (e *event) Mode() os.FileMode  { return 0400 }
func (e *event) IsDir() bool        { return false }
func (e *event) Sys() interface{}   { return nil }

func (u *Srv) readEvent(user string) *event {
	u.event[user] = make(chan []byte)
	done = make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				delete(u.event, user)
			}
		}
	}()
	return &event{client: user, mtime: time.Now(), done: done, wait: u.event[user], off: 0}
}

