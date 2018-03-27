package main

import (
	"io/ioutil"
	"log"
	"fmt"
	"os"
	"path"
	"sync"
	"strings"

	"aqwari.net/net/styx"
)

type Server struct {
	file map[string]interface{}
	tabs map[string]struct{}
	sync.Mutex
}

// OpenTab - add a buffer to our list of open tabs
func (srv *Server) OpenTab(name string) {
	srv.Lock()
	defer srv.Unlock()
	var s struct{}
	if _, ok := srv.tabs[name]; !ok {
		srv.tabs[name] = s
	}
}

// CloseTab - remove a buffer from our list of open tabs
func (srv *Server) CloseTab(name string) {
	srv.Lock()
	defer srv.Unlock()
	if _, ok := srv.tabs[name]; ok {
		delete(srv.tabs, name)
	}
}

// UpdateTab - Change the title of the current open tab
func (srv *Server) UpdateTab(oldTab, newTab string) {
	srv.CloseTab(oldTab)
	srv.OpenTab(newTab)
}

func (srv *Server) Tabs() (ret string) {
	for k, _ := range srv.tabs {
		ret += k + "\n"
	}
	return
}
// NewServer will return a top-level overview, typically only invoked on server startup listing all current fileservers and a ctl/events structure.
func NewServer() (*Server) {
	file := make(map[string]interface{})
	listing := "Welcome to Ubqt! The following servers are available.\n"
	dirs, err := ioutil.ReadDir(*inpath)
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range dirs {
		if d.IsDir() {
			listing = listing + "	 - " + d.Name() + "\n"
		} else {
			file[d.Name()] = d
		}
	}
	file["doc"] = listing

	// We add ctl and events here, but usually they are from an underlying FS
	synths := map[string]string {"ctl": "ctl", "event": "event", "tabs": "main"}
	for s, d := range synths {
		file[s] = d
	}
	var s struct{}
	tabs := make(map[string]struct{})
	tabs["main"] = s

	return &Server{file: file, tabs: tabs}
}

func newChrootClient(root string, srv *Server) map[string]interface{} {
	file := make(map[string]interface{})
	dir, err := ioutil.ReadDir(path.Join(*inpath, root))
	if err != nil {
		log.Println(err)
		return nil
	}
	for _, d := range dir {
		file[d.Name()] = d
	}
	file["tabs"] = srv.Tabs()
	return file
}

// Serve9P is called by styx.ListenAndServe on a client connection, handling requests for various file operations
func (srv *Server) Serve9P(s *styx.Session) {
	buffer := s.Access
	/* TODO: Switch this shit to use proper os.FileInfo arrays
	var client map[string]interface{}
	if s.Access == "" {
		client = make(map[string]interface{})
		client = srv.file
	} else {
		client = newChrootClient(s.Access, srv) 
	}
	*/
	// As well, we need an array of files backing
	// We don't need an interface here, it should be simply os.FileInfo or arrays of them in the case of dir.
	// We also need our own version of event
	for s.Next() {
		t := s.Request()
		// Main loop to handle requests
		switch t := t.(type) {
		case styx.Twalk:
			t.Rwalk(fi, nil)
		case styx.Topen:
			// TODO: switch to names here for case
			switch v := file.(type) {
			// Real file
			case os.FileInfo:
				fullPath := path.Join(*inpath, buffer, t.Path())
				t.Ropen(os.OpenFile(fullPath, os.O_RDWR, 0755))
			// Dir file
			case map[string]interface{}:
				t.Ropen(mkdir(v), nil)
			// Synthesized file
			default:
				t.Ropen(strings.NewReader(fmt.Sprint(v)), nil)
			}
		case styx.Tstat:
			t.Rstat(fi, nil)
		}
	}
}
