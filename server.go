package main

import (
	"log"
	"os"
	"path"

	"aqwari.net/net/styx"
	"github.com/google/uuid"
)

// List of clients' current buffers, useful for filtering events that each client receives
type Server struct {
	client map[uuid.UUID]string
}

func NewServer() *Server {
	client := make(map[uuid.UUID]string)
	return &Server{client: client}
}

// Serve9P is called by styx.ListenAndServe on a client connection, handling requests for various file operations
func (srv *Server) Serve9P(s *styx.Session) {
	// Verify service exists (a named directory in *inpath)
	service := path.Join(*inpath, s.Access)
	_, err := os.Stat(service)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	// Establish initial buffer
	cid := uuid.New()
	srv.client[cid] = DefaultBuffer(service)

	for s.Next() {
		t := s.Request()
		fp := path.Join(srv.client[cid], t.Path())
		stat, err := os.Stat(fp)
		// This can happen if we're trying to stat a ctl file, for example
		if err != nil {
			stat, _ = os.Stat(getBase(fp))
		}
		switch t := t.(type) {
		case styx.Twalk:
			t.Rwalk(stat, nil)
		case styx.Topen:
			switch t.Path() {
			case "/":
				t.Ropen(mkdir(fp), nil)
//			case "event":
//				t.Ropen(mkEvent())
//			case "ctl":
//				t.Ropen(mkCtl(fp), nil)
			default:
				t.Ropen(os.OpenFile(fp, os.O_RDWR, 0755))
			}
		case styx.Tstat:
			t.Rstat(stat, nil)
		}
	}
}
