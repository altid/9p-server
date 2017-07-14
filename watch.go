package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
	"github.com/hpcloud/tail"
)

/*
Registering a client in a multiple-server paradigm:
SIGUSR won't work with multiple servers, especially if arbitrarily named
FIFO won't work, if we have multiple servers digesting them
Inotify, recursive would be fine likely, but webfs and such will grow well beyond the watch limits

Inotify on inpath, add watch to folder until we see `event`, then tail event
fs's will append to events - `printf '%s\n' "title" >> event
If event is deleted, add back to watch
We end up with the following structure:

inpath/
    ircfs/
        event
        ctl
        irc.freenode.net/
        ...
    webfs/
        event
        ctl
        https/
        ...
    ...

We let the os handle write contentions on our behalf, and multiple servers can register to listen to these directories (9p, http, circle (from tickit)?)
File servers should periodically flush their event file as well, to keep the size minimal
*/

func testEvent(name string) bool {
	_, err := os.Stat(path.Join(name, "event"))
	if err != nil {
		return false
	}
	return true
}

func addTail(filename string, event chan string, config tail.Config) {
	t, err := tail.TailFile(filename, config)
	defer t.Cleanup()
	if err != nil {
		//TODO: Log error
		fmt.Printf("Error in addTail: %s\n", err)
	}
	for {
		select {
		case line := <-t.Lines:
			event <- line.Text
		}
	}
}

// Watch will observe our directory, tailing any events file that exists within a second-level directory
func Watch() chan string {

	config := &tail.Config{Follow: true, Location: &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}, Poll: false, ReOpen: true, Logger: tail.DiscardingLogger}
	event := make(chan string)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		// TODO: Log fatal error
		fmt.Println("Error occured %s", err)
	}
	go func() {
		for {
			select {
			case e := <-watcher.Events:
				switch e.Op {
				// CREATE
				case 1:
					if path.Dir(e.Name) != *inpath {
						_, name := path.Split(e.Name)
						if name != "event" {
							continue
						}
						err := watcher.Remove(path.Dir(e.Name))
						if err != nil {
							// Log error
							fmt.Printf("Error removing watch from %s\n", e.Name)
						}
						go addTail(e.Name, event, *config)
						continue
					}
					if testEvent(e.Name) {
						err := watcher.Remove(path.Dir(e.Name))
						if err != nil {
							// Log error
							fmt.Printf("Error removing %s from watch\n", e.Name)
							break
						}
						go addTail(e.Name, event, *config)
						continue
					}
					// If we're here, we can safely add
					//TODO: More testing!
					watcher.Add(e.Name)
				}
			case err := <-watcher.Errors:
				fmt.Printf("error logged %s\n", err)
				// TODO: Log error
			}
		}
	}()

	err = watcher.Add(*inpath)
	if err != nil {
		fmt.Printf("error in adding %s\n", *inpath)
		// TODO: Log fatal error
	}
	// For each directory contained in *inpath, add watch if directory/events is absent
	files, err := ioutil.ReadDir(*inpath)
	if err != nil {
		fmt.Printf("error listing files in %s\n", *inpath)
	}
	for _, file := range files {
		myfile := path.Join(*inpath, file.Name())
		switch testEvent(myfile) {
		case true:
			// We have a directory with events file already
			if file.IsDir() {
				go addTail(path.Join(myfile, "event"), event, *config)
			}
		case false:
			if file.IsDir() {
				watcher.Add(myfile)
			}
			// We happily ignore any normal file in our base dir
		}
	}
	return event
}
