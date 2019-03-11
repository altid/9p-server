package main


// TODO: Clean up flow into simple events loop 

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

type tail struct {
	event chan string
	name  string
}

type tailReader struct {
	io.ReadCloser
}

func newTailReader(t *tail) (*tailReader, error) {
	f, err := os.OpenFile(t.name, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return &tailReader{}, err
	}
	if _, err := f.Seek(0, os.SEEK_END); err != nil {
		return &tailReader{f}, err
	}
	return &tailReader{f}, err
}

func (r *tailReader) Read(p []byte) (n int, err error) {
	for {
		n, err := r.ReadCloser.Read(p)
		if n > 0 {
			return n, nil
		} else if err != io.EOF {
			return n, err
		}
		time.Sleep(300 * time.Millisecond)
	}
}

type watcher struct {}

func dirWatch() (*watcher, chan string) {
	events := make(chan string)
	w := &watcher{}
	return w, events
}

// walk *inpath every 10 seconds to test for new services with events file
func (w *watcher) start(event chan string, ctx context.Context) {
	servlist := make(map[string]*tail)
	for {
		findClosed(event, &servlist)
		listeners := findListeners(event, &servlist)
		for _, listener := range listeners {
			go startListeners(event, listener, ctx)
		}
		select {
		case <- ctx.Done():
			return
		case <- time.After(10 * time.Second):
			continue
		}
	}
}

func findClosed(event chan string, servlist *map[string]*tail) {
	glob := path.Join(*inpath, "*", "event")
	files, _ := filepath.Glob(glob)
LOOP:
	for sname, _ := range *servlist {
		for _, file := range files {
			if file == sname {
				continue LOOP
			}
		}
		delete(*servlist, sname)
		event <- "closed " + sname
	}
}

func findListeners(event chan string, servlist *map[string]*tail) []*tailReader {
	var listeners []*tailReader
	glob := path.Join(*inpath, "*", "event")
	files, _ := filepath.Glob(glob)
	s := *servlist
	for _, file := range files {
		if s[file] != nil {
			continue
		}
		t := &tail{
			event: event,
			name:  file,
		}
		reader, err := newTailReader(t)
		if err != nil {
			continue
		}
		s[file] = t
		listeners = append(listeners, reader)
		event <- "new " + path.Dir(file)
	}
	*servlist = s
	return listeners
}

func startListeners(event chan string, t *tailReader, ctx context.Context) {
	scanner := bufio.NewScanner(t)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return	
		case event <- scanner.Text():
		}
	}
}
