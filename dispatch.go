package main

import (
	"context"
	"path"
	"path/filepath"
	"strings"

	"aqwari.net/net/styx"
	"aqwari.net/net/styx/styxauth"
)

type servlist struct {
	servers map[string]*server
}

func dispatchAndServe(ctx context.Context, events chan string) {
	s := &servlist{
		servers: make(map[string]*server),
	}
	for {
		select {
		case <-ctx.Done():
			break
		case e := <-events:
			sendEvent(ctx, s, e)
		}
	}
}

func (sl *servlist) startService(ctx context.Context, service string) {
	addr := findListenAddress(service)
	if sl.servers[addr] != nil { // Server already exists
		return
	}
	srv, err := newServer(addr, service)
	if err != nil {
		return
	}
	var auth styx.AuthFunc
	if *useTLS {
		auth = styxauth.TLSSubjectCN
	}
	styx := styx.Server{
		Addr:    addr,
		Handler: srv,
		Auth:    auth,
	}
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

func sendEvent(ctx context.Context, s *servlist, e string) {
	token := strings.Fields(e)
	switch token[0] {
	case "quit":
		return
	case "new":
		s.startService(ctx, token[1])
		return
	case "closed":
		s.stopService(token[1])
		return
	}
	srv := findServer(s, e)
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
