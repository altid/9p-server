package main

import (
	"context"
	"io"
	"log"
	"os"
	"path"
	"time"
)

type event struct {
	events chan string
	done   chan struct{}
	size   int64
	uid    string
}

func (f *event) Read(p []byte) (n int, err error) {
	f.size += int64(len(p))
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

func (f *event) Close() error { return nil }
func (f *event) Uid() string { return f.uid }
func (f *event) Gid() string { return f.uid }

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
	return &event{uid: u, events: cl.event, done: cl.done}, nil
}

func (srv *server) dispatch(events chan string, ctx context.Context) {
	// TODO: context.Context on srv
	// client will match `buffer` of event string to receive the event
	for {
		select {
		case <- ctx.Done():
			break
		case e := <-events:
			if e == "quit" {
				log.Println("quit message received")
				break
			}	
			log.Println(e)
			for _, c := range srv.c {
				// Parse ubqt.cfg, make a map[IP] []servers
				// Find entry based on c's pathings, if server element is nil (which wil be the case until styx starts listening) add the entry and listen, otherwise only dispatch event
				current := path.Join(path.Base(c.service), path.Base(c.buffer))
				if current == path.Dir(e) {
					// Print just the buffname to the clients' event file
					c.event <- path.Base(e) + "\n"
				}	
			}
		}
	}
}
