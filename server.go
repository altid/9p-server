package main
// TODO: There's currently a lot of variable shuffling in order to track our data types. Research using channels to listen on for read/write to and from `files`
import (
	"os"
	"path"

	"aqwari.net/net/styx"
	"github.com/google/uuid"
)

type Client struct {
	buffer string
	service string
	event chan string
}

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
	srv.client[cid] = &Client{ buffer: buffer, service: service, event: ch }
	return srv.client[cid]
}

// Get a useful stat for the requested path
func walkTo(c *Client, filep string, uid string) (os.FileInfo, string, error) {
	// We prematurely create a stat type for each file here
	fp := path.Join(c.buffer, filep)
	switch fp {
	// TODO: Implement `stat` type for event
	case "/ctl":
		base := getBase(fp)
		ctlfile, err := mkctl(base, uid, c)
		if err != nil {
			return nil, fp, err
		}
		return &ctlStat{name: "ctl", file: ctlfile }, base, nil
	default:
		stat, err := os.Stat(fp)
		// If we have an error here, try to get a base-level stat. 
		if err != nil {
			stat, err = os.Stat(getBase(fp))
			if err != nil {
				return nil, fp, err
			}
			return stat, getBase(fp), nil
		}
		return stat, fp, nil
	}
}

// TODO: Research if it's worthwhile to switch to a stack. One for ctl, one for event, one for dir, and one for the rest.
// May be difficult to do bookkeeping between all the different handlers, but it'd keep it much cleaner.
// Serve9P is called by styx.ListenAndServe on a client connection, handling requests for various file operations
func (srv *Server) Serve9P(s *styx.Session) {
	client := srv.newClient(path.Join(*inpath, s.Access))

	for s.Next() {
		t := s.Request()
		stat, fp, err := walkTo(client, t.Path(), s.User)
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
				t.Ropen(mkdir(fp, s.User, client), nil)
			case "/event":
				t.Ropen(mkevent(*client))
			case "/ctl":
				t.Ropen(mkctl(fp, s.User, client))
			default:
				t.Ropen(os.OpenFile(fp, os.O_RDWR|os.O_APPEND, 0666))
			}
		case styx.Tstat:
			t.Rstat(stat, nil)
		// These are handled by the underlying OS calls
		case styx.Tutimes:
			switch t.Path() {
			case "/", "/event", "/ctl":
				t.Rutimes(nil)
			default:
				t.Rutimes(os.Chtimes(fp, t.Atime, t.Mtime))
			}
		case styx.Ttruncate:
			switch t.Path() {
			case "/", "/event", "/ctl":
				t.Rtruncate(nil)
			default:
				t.Rtruncate(os.Truncate(fp, t.Size))
			}
		}
	}
}
