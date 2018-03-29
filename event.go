package main

import (
	"fmt"
)

// TODO: Our event fakefile and stat for client-facing FIFO
// TODO: Will need to make this a server method.

func (srv *Server) Dispatch(events chan string) {
	// TODO: Events received will update our clients' content
	for {
		select {
		case e := <-events:
			// Debugging currently
			fmt.Println(e)
		}
	}
}
