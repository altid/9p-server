package main

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
			type tailReader struct {
				io.ReadCloser
			}
		}
		time.Sleep(300 * time.Millisecond)
	}
}

// walk *inpath every 10 seconds to test for new services with events file
func watchServiceDir(ctx context.Context) chan string {
	event := make(chan string)
	go func(chan string) {
		services := make(map[string]*tail)
		for {
			glob := path.Join(*inpath, "*", "event")
			files, _ := filepath.Glob(glob)
			listeners := findListeners(event, files, services)
			for _, listener := range listeners {
				go startListeners(event, listener, ctx)
			}
			// Do a timeout here
			select {
			case <- ctx.Done():
				return
			case <- time.After(10 * time.Second):
				continue
			}
		}
	}(event)
	return event
}

// findListeners will check all files for new listeners
// It will return an array of any that are found
func findListeners(event chan string, files []string, services map[string]*tail) []*tailReader {
	var listeners []*tailReader
	for _, file := range files {
		if services[file] != nil {
			continue
		}
		log.Print(file)
		t := &tail{
			event: event,
			name:  file,
		}
		reader, err := newTailReader(t)
		if err != nil {
			log.Print(err)
			continue
		}
		services[file] = t
		listeners = append(listeners, reader)
	}
	return listeners
}

func startListeners(event chan string, t *tailReader, ctx context.Context) {
	scanner := bufio.NewScanner(t)
	for scanner.Scan() {
		select {
		// This will eventually get garbage collected, but close anyways on shutdown
		case <-ctx.Done():
			return	
		case event <- scanner.Text():
		}
	}
}
