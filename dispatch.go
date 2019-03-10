package main

import (
	"context"
	"log"
	"path"
	"strings"

	//"aqwari.net/net/styx"
)

/*
	var styxServer styx.Server
	// (bug)halfwit: debug causes reads to the control file to hang on some systems
	if *debug {
		styxServer.TraceLog = log.New(os.Stderr, "", 0)
		styxServer.ErrorLog = log.New(os.Stderr, "", 0)
	}

	styxServer.Addr = ":" + *addr

	// TODO (after dispatch) styxServer.Handler := newServer()
	styxServer.Handler = srv
	//styxServer.Auth = styxauth.Whitelist(rules)


	// ListenAndServe --> err := Serve(l net.Listener)
	// l may be TLS or TCP, set address etc (look at ListenAndServeTLS for example)
	log.Fatal(styxServer.ListenAndServe())
}
*/

type servlist struct {
	servers map[string]*server
}

func dispatchAndServe(events chan string, ctx context.Context) {
	// TODO: context.Context on srv
	// client will match `buffer` of event string to receive the event
	s := &servlist{
		servers: make(map[string]*server),
	}
	for {
		select {
		case <- ctx.Done():
			break
		case e := <-events:
			token := strings.Fields(e)
			switch token[0] {
			case "quit":
				return
			case "new":
				log.Printf("%s", e)
				//startService(token[1], ctx)
			case "closed":
				log.Printf("closed %s", e)
				//stopService(token[1])
			default:
				s.sendEvent(e)
			}
		}
	}
}

func findServer(s *servlist, e string) *server {
	// Strip back the last element of the path until we find a service name
	for dir := path.Dir(e); dir != "."; dir = path.Dir(dir) {
		if s.servers[dir] != nil {
			return s.servers[dir]
		}
	}
	log.Println(path.Dir(e))
	//dirs = path.Dir(e)
	//if s.servers[dirs] != nil {
	//	return s.servers.dir
	//}
	return nil
}

func (s *servlist) sendEvent(e string) {

	srv := findServer(s, e)
	if srv == nil {
		return
	}
	// Range through clients and send events to clients connected to service
	for _, c := range srv.c {
		current := path.Join(path.Base(c.service), path.Base(c.buffer))
		if current == path.Dir(e) {
			// Print just the buffname to the clients' event file
			c.event <- path.Base(e) + "\n"
		}
	}
}
