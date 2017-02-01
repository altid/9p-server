package ubqtlib

import (
	"errors"
	"log"
	"os"
	"path"

	"aqwari.net/net/styx"
)

// ClientHandler must be satisfied to use this library
type ClientHandler interface {
	ClientWrite(filename string, client string, data []byte) (int, error)
	ClientRead(filename string, client string) ([]byte, error)
	ClientClose(filename string, client string) error
}

// Srv - Defaults to port :4567
type Srv struct {
	show    map[string]bool
	port    string
	verbose bool
	debug   bool
}

// Event sends back client events (Reads, writes, closes)
type Event struct {
	filename string
	client   string
}

// NewSrv returns a server type
func NewSrv() *Srv {
	show := make(map[string]bool)
	return &Srv{port: ":4567", show: show, debug: false, verbose: false}
}

// SetPort - Accepts a string in the form ":nnnn", representing the port to listen on for the 9p connection
func (u *Srv) SetPort(s string) {
	//TODO: Sanitize s
	u.port = s
}

// Debug - enable debugging output to the log
func (u *Srv) Debug() {
	u.debug = true
}

// Verbose - Enable verbose logging of messages
func (u *Srv) Verbose() {
	u.verbose = true
}

// AddFile - Adds file to the directory structure
func (u *Srv) AddFile(filename string) error {
	_, ok := u.show[filename]
	if !ok {
		u.show[filename] = true
		return nil
	}
	return errors.New("File already exists")
}

// Loop - Starts up ListenAndServe instance of 9p with our settings
func (u *Srv) Loop(client ClientHandler) error {
	fs := styx.HandlerFunc(func(s *styx.Session) {
		for s.Next() {
			t := s.Request()
			name := path.Base(t.Path())
			fi := &stat{name: name, file: &fakefile{name: name, handler: client, client: s.User}}
			switch t := t.(type) {
			case styx.Twalk:
				t.Rwalk(fi, nil)
			case styx.Topen:
				switch name {
				case "/":
					t.Ropen(mkdir(u), nil)
				default:
					t.Ropen(fi.file, nil)
				}
			case styx.Tstat:
				t.Rstat(fi, nil)
			case styx.Tcreate:
				t.Rerror("permission denied")
			case styx.Tremove:
				t.Rerror("permission denied")

			}
		}
	})
	var srv styx.Server
	if u.verbose {
		srv.ErrorLog = log.New(os.Stderr, "", 0)
	}
	if u.debug {
		srv.TraceLog = log.New(os.Stderr, "", 0)
	}
	srv.Addr = u.port
	srv.Handler = fs
	log.Fatal(srv.ListenAndServe())
	return nil
}
