package main

import (
	"fmt"
	"github.com/mortdeus/go9p/srv"
)

/* Dir
Normalfiles (
	ctl
	status
	input
	feed
	doc
	title
	tabs
	etc
)
	// Readdirnames(n int) (names []string, err error)
	// add list of files in the buffer we're watching
	// add list of Normalfiles we want to check for
	// - if file already exists in our dir object, don't add. (map by string)
*/
type Fakefile struct {
	name string
}

func (f *Fakefile) ConnOpened(conn *srv.Conn) {
	// Handle client setup
}

func (f *Fakefile) ConnClosed(conn *srv.Conn) {
	// Handle client cleanup
}

func (f *Fakefile) FidDestroy(sfid *srv.Fid) {
}

// Attach - Client attaches, set up 
func (f *Fakefile) Attach(req *srv.Req) {
	fmt.Println("Attached")
}

// Read - switch for ctl, input, tabs, etc to fabricate content
func (f *Fakefile) Read(req *srv.Req) {
}

// Write - Make sure we grab our content we want
func (f *Fakefile) Write(req *srv.Req) {
}

// Wstat - flush stat, A no-op for now. 
func (f *Fakefile) Wstat(req *srv.Req) {
}

// Open - Just OpenFile really, it'll match our buffer, fallback, or fail.
func (f *Fakefile) Open(req *srv.Req) {
}

func (f *Fakefile) Stat(req *srv.Req) {
}

// No Create, Remove
