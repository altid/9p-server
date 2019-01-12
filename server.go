package main

import (
	"os"
	"path"

	"aqwari.net/net/styx"
	"github.com/google/uuid"
)

type Client struct {
	buffer  string
	service string
	event   chan string
	done    chan struct{}
}

type Server struct {
	client map[uuid.UUID]*Client
}

func NewServer() *Server {
	return &Server{client: make(map[uuid.UUID]*Client)}
}

func (srv *Server) newClient(service string) (*Client, uuid.UUID) {
	cid := uuid.New()
	buffer := DefaultBuffer(service)
	ch := make(chan string)
	done := make(chan struct{})
	srv.client[cid] = &Client{
		buffer: buffer,
		service: service,
		event: ch,
		done: done,
	}
	go func(c chan string, done chan struct{}) {
		for {
			defer close(c)
			select {
			case <- done:
				return
			}
		}
	}(ch, done)
	return srv.client[cid], cid

}

// Get a useful stat for the requested path
func walkTo(c *Client, req string, uid string) (os.FileInfo, string, error) {
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

// Main server loop
func (srv *Server) Serve9P(s *styx.Session) {
	// TODO: listen on path that maps to IP the request came from
	client, uuid := srv.newClient(path.Join(*inpath, s.Access))
	defer delete(srv.client, uuid)  
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
		// Clients have the ability to remove notifications
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
