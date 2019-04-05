package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mischief/ndb"
	"github.com/altid/fslib"
)

// NOTE(halfwit): getBase will see if file exists at base of current file server
// example: running with the irc fileserver at /home/user/altid/irc
// it will search for /home/user/altid/irc/somebuffer/somefile
// and walk back in the dir structure until it finds the file
// or the path gets trimmed down to *inpath or "/"
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

func defaultBuffer(root string) string {
	// Recursively walk the tree down until we find a useful file
	var result string
	err := filepath.Walk(root, func(fullpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		switch info.Name() {
		case "feed", "document", "stream":
			result = path.Dir(fullpath)
		}

		return nil
	})
	if err != nil {
		return ""
	}
	return result
}

// NOTE(halfwit): If the listen address isn't found here, it returns
// just ":564", the default 9p listen port. The IP will default
// to the first IP on the system network stack.
func findListenAddress(service string) string {
	listen_address := ":564"
	confdir, err := fslib.UserConfDir()
	if err != nil {
		return listen_address
	}
	conf, err := ndb.Open(path.Join(confdir, "altid", "config"))
	if err != nil {
		return listen_address
	}
	service = path.Base(service)
	listen_address = conf.Search("service", service).Search("listen_address")
	if listen_address == "" {
		return ":564"
	}
	return listen_address
}
