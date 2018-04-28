package main

import (
	"io"
	"os"
	"time"
	"path"
)

// TODO: Our event fakefile and stat for client-facing FIFO
type Event struct {
	events chan string
	done chan struct{}
	uid string
}

func (f *Event) Read(p []byte) (n int, err error) {
	s, ok := <- f.events
	if ! ok { 
		return 0, io.EOF
	}
	n = copy(p, s)
	return n, nil
}

func (f *Event) Close() error { 
	close(f.done)
	return nil
}
func (f *Event) Uid() string { return f.uid }
func (f *Event) Gid() string { return f.uid }

type eventStat struct {
	name string
	file *Event
}

func (s *eventStat) Name() string { return s.name }
func (s *eventStat) Sys() interface{} { return s.file }
func (s *eventStat) ModTime() time.Time { return time.Now().Truncate(time.Hour) }
func (s *eventStat) IsDir() bool { return false }
func (s *eventStat) Mode() os.FileMode { return os.ModeNamedPipe | 0666 }
func (s *eventStat) Size() int64 { return 0 }

// Return an event type
// See if we need access to an underlying channel here for the type.
func mkevent(u string, client *Client) (*Event, error) { 
	done := make(chan struct{})
	return &Event{uid: u, events: client.event, done: done}, nil
}

func (srv *Server) Dispatch(events chan string) {
	// TODO: We need to be able to close here as well.
	// TODO: Give each client a `done` channel as well to close on sending side
	// client will match `buffer` of event string to receive the event
	for {
		select {
		case e := <-events:
			for _, c := range srv.client {
				switch path.Dir(e) {
				case c.service: // ctl, tabs
					c.event <- path.Base(e) + "\n"
				case c.buffer: // all others
					c.event <- path.Base(e) + "\n"
				}
			}
		}
	}
}
