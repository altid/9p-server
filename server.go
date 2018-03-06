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

// UpdateDab - Change the title of the current open tab
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

func walkTo(v interface{}, loc string) (interface{}, bool) {
	cwd := v
	parts := strings.FieldsFunc(loc, func(r rune) bool {return r == '/' })
	for _, p := range parts {
		switch v := cwd.(type) {
		// Dir or synthesized file
		case map[string]interface{}:
			if file, ok := v[p]; !ok {
				return nil, false
			} else {
				cwd = file
			}
		default:
			return nil, false
		}
	}
	// File requested is at the end of the tree.
	return cwd, true
}

// Serve9P is called by styx.ListenAndServe on a client connection, handling requests for various file operations
func (srv *Server) Serve9P(s *styx.Session) {
	buffer := s.Access
	var client map[string]interface{}
	if s.Access == "" {
		client = make(map[string]interface{})
		client = srv.file
	} else {
		client = newChrootClient(s.Access, srv) 
	}
	for s.Next() {
		t := s.Request()
		file, ok := walkTo(client, t.Path())
		if !ok {
			t.Rerror("No such file or directory")
			continue
		}		
		var fi os.FileInfo
		fullPath := path.Join(*inpath, buffer, t.Path())
		// Decide whether we need a real file or fake one
		switch t.Path() {
		case "/input", "/title", "/status", "/feed", "/doc", "/stream", "/tabs", "/events":

			fi, _  = os.Stat(fullPath)
		case "/ctl":
			fi = &cstat{name: path.Base(t.Path()), file: &ctl{path: fullPath, v: file}}
		default:
			fi = &stat{name: path.Base(t.Path()), file: &fakefile{v: file}}
		}		

		// Main loop to handle requests
		switch t := t.(type) {
		case styx.Twalk:
			t.Rwalk(fi, nil)

		case styx.Topen:
			switch v := file.(type) {
			// Real file
			case os.FileInfo:
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
