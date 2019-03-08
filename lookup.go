package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
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

// TODO: Modify this to only search in trees related to the server
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
		case "feed", "document", "stream", "form":
			result = path.Dir(fullpath)
		}

		return nil
	})
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return result
}
