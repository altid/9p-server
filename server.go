package ubqtlib

import (
	"errors"
	"log"
	"os"
	"path"
	"time"

	"aqwari.net/net/styx"
)

// ClientHandler must be satisfied to use this library
type ClientHandler interface {
	ClientWrite(filename string, client string, data []byte) (int, error)
	ClientRead(filename string, client string) ([]byte, error)
}

// Client is a map of our files
type Client map[string]*fakefile

// Srv - Defaults to port :4567
type Srv struct {
	show    map[string]bool
	port    string
	verbose bool
	debug   bool
	input   []byte
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

func (u *Srv) newclient(h ClientHandler, c string) Client {
	files := make(map[string]*fakefile)
	for n, show := range u.show {
		if show {
			files[n] = &fakefile{name: n, handler: h, client: c, mtime: time.Now()}
		}
	}
	return files
}

// Loop - Starts up ListenAndServe instance of 9p with our settings
func (u *Srv) Loop(client ClientHandler) error {
	fs := styx.HandlerFunc(func(s *styx.Session) {
		files := u.newclient(client, s.User)
		for s.Next() {
			t := s.Request()
			name := path.Base(t.Path())
			fi, ok := files[name]
			if !ok {
				// We're at either /, or an arbitrary file
				fi = &fakefile{name: name, mtime: time.Now(), handler: client, client: s.User}
			}
			switch t := t.(type) {
			case styx.Twalk:
				t.Rwalk(fi, nil)
			case styx.Topen:
				switch fi.name {
				case "/":
					t.Ropen(mkdir(files), nil)
				default:
					t.Ropen(fi, nil)
				}
			case styx.Tstat:
				t.Rstat(fi, nil)
			case styx.Ttruncate:
				t.Rtruncate(nil)
			case styx.Tutimes:
				t.Rutimes(nil)
			case styx.Tsync:
				t.Rsync(nil)
			case styx.Tcreate:
				if fi.IsDir() {
					t.Rerror("Cannot create directories")
				} else {
					t.Rcreate(fi, nil)
				}
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
