package main

import (
	"fmt"
)

func DispatchEvents(events chan string) {
	// TODO: Events received will update our clients' content
	for {
		select {
		case e := <-events:
			// Debugging currently
			fmt.Println(e)
		}
	}
}
