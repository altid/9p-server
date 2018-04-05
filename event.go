package main

import (
	"fmt"
)

// TODO: Our event fakefile and stat for client-facing FIFO
type Event struct {
	
}

// Retern an event type
// See if we need access to an underlying channel here for the type.
func mkevent(client Client) (*Event, error) { 
	return &Event{}, nil
}

func (srv *Server) Dispatch(events chan string) {
	// TODO: Events received will update our clients' content
	// TODO: Test event string against client list
	// client will match `buffer` of event string to receive the event
	for {
		select {
		case e := <-events:
			// Debugging currently
			fmt.Println(e)
		}
	}
}
