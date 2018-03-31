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
		var stat os.FileInfo
		// Make sure we try to catch most common files
		switch t.Path() {
		// TODO: event has a seperate stat from it's own type (For FIFO)
		case "ctl", "event":
			stat, err = os.Stat(getBase(fp))
			if err != nil { 
				//t.Rerror(e)
				log.Println(err)
			}
		default:
			stat, err = os.Stat(fp)
			// If we have an error here, try to get a good stat. 
			if err != nil {
				stat, err = os.Stat(getBase(fp))
				if err != nil {
					// This is getting rediculous...
					log.Println(err)
				}
			}
		}

		switch t := t.(type) {
		case styx.Twalk:
			t.Rwalk(stat, nil)
		case styx.Topen:
			switch t.Path() {
			case "/":
				t.Ropen(mkdir(fp), nil)
// TODO: Write functions for mkEvent and mkCtl
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
