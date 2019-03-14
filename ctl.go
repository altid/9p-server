package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
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
	if err != nil && err != io.EOF {
		return
	}
	switch command {
	case "buffer":
		// NOTE(halfwit): This abuses semantics of String()
		// String() sets the value of buffer to <nil> should it be empty
		// The Lstat will fail, and the message will be descriptive for all cases.
		current := path.Join(f.cl.service, action)
		if _, err = os.Lstat(current); err != nil {
			return 0, fmt.Errorf("No such buffer: %s\n", action)
		}
		f.cl.buffer = current
		return size, nil
	case "close":
		if f.cl.buffer == action {
			f.cl.buffer = defaultBuffer(f.cl.service)
		}
	case "open":
		if action != "<nil>" {
			f.cl.buffer = path.Join(f.cl.service, action)
		}
	}
	name := path.Join(f.cl.service, "ctrl")
	fp, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return 0, err
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

// This returns a ready rwc for future reads/writes
func mkctl(ctl, uid string, cl *client) (*ctlFile, error) {
	buff, err := ioutil.ReadFile(ctl)
	if err != nil {
		return nil, err
	}
	return &ctlFile{
		data:    buff,
		size:    int64(len(buff)),
		off:     0,
		modTime: time.Now(),
		uid:     uid,
		cl:  cl,
	}, nil
}
