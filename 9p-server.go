package main

import (
	"flag"
	"log"
	"os"

	"aqwari.net/net/styx"
	//"aqwari.net/net/styx/styxauth"
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
	// Verify our directory exists http//stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-denoted-by-a-path-exists-in-golang
	if _, err := os.Stat(*inpath); os.IsNotExist(err) {
		log.Fatal(err)
	}

	services := watchServiceDir()

	// TODO: This goes down into the service loop
	var styxServer styx.Server
	// (bug)halfwit: debug causes reads to the control file to hang on some systems
	if *debug {
		styxServer.TraceLog = log.New(os.Stderr, "", 0)
		styxServer.ErrorLog = log.New(os.Stderr, "", 0)
	}
	srv := newServer()
	styxServer.Addr = ":" + *addr
	styxServer.Handler = srv
	//styxServer.Auth = styxauth.Whitelist(rules)

	// TODO: listen on a specific IP per connected service, such that we can dial it directly
	// Change this interface, we want a loop for new/deleted service.
	// New:
	// On service start, parse ubqt.cfg for listen_address=, or default. (ie, :564)
	// Check if the address is in our address map
	// If it is, add watcher to events aggregater for given address map item
	// If it is not, start new events tailer, including new server
	// Remove normal inotify based watcher, move to just walking the directory from watch_plan9.go
	// Deleted:
	// remove watch from watch aggregate
	// Watch lists delete their own map entry when they exit

	// TODO: srv.dispatch goes away in favour of dispatching here in the main loop off of servicesw

	go srv.dispatch(services)

	// ListenAndServe --> err := Serve(l net.Listener)
	// l may be TLS or TCP, set address etc (look at ListenAndServeTLS for example)
	log.Fatal(styxServer.ListenAndServe())
}
