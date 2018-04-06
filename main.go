package main

import (
	"flag"
	"log"
	"os"
	"path"

	"aqwari.net/net/styx"
	//"aqwari.net/net/styx/styxauth"
)

var (
	addr     = flag.String("a", "4567", "port to listen on")
	inpath   = flag.String("p", path.Join(os.Getenv("HOME"), "ubqt"), "directory to watch")
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
_, err := os.Stat(*inpath)
	if err != nil {
		log.Fatalf("directory does not exist: %s\n", *inpath)
	}
	var styxServer styx.Server
	if *debug {
		styxServer.TraceLog = log.New(os.Stderr, "", 0)
		styxServer.ErrorLog = log.New(os.Stderr, "", 0)
	}
	srv := NewServer()
	styxServer.Addr = ":"+*addr
	styxServer.Handler = srv
	//styxServer.Auth = styxauth.Whitelist(rules)
	// This will orchestrate events being sent out on all listeners
	events := Watch()
	go srv.Dispatch(events)
	log.Fatal(styxServer.ListenAndServe())
}
