// +build !plan9

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/hpcloud/tail"
)

/*
Registering a client in a multiple-server paradigm:
SIGUSR won't work with multiple servers, especially if arbitrarily named
FIFO won't work, if we have multiple servers digesting them
Inotify, recursive would be fine likely, but webfs and such will grow well beyond the watch limits
fs's will append to events - `printf '%s\n' "title" >> event
If event is deleted, add back to watch
We end up with the following structure:

inpath/
    irc.freenode.net
        event
        ctrl
        ...
    docs
        event
        ctrl
        ...
    ... somefile/

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

// Using github.com/hpcloud/tail watch file for appended text
func addTail(filename string, event chan string, config tail.Config, watcher *fsnotify.Watcher) {
	t, err := tail.TailFile(filename, config)
	defer t.Cleanup()
	if err != nil {
		log.Printf("Error in addTail: %s\n", err)
		return
	}
DONE:
	for {
		select {
		case <-t.Dead():
			// TODO: Read on t.Dying until the tail is properly closed
			// A timeout is required here until then
			time.Sleep(100 * time.Millisecond)
			break DONE // Break from labelled block
		case line := <-t.Lines:
			event <- line.Text
		}
	}
	watcher.Add(path.Dir(filename))
}

// Watch will observe our directory, tailing any events file that exists within a second-level directory
// This should only send messages when it finds events files back to main loop
// When we get an event, we will add a tailer there, and set up the Serve().
func watchServiceDir() chan string {

	config := &tail.Config{
		Follow: true, 
		Location: &tail.SeekInfo{
			Offset: 0, 
			Whence: os.SEEK_END,
		}, 
		Poll: false, 
		MustExist: true, 
		ReOpen: true, 
		Logger: tail.DiscardingLogger,
	}
	event := make(chan string)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error occured %s", err)
	}
	go func() {
		for {
			select {
			case e := <-watcher.Events:
				switch e.Op {
				// CREATE
				case 1:
					if path.Dir(e.Name) != *inpath {
						if path.Base(e.Name) != "event" {
							continue
						}
						err := watcher.Remove(path.Dir(e.Name))
						if err != nil {
							log.Printf("Error removing watch from %s\n", e.Name)
						}
						// TODO: skip addtail, move to main loop
						go addTail(e.Name, event, *config, watcher)
						event <- e.Name
						continue
					}
					if testEvent(e.Name) {
						err := watcher.Remove(path.Dir(e.Name))
						if err != nil {
							log.Printf("Error removing %s from watch\n", e.Name)
							break
						}
						// TODO: skip addtail, move to main loop
						go addTail(e.Name, event, *config, watcher)
						event <- e.Name
						continue
					}
					// If we're here, we can safely add
					watcher.Add(e.Name)
				case 3:
					// TODO: REMOVE event - test if directory is still present
				case 4:
					// TODO: RENAME event - test if directory is still present
				}
			case err := <-watcher.Errors:
				log.Printf("error logged %s\n", err)
			}
		}
	}()

	err = watcher.Add(*inpath)
	if err != nil {
		log.Fatalf("error in adding %s\n", *inpath)
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
				// addtail to main event loop
				go addTail(path.Join(myfile, "event"), event, *config, watcher)
				event <- myfile
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
