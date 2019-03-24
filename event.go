package main

import (
	"io"
	"os"
	"time"
)

type event struct {
	client *client
	events chan string
	done   chan struct{}
	size   int64
	uid    string
}

func (f *event) Read(p []byte) (n int, err error) {
	f.client.polling = true
	select {
	case <-f.done:
		return 0, io.EOF
	case s := <-f.events:
		n = copy(p, s)
	}
	f.size += int64(n)
	f.client.polling = false
	return n, io.EOF
}

func (f *event) Close() error { 
	return nil 
}

func (f *event) Uid() string  { return f.uid }
func (f *event) Gid() string  { return f.uid }

type eventStat struct {
	name string
	file *event
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
func mkevent(u string, cl *client) (*event, error) {
	e := &event{
		client: cl,
		uid:    u,
		events: cl.event,
		done:   cl.done,
	}
	return e, nil
}
