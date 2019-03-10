package main

import (
	"crypto/tls"
	"net"
	"os"
	"path"

	"aqwari.net/net/styx"
	"github.com/google/uuid"
)

type client struct {
	buffer  string
	service string
	event   chan string
	done    chan struct{}
}

type server struct {
	c map[uuid.UUID]*client
	l net.Listener
}

func newServer(addr string) (*server, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	srv := &server{
		c: make(map[uuid.UUID]*client),
		l: l,
	}
	if *useTLS == true {
		tlsConfig := &tls.Config{
			// TODO: Certificates: []tls.Certificate{}
			// Remove this after we implement certs
			InsecureSkipVerify: true,
			ServerName: addr,
		}
		srv.l = tls.NewListener(l, tlsConfig)
	}  
	return srv, nil
}

func (srv *server) newClient(service string) (*client, uuid.UUID) {
	cid := uuid.New()
	buffer := defaultBuffer(service)
	ch := make(chan string)
	done := make(chan struct{})
	srv.c[cid] = &client{
		buffer: buffer,
		service: service,
		event: ch,
		done: done,
	}
	// Make sure we close off events channel when we're done
	go func(ch chan string, done chan struct{}) {
		for {
			defer close(ch)
			select {
			case <- done:
				return
			}
		}
	}(ch, done)
	return srv.c[cid], cid

}

func walkTo(c *client, req string, uid string) (os.FileInfo, string, error) {
	fp := path.Join(c.buffer, req)
	switch req {
	case "/":
		stat, err := os.Stat(c.buffer)
		return stat, fp, err
	case "/ctrl":
		clientCtl := getBase(fp)
		ctlfile, err := mkctl(clientCtl, uid, c)
		if err != nil {
			return nil, fp, err
		}
		return &ctlStat{name: "ctrl", file: ctlfile}, clientCtl, nil
	case "/event":
		clientEvent := getBase(fp)
		eventfile, err := mkevent(uid, c)
		if err != nil {
			return nil, fp, err
		}
		return &eventStat{name: "event", file: eventfile}, clientEvent, nil
	// TODO: case "tabs":
	default:
		stat, err := os.Stat(fp)
		// If we have an error here, try to get a base-level stat.
		if err != nil {
			clientFp := getBase(fp)
			stat, err := os.Stat(clientFp)
			return stat, clientFp, err
		}
		return stat, fp, nil
	}
}

// Called when a client connects
func (srv server) Serve9P(s *styx.Session) {
	// TODO: Server will contain connection address, which will map to services requested
	// Choose the first on the list as a default
	// Server is an aggregate of services based on listen_address
	client, uuid := srv.newClient(path.Join(*inpath, s.Access))
	defer delete(srv.c, uuid)  
	defer close(client.done)
	for s.Next() {
		req := s.Request()
		stat, fp, err := walkTo(client, req.Path(), s.User)
		if err != nil {
			req.Rerror("%s", err)
			continue
		}
		switch t := req.(type) {
		case styx.Twalk:
			t.Rwalk(stat, nil)
		case styx.Topen:
			switch t.Path() {
			case "/":
				t.Ropen(mkdir(fp, s.User, client), nil)
			case "/event":
				t.Ropen(mkevent(s.User, client))
			case "/ctrl":
				t.Ropen(mkctl(fp, s.User, client))
			//case "tabs"
			default:
				t.Ropen(os.OpenFile(fp, os.O_RDWR, 0644))
			}
		case styx.Tstat:
			t.Rstat(stat, nil)
		case styx.Tutimes:
			switch t.Path() {
			case "/", "/event", "/ctrl":
				t.Rutimes(nil)
			default:
				t.Rutimes(os.Chtimes(fp, t.Atime, t.Mtime))
			}
		case styx.Ttruncate:
			switch t.Path() {
			case "/",  "/event", "/ctrl":
				t.Rtruncate(nil)
			default:
				t.Rtruncate(os.Truncate(fp, t.Size))
			}
		// When clients are done with a notification, they delete it. Allow this
		case styx.Tremove:
			switch t.Path() {
			case "/notify":
				t.Rremove(os.Remove(fp))
			default:
				t.Rerror("%s", "permission denied")
			}
		}
	}
}
