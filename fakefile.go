package main

import (
	"log"
	"github.com/mortdeus/go9p/srv"
)

// TODO: Research ordering of connOpened and attach requests from go9p
/* TODO: When we encounter one of these files, we need to synthesize the response; else we pass the regular requested file along, if it exists 
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
	// TODO: Define our struct for fakefile
	name string
}

func (f *Fakefile) ConnOpened(conn *srv.Conn) {
	if *debug > 0 {
		log.Println("Connected")
	}
	// TODO: Handle client setup
	// Client will be handed the default (generally last used) buffer
	// Requires DefaultFile as well
}

func (f *Fakefile) ConnClosed(conn *srv.Conn) {
	if *debug > 0 {
		log.Println("Disconnected")
	}
	// TODO: Handle client cleanup
}

func (f *Fakefile) FidDestroy(sfid *srv.Fid) {
	// TODO: Clean up tabs, clients
}

// Attach - Client attaches, set up 
func (f *Fakefile) Attach(req *srv.Req) {
	// TODO: If we have a qualified path on the attach request, hand client access only to that requested directory
	if *debug > 0 {
		log.Println("Attached")
	}
}

// Read - switch for ctl, input, tabs, etc to fabricate content
func (f *Fakefile) Read(req *srv.Req) {
	// TODO: Implement main read function, which will synthesize content or pass along fd of existing files
}

// Write - Make sure we grab our content we want
func (f *Fakefile) Write(req *srv.Req) {
	// TODO: Implement main write function, which will handle control and input to our underlying filesystems, or pass along to fd of existing files
}

// Wstat - flush stat, A no-op for now. 
func (f *Fakefile) Wstat(req *srv.Req) {
	// TODO: Implement for all normal files
}

// Open - Just OpenFile really, it'll match our buffer, fallback, or fail.
func (f *Fakefile) Open(req *srv.Req) {
	// TODO: Wrap OpenFile
}

func (f *Fakefile) Stat(req *srv.Req) {
	// TODO: Synthesize stat or os.Stat on a regular fd
}

// TODO: See if there's any case where `create` or `remove` on the client side is beneficial
