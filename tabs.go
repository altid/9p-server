package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

// TODO: Rename to dir.go, just have it gen for dir. Create ctl, events, and tabs.go to include types and methods on those types for use here
// Interface matching normal files
type fakefile struct {
	v      interface{}
	offset int64
	set    func(s string)
}

func (f *fakefile) ReadAt(p []byte, off int64) (int, error) {
	var s string
	if v, ok := f.v.(fmt.Stringer); ok {
		s = v.String()
	} else {
		s = fmt.Sprint(f.v)
	}
	if off > int64(len(s)) {
		return 0, io.EOF
	}
	n := copy(p, s)
	return n, nil
}

func (f *fakefile) WriteAt(p []byte, off int64) (int, error) {
	// TODO: We need to wrap ctl writes and intercept any server-side commands
	buf, ok := f.v.(*bytes.Buffer)
	if !ok {
		return 0, errors.New("not supported")
	}
	if off != f.offset {
		return 0, errors.New("no seeking")
	}
	n, err := buf.Write(p)
	f.offset += int64(n)
	return n, err
}

func (f *fakefile) Close() error {
	if f.set != nil {
		f.set(fmt.Sprint(f.v))
	}
	return nil
}

func (f *fakefile) size() int64 {
	switch f.v.(type) {
	case map[string]interface{}, []interface{}:
		return 0
	}
	return int64(len(fmt.Sprint(f.v)))
}

type stat struct {
	name string
	file *fakefile
}

func (s *stat) Name() string     { return s.name }
func (s *stat) Sys() interface{} { return s.file }

func (s *stat) ModTime() time.Time {
	return time.Now().Truncate(time.Hour)
}

func (s *stat) IsDir() bool {
	return s.Mode().IsDir()
}

func (s *stat) Mode() os.FileMode {
	switch s.file.v.(type) {
	case map[string]interface{}:
		return os.ModeDir | 0755
	case []interface{}:
		return os.ModeDir | 0755
	}
	return 0644
}

func (s *stat) Size() int64 {
	return s.file.size()
}

// This all stays as is. Everything above is obliterated.
type dir struct {
	c    chan stat
	done chan struct{}
}

func mkdir(path string) *dir {
	c := make(chan stat, 10)
	done := make(chan struct{})
	list := ioutil.ReadDir(path)
	// TODO: Investigate how to generate stats for each type
	ctl := &ctlFile{ ... }
	event := &eventFile{ ... }
	append(list, ctl.Stat())
	append(list, event.Stat())
	go func() {
		for f := range list {
			select {
			case c <- f
			case <- done:
				break
			}
		}
		close(c)
	}()
	return &dir{ c: c, done: done, }
}

// Listen for os.FileInfo members to come in from mkdir
func (d *dir) Readdir(n int) ([]os.FileInfo, error) {
	var err error
	fi := make([]os.FileInfo, 0, 10)
	for i := 0; i < n; i++ {
		s, ok := <-d.c
		if !ok {
			err = io.EOF
			break
		}
		fi = append(fi, &s)
	}
	return fi, err
}

func (d *dir) Close() error {
	close(d.done)
	return nil
}