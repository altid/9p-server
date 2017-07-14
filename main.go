package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
)

var (
	addr   = flag.String("a", ":4567", "port to listen on")
	inpath = flag.String("d", path.Join(os.Getenv("HOME"), "ubqt"), "directory to watch")
)

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(0)
	}

	// Verify our directory exists https://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-denoted-by-a-path-exists-in-golang
	_, err := os.Stat(*inpath)
	if err != nil {
		// TODO: Log fatal error
		log.Fatalf("directory does not exist: %s\n", *inpath)
	}
	events := Watch()
	// Intercept events, updating synthesized files for a client and finally write to their event file.
	for {
		select {
		case line := <-events:
			fmt.Printf("event %s\n", line)
		}
	}
	// Create a default connection
	// default.buffer = DefaultFile()
	// Create our fakefile, send it read/write requests
	// Serve up and listen on port
	// Client connect gets default.buffer
}
