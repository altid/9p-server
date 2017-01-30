package ubqt

import (
	"io"
	"log"
	"os"
	"path"

	"aqwari.net/net/styx"
)

//TODO: Testing. connect and start up, and verify that any numerical port works, verify that our path is a good variable

// Default files that clients will use to draw UI
const (
	INPUT   = "input"
	CTL     = "ctl"
	STATUS  = "status"
	TITLE   = "title"
	TABS    = "tabs"
	SIDEBAR = "sidebar"
	MAIN    = "main"
)

// Ubqt - Defaults to port :4567, ~/ubqt
type Ubqt struct {
	show    map[string]bool
	port    string
	path    string
	debug   bool
	verbose bool
}

func newUbqt() *Ubqt {
	return &Ubqt{port: ":4567", path: "~/ubqt"}
}

// SetPort - Accepts a string in the form ":nnnn", representing the port to listen on for the 9p connection
func (u *Ubqt) SetPort(s string) {
	u.port = s
}

// Debug - Enable debugging output
func (u *Ubqt) Debug() {
	u.debug = true
}

// Verbose - Enable verbose logging
func (u *Ubqt) Verbose() {
	u.verbose = true
}

// Start - Starts up ListenAndServe instance of 9p with our settings
func (u *Ubqt) Start() error {
	var srv styx.Server
	if u.verbose {
		srv.ErrorLog = log.New(os.Stderr, "", 0)
	}
	if u.debug {
		srv.TraceLog = log.New(os.Stderr, "", 0)
	}
	srv.Addr = u.path
	srv.Handler = u
	err := srv.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

// Serve9P - Called on client connection (internal)
func (u *Ubqt) Serve9P(s *styx.Session) {
	for s.Next() {
		t := s.Request()
		name := path.Base(t.Path())
		fi := &stat{name: name, file: &fakefile{name: name}}
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
}
