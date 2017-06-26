package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

// getBase will see if file exists at base of current file server
// for example, given a path /home/user/ubqt, running with the irc fileserver at /home/user/ubqt/irc; it will search for /home/user/ubqt/irc/<myfile>.
func getBase(p string) string {
	if !strings.Contains(p, *inpath) {
		return p
	}
	dir, target := path.Split(p)
	for {
		current := path.Dir(dir)
		if current == *inpath {
			break
		}
		// Our path cannot be any shorter, and still be nested. Exit
		if len(current) <= 1 {
			break
		}
		dir = current
	}
	return path.Join(dir, target)
}

// OpenFile - Attempts to open either named file, or that same file at the base of server
func OpenFile(name string) (*os.File, error) {
	file, err := os.Open(name)
	if err != nil {
		file, err = os.Open(getBase(name))
	}
	return file, err
}

// Grab the first directory from base path, use it as buffer (will have ctl file to write to, useful for :open commands only)
func DefaultFile() string {
	files, err := ioutil.ReadDir(*inpath)
	if err != nil {
		log.Fatal(err)
	}
	return files[0].Name()
}
