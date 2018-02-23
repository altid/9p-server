package main

import (
	"fmt"
	"github.com/mortdeus/go9p/srv"
)

type Ubqtfs struct {
	srv.Srv
	Root string
}

func NewUbqtfs() *Ubqtfs {
	return &Ubqtfs{Root: *inpath}
}

func DispatchEvents(events chan string) {
	// TODO: Events received will update our clients' content
	for {
		select {
			case e := <-events:
				fmt.Println(e)
		}
	}
}
