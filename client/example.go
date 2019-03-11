package main

// CURRENTLY PSEUDOCODE

import (
	"net"
)


type msg struct {
	srv string
	msg string
}

type service struct {
	net.Conn
}

func main() {
	servlist := make(map[string]*service)
	events := make(chan *msg)
	for _, service := range argv[n] {
		servlist[service] = attach(service, events)
	}
	go dispatch(events)
	readStdin(servlist)		
}