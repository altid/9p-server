package main

import (
	"io"
	"os"
	"path"
	"time"
)

type Event struct {
	events chan string
	done   chan struct{}
	size   int64
	uid    string
}

func (f *Event) Read(p []byte) (n int, err error) {
	select {
	case <-f.done:
		return 0, io.EOF
	case s, ok := <-f.events:
		if !ok {
			return 0, io.EOF	
		}
		n = copy(p, s)
	}
	return n, err
}

func (f *Event) Close() error {
	return nil
}

func (f *Event) Uid() string { return f.uid }
func (f *Event) Gid() string { return f.uid }

type eventStat struct {
	name string
	file *Event
}

// Make the size larger than any conceivable message we'll receive
func (s *eventStat) Name() string       { return s.name }
func (s *eventStat) Sys() interface{}   { return s.file }
func (s *eventStat) ModTime() time.Time { return time.Now() }
func (s *eventStat) IsDir() bool        { return false }
func (s *eventStat) Mode() os.FileMode  { return 0444 }
func (s *eventStat) Size() int64        { return s.file.size }

// Return an event type
// See if we need access to an underlying channel here for the type.
func mkevent(u string, client *Client) (*Event, error) {
	return &Event{uid: u, events: client.event, done: client.done}, nil
}

func (srv *Server) Dispatch(events chan string) {
	// TODO: We need to be able to close here as well.
	// TODO: Give each client a `done` channel as well to close on sending side
	// client will match `buffer` of event string to receive the event
	for {
		select {
		case e := <-events:
			for _, c := range srv.client {
				current := path.Join(path.Base(c.service), path.Base(c.buffer))
				if current == path.Dir(e) {
					c.event <- path.Base(e) + "\n"
				}	
			}
		}
	}
}
