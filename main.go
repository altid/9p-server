package main

import (
	"flag"
	"fmt"
	"path"
	"os"
)

var (
	addr	= flag.String("a", ":4567", "port to listen on")
	inpath	= flag.String("d", path.Join(os.Getenv("HOME"), "ubqt"), "directory to watch")
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
		fmt.Printf("directory does not exist: %s\n", *inpath)
		os.Exit(1)
	}

	// Create a default connection
	// default.buffer = DefaultFile()

	// Create our fakefile, send it read/write requests 
	// Write all events out to event FIFO
	// Serve up and listen on port
	// Client connect gets default.buffer

}
