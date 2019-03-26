package main

import (
	"context"
	"flag"
	"log"
	"os"
)

var (
	inpath   = flag.String("p", "/tmp/ubqt", "directory to watch")
	key      = flag.String("k", "/etc/ssl/private/ubqt.pem", "Path to key file for TLS")
	cert     = flag.String("c", "/etc/ssl/certs/ubqt.pem", "Path to cert file for TLS")
	username = flag.String("u", "", "user name")
	useTLS   = flag.Bool("t", false, "Use TLS for connections")
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
	events := make(chan string)
	go startWatcher(ctx, events)
	dispatchAndServe(ctx, events)
}
