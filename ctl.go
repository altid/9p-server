package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

type ctlFile struct {
	data    []byte
	cl      *client
	modTime time.Time
	size    int64
	off     int64
	uid     string
}

func (f *ctlFile) ReadAt(b []byte, off int64) (n int, err error) {
	n = copy(b, f.data[off:])
	if int64(n)+off > f.size {
		return n, io.EOF
	}
	return
}

func (f *ctlFile) WriteAt(p []byte, off int64) (n int, err error) {
	size := len(p)
	f.modTime = time.Now().Truncate(time.Hour)
	f.off += off + int64(size)

	buff := bytes.NewBuffer(p)
	command, err := buff.ReadString(' ')
	action := buff.String()
	action = strings.TrimSpace(action)
	if err != nil && err != io.EOF {
		return
	}
	// NOTE(halfwit): This abuses semantics of String()
	// String() sets the value of buffer to <nil> should it be empty
	// The Lstat will fail, and the message will be descriptive for all cases
	switch strings.TrimSpace(command) {
	case "buffer":
		f.cl.tabs[f.cl.buffer] = "grey"
		current := path.Join(f.cl.service, action)
		if _, err = os.Lstat(current); err != nil {
			return 0, fmt.Errorf("Error swapping buffers to %s: %s\n", action, err)
		}
		f.cl.buffer = current
		f.cl.tabs[current] = "purple"
		return size, nil
	case "close":
		if _, ok := f.cl.tabs[action]; ok {
			delete(f.cl.tabs, action)
		}
		if f.cl.buffer == action {
			buffer := defaultBuffer(f.cl.service)
			if buffer != "" {	
				f.cl.buffer = buffer
				f.cl.tabs[buffer] = "purple"
			}
		}
	// NOTE(halfwit): Same as above, nil means the buffer was empty
	// when we tried to read on it
	case "open":
		if action == "<nil>" {
			return 0, errors.New("No resource specified to open")
		}
		f.cl.tabs[f.cl.buffer] = "grey"
		if _, ok := f.cl.tabs[action]; ! ok {
			f.cl.tabs[action] = "purple"
		}
		f.cl.buffer = path.Join(f.cl.service, action)
	}
	name := path.Join(f.cl.service, "ctrl")
	fp, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	return fp.WriteString(command + " " + action)
}

func (f *ctlFile) Close() error { return nil }
func (f *ctlFile) Uid() string  { return f.uid }
func (f *ctlFile) Gid() string  { return f.uid }

type ctlStat struct {
	name string
	file *ctlFile
}

func (s *ctlStat) Name() string       { return s.name }
func (s *ctlStat) Sys() interface{}   { return s.file }
func (s *ctlStat) ModTime() time.Time { return s.file.modTime }
func (s *ctlStat) IsDir() bool        { return false }
func (s *ctlStat) Mode() os.FileMode  { return 0644 }
func (s *ctlStat) Size() int64        { return s.file.size }

func mkctl(ctl, uid string, cl *client) (*ctlFile, error) {
	buff, err := ioutil.ReadFile(ctl)
	if err != nil {
		return nil, err
	}
	c := &ctlFile{
		data:    buff,
		size:    int64(len(buff)),
		off:     0,
		modTime: time.Now(),
		uid:     uid,
		cl:  cl,
	}
	return c, nil
}
