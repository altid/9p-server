package ubqt

import (
	"io"
	"log"
	"os"
	"path"

	"aqwari.net/net/styx"
)

// FileHandler must be implemented by any program wishing to use this library
type FileHandler interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	CloseFile(filename string) error
}

// Srv - Defaults to port :4567
type Srv struct {
	show    map[string]bool
	port    string
	debug   bool
	verbose bool
}

// Event sends back client events (Reads, writes, closes)
type Event struct {
	filename string
	client   string
}

func newSrv() *Srv {
	return &Srv{port: ":4567"}
}

// SetPort - Accepts a string in the form ":nnnn", representing the port to listen on for the 9p connection
func (u *Srv) SetPort(s string) {
	//TODO: Sanitize s
	u.port = s
}

// Debug - Enable debugging output
func (u *Srv) Debug() {
	u.debug = true
}

// Verbose - Enable verbose logging
func (u *Srv) Verbose() {
	u.verbose = true
}

// Loop - Starts up ListenAndServe instance of 9p with our settings
func (u *Srv) Loop(f *FileHandler) error {
	log := styx.HandlerFunc(func(s *styx.Session) {
		for s.Next() {
			if u.verbose {
				log.Printf("%s %q %s", s.User, s.Access, s.Reuest())
			}
			log.Printf("session %s %q ended", s.User, s.Access)
		}
	})
	//TODO: Modify files.go to utilize our FileHandler
	fs := styx.HandlerFunc(func(s *styx.Session) {
		for s.Next() {
			t := s.Request()
			name := path.Base(t.Path())
			fi := &stat{name: name, file: &fakefile{v: f, name: name}}
			//TODO: e := &Event{Filename: name, client: s.User}
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
	styx.ListenAndServe(u.path, styx.Stack(log, fs))
}
