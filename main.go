package main

import (
	"flag"
	"log"
	"os"
	"path"
)

var (
	addr   = flag.String("a", ":4567", "port to listen on")
	inpath = flag.String("p", path.Join(os.Getenv("HOME"), "ubqt"), "directory to watch")
	debug = flag.Int("d", 0, "debug level (0-3)")	
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
		// TODO: Log fatal error
		log.Fatalf("directory does not exist: %s\n", *inpath)
	}
	//events := Watch()
	// TODO: Map of clients will each recieve event, loop through each client and send event
	//go DispatchEvents()
	s := NewUfs()
	s.Dotu = true
	s.Id = "ubqt"
	s.Debuglevel = *debug
	s.Start(s)
	if *username != "" {
		u := s.Upool.Uname2User(*username)
		if u == nil {
			log.Printf("Warning: Adding %v failed", *username)
		}
	}
	err = s.StartNetListener("tcp", *addr)
	if err != nil {
		log.Fatalf("error starting network listener: %s\n", err)
	}
}

