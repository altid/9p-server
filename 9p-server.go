package main

import (
	"context"
	"flag"
	"log"
	"os"
)

var (
	addr     = flag.String("a", "4567", "port to listen on")
	inpath   = flag.String("p", "/tmp/ubqt", "directory to watch (default /tmp/ubqt)")
	debug    = flag.Bool("d", false, "Enable debugging output")
	username = flag.String("u", "", "user name")
)

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(0)
	}
	// Verify our directory exists https://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-denoted-by-a-path-exists-in-golang
	if _, err := os.Stat(*inpath); os.IsNotExist(err) {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	watcher, events := dirWatch()
	go watcher.start(events, ctx)
	dispatchAndServe(events, ctx)
}
