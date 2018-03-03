package main

import (
	"io/ioutil"
	"log"
	"fmt"
	"os"
	"path"
	"sync"
	"strconv"
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
	// This will lay out our starting page
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

	// We add ctl and events here, but usually they are from a fs. This is only for a first-run
	synths := map[string]string {"ctl": "ctl", "event": "event", "tabs": "main"}
	for s, d := range synths {
		file[s] = d
	}
	var s struct{}
	tabs := make(map[string]struct{})
	tabs["main"] = s
	
	return &Server{file: file, tabs: tabs}
}

func newClient(root string, srv *Server) map[string]interface{} {
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

// Walk through directory, check if in path is valid
// return requested file or dir
func walkTo(v interface{}, loc string) (interface{}, bool) {
	cwd := v
	parts := strings.FieldsFunc(loc, func(r rune) bool {return r == '/' })

	for _, p := range parts {
		switch v := cwd.(type) {
		case map[string]interface{}:
			if file, ok := v[p]; !ok {
				return nil, false
			} else {
				cwd = file
			}
		case []interface{}:
			i, err := strconv.Atoi(p)
			if err != nil {
				return nil, false
			}
			if len(v) <= i {
				return nil, false
			}
			cwd = v[i]
		default:
			return nil, false
		}
	}
	return cwd, true
}

func (srv *Server) Serve9P(s *styx.Session) {
	buffer := s.Access
	var client map[string]interface{}
	if s.Access == "" {
		client = make(map[string]interface{})
		client = srv.file
	} else {
		client = newClient(s.Access, srv) 
	}
	for s.Next() {
		t := s.Request()
		file, ok := walkTo(client, t.Path())
		if !ok {
			t.Rerror("No such file or directory")
			continue
		}		
		fi := &stat{name: path.Base(t.Path()), file: &fakefile{v: file}}
		fp := path.Join(*inpath, buffer, t.Path())
		switch t := t.(type) {
		case styx.Twalk:
			switch v := file.(type) {
			case os.FileInfo:
				t.Rwalk(v, nil)
			default:
				t.Rwalk(fi, nil)
			}
		case styx.Topen:
			switch v := file.(type) {
			case os.FileInfo:
				// TODO: Implement a ctl type to handle our reads/writes
				t.Ropen(os.OpenFile(fp, os.O_RDWR, 0644 ))
			case map[string]interface{}, []interface{}:
				t.Ropen(mkdir(v), nil)
			default:
				t.Ropen(strings.NewReader(fmt.Sprint(v)), nil)
			}
		case styx.Tstat:
			switch v := file.(type) {
			case os.FileInfo:
				t.Rstat(v, nil)
			default:
				t.Rstat(fi, nil)
			}
		}
	}
}
