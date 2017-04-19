package ubqtlib

import (
	"errors"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"aqwari.net/net/styx"
)

// ClientHandler must be satisfied to use this library
type ClientHandler interface {
	ClientWrite(filename string, client string, data []byte) (int, error)
	ClientRead(filename string, client string) ([]byte, error)
	ClientOther(filename string, client string) (*os.File, error)
	ClientConnect(client string)
	ClientDisconnect(client string)
}

// Client is a map of our files
type Client map[string]*fakefile

// Srv - Defaults to port :4567
type Srv struct {
	show    map[string]bool
	event   map[string]chan []byte
	port    string
	verbose bool
	debug   bool
	input   []byte
	sync.Mutex
}

// NewSrv returns a server type
func NewSrv() *Srv {
	show := make(map[string]bool)
	event := make(map[string]chan []byte)
	return &Srv{port: ":4567", show: show, debug: false, verbose: false, event: event}
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

// SendEvent - Send an event to any clients that are currently blocking for data
func (u *Srv) SendEvent(file []byte) {
	for _, name := range u.event {
		go func(c chan []byte) {
			c <- file
		}(name)
	}
}

// Loop - Starts up ListenAndServe instance of 9p with our settings
func (u *Srv) Loop(client ClientHandler) error {
	fs := styx.HandlerFunc(func(s *styx.Session) {
		client.ClientConnect(s.User)
		files := u.newclient(client, s.User)
		u.AddFile("event")
		for s.Next() {
			t := s.Request()
			name := path.Base(t.Path())
			fi, ok := files[name]
			if !ok {
				fi = &fakefile{name: name, mtime: time.Now(), handler: client, client: s.User}
			}
			switch t := t.(type) {
			case styx.Twalk:
				t.Rwalk(fi, nil)
			case styx.Topen:
				switch name {
				case "/":
					t.Ropen(mkdir(files), nil)
				case "event":
					t.Ropen(newEvent(u, s.User), nil)
				case "input", "ctl", "status", "tabs", "title", "sidebar":
					t.Ropen(fi, nil)
				default:
					fi, _ := client.ClientOther(name, s.User)
					t.Ropen(fi, nil)
				}
			case styx.Tstat:
				switch name {
				case "event":
					t.Rstat(newEvent(u, s.User), nil)
				default:
					t.Rstat(fi, nil)
				}
			case styx.Ttruncate:
				t.Rtruncate(nil)
			case styx.Tutimes:
				t.Rutimes(nil)
			case styx.Tsync:
				t.Rsync(nil)
			case styx.Trename:
				t.Rrename(nil)
			//TODO: Allow the creation of files to be managed by the implementation
			case styx.Tcreate:
				t.Rerror("cannot create files")
			case styx.Tchmod:
				t.Rchmod(nil)
			}
		}
		client.ClientDisconnect(s.User)
	})
	var srv styx.Server
	//BUG:(halfwit) On verbose, writes will fail
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
