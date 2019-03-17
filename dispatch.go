package main

import (
	"context"
	"log"
	"path"
	"path/filepath"
	"strings"

	"aqwari.net/net/styx"
)

type servlist struct {
	servers map[string]*server
}

func dispatchAndServe(events chan string, ctx context.Context) {
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
				s.startService(token[1], ctx)
			case "closed":
				s.stopService(token[1])
			default:
				s.sendEvent(e)
			}
		}
	}
}

func (sl *servlist) startService(service string, ctx context.Context) {
	addr := findListenAddress(service)
	if sl.servers[addr] != nil { // Server already exists
		return
	}
	srv, err := newServer(addr, service)
	if err != nil {
		return
	}
	styx := styx.Server{
		Addr: addr,
		Handler: srv,
		//Auth: styxauth.TLS,
		//TLSConfig: 
	}
	log.Println("Adding address " + addr)
	go styx.Serve(srv.l)
	sl.servers[addr] = srv
}

func (servlist *servlist) stopService(service string) {
	addr := findListenAddress(service)
	srv := servlist.servers[addr]
	if srv != nil {
		srv.l.Close()
	}
	delete(servlist.servers, addr)
}

func findServer(s *servlist, e string) *server {
	for _, srv := range s.servers {
		testPath := path.Join(*inpath, srv.service)
 		if filepath.HasPrefix(e, testPath) {
			return srv
		}
	}
	return nil
}

func (s *servlist) sendEvent(e string) {
	srv := findServer(s, e)
	if srv == nil {
		return
	}
	// Range through clients and send events to clients connected to service
	for _, c := range srv.c {
		if path.Base(e) == "notification" {
			c.tabs[path.Dir(e)] = "red"
		}
		current := path.Join(*inpath, path.Base(c.service), path.Base(c.buffer))
		if current == path.Dir(e) {
			c.event <- path.Base(e) + "\n"
		}
	}
}
