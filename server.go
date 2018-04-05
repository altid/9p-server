package main

import (
	"log"
	"os"
	"path"

	"aqwari.net/net/styx"
	"github.com/google/uuid"
)

type Client struct {
	buffer string
	event chan string
}

// List of clients' current buffers, useful for filtering events that each client receives
type Server struct {
	client map[uuid.UUID]*Client
}

func NewServer() *Server {
	return &Server{client: make(map[uuid.UUID]*Client)}
}

func (srv *Server) newClient(service string) *Client {
	cid := uuid.New()
	buffer := DefaultBuffer(service)
	ch := make(chan string)
	srv.client[cid] = &Client{ buffer: buffer, event: ch }
	return srv.client[cid]
}

// Get a useful stat for the requested path
func walkTo(path string) (os.FileInfo, string, error) {
	// We prematurely create a stat type for each file here
	switch path {
	// TODO: Implement `stat` type for event
	case "/ctl":
		base := getBase(path)
		cl, err := os.Stat(base)
		if err != nil {
			return nil, path, err
		}
		ctlfile, err := mkctl(base)
		if err != nil {
			return nil, path, err
		}
		return &ctlStat{name: "ctl", file: ctlfile, stat: cl, }, base, nil
	default:
		stat, err := os.Stat(path)
		// If we have an error here, try to get a base-level stat. 
		if err != nil {
			stat, err = os.Stat(getBase(path))
			if err != nil {
				return nil, path, err
			}
			return stat, getBase(path), nil
		}
		return stat, path, nil
	}
}

// Serve9P is called by styx.ListenAndServe on a client connection, handling requests for various file operations
func (srv *Server) Serve9P(s *styx.Session) {
	service := path.Join(*inpath, s.Access)
	_, err := os.Stat(service)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	// Establish initial buffer
	client := srv.newClient(path.Join(*inpath, s.Access))

	for s.Next() {
		t := s.Request()
		stat, fp, err := walkTo(path.Join(client.buffer, t.Path()))
		if err != nil {
			t.Rerror("%s", err)
			continue
		}
		switch t := t.(type) {
		case styx.Twalk:
			t.Rwalk(stat, nil)
		case styx.Topen:
			switch t.Path() {
			case "/":
				t.Ropen(mkdir(fp), nil)
			case "/event":
				t.Ropen(mkevent(*client))
			case "/ctl":
				t.Ropen(mkctl(getBase(fp)))
			default:
				t.Ropen(os.OpenFile(fp, os.O_RDWR, 0755))
			}
		case styx.Tstat:
			t.Rstat(stat, nil)
		}
	}
}
