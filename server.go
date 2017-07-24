package main

import (
	"github.com/mortdeus/go9p/srv"
)

// Init work for fakefiles
// Init work for clients and server
func init() {

}

// Serve - Listen and serve client connections
func Serve() {
	// After data is initialized for fakefiles, start listening on port for connections
	// Check if connection exists in map, add to map so we can track otherwise.
	err = srv.StartNetListener("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
}
