package main

import (
	"fmt"
	"github.com/mortdeus/go9p/srv"
)

type Ufs struct {
	srv.Srv
	Root string
}

func NewUfs(inpath *string) *Ufs {
	return &Ufs{Root: *inpath}
}

func DispatchEvents(events chan string) {
	// TODO: Events received will update our clients' content
	fmt.Println("We're here")	
	for {
		select {
			case e := <-events:
				fmt.Println(e)
		}
	}
}
