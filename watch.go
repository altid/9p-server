package main

import (
	"bufio"
	"context"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"
)

var Servlist = make(map[string]*tail)

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

// walk *inpath every 10 seconds to test for new services with events file
func startWatcher(ctx context.Context, event chan string) {
	for {
		findClosed(event)
		listeners := findListeners(event)
		for _, listener := range listeners {
			go startListeners(ctx, event, listener)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
			continue
		}
	}
}

func findClosed(event chan string) {
	glob := path.Join(*inpath, "*", "event")
	files, _ := filepath.Glob(glob)
LOOP:
	for sname, _ := range Servlist {
		for _, file := range files {
			if file == sname {
				continue LOOP
			}
		}
		delete(Servlist, sname)
		event <- "closed " + sname
	}
}

// TODO(halfwit): For something with possibly nested buffers, we need walk to recurse through all buffers in our *mtpt
// In an example, we could have httpfs with clients connected at `github.com`, `github.com/altid`, and `github.com/altid/9p-server`; we want to be able to list all of them as buffers 
// (httpfs will be designed so that only pages that clients are currently visiting are available under $mtpt/http/, to keep buffer managament sane)
// https://github.com/altid/9p-server/issues/13
func findListeners(event chan string) []*tailReader {
	var listeners []*tailReader
	glob := path.Join(*inpath, "*", "event")
	files, _ := filepath.Glob(glob)
	for _, file := range files {
		if Servlist[file] != nil {
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
		Servlist[file] = t
		listeners = append(listeners, reader)
		event <- "new " + path.Dir(file)
	}
	return listeners
}

func startListeners(ctx context.Context, event chan string, t *tailReader) {
	scanner := bufio.NewScanner(t)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			t.Close()
			return
		case event <- scanner.Text():
		}
	}
}
