package main

import (
	"bufio"
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

// TODO: context.Context for cancellation of these tailers
// Read for file changes every 300ms
func (t *tail) start() {
	reader, err := newTailReader(t.name)
	if err != nil {
		log.Print(err)
		return
	}
	go func(rd *tailReader, tl *tail) {
		scanner := bufio.NewScanner(rd)
		for scanner.Scan() {
			tl.event <- scanner.Text()
		}
	}(reader, t)
}

func newTailReader(name string) (*tailReader, error) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDONLY, 0644)
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
func Watch() chan string {
	event := make(chan string)
	tails := make(map[string]*tail)
	go func(chan string) {
		glob := path.Join(*inpath, "*", "event")
		files, err := filepath.Glob(glob)
		if err != nil {
			log.Print(err)
		}
		for _, file := range files {
			log.Print(file)
			if tails[file] == nil {
				tail := &tail{
					event: event,
					name:  file,
				}
				tail.start()
				tails[file] = tail
			}
		}
		time.Sleep(10 * time.Second)
	}(event)
	return event
}
