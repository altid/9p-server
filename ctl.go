package main

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

type ctlFile struct {
	data    []byte
	client  *Client
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
	line := string(p)
	token := strings.Fields(line)
	f.modTime = time.Now().Truncate(time.Hour)

	switch token[0] {
	case "buffer":
		if (len(token) < 2) {
			return 0, errors.New("No buffers specified")
		}
		current := path.Join(f.client.service, token[1])
		if _, err = os.Lstat(current); err != nil {
			return 0, err
		}
		f.client.buffer = current
		return len(p), nil
	case "close":
		if (len(token) < 2) {
			return 0, errors.New("No buffer specified")
		}
		// TODO: Remove from `tabs`
		f.client.buffer = DefaultBuffer(f.client.service)
		return len(p), nil
	case "open", "join":
		if (len(token) < 2) {
			return 0, errors.New("No buffer specified")
		}
		if err != nil {
			return 0, err
		}
		current := path.Join(f.client.service, token[1])
		f.client.buffer = current
	
	}
	name := path.Join(f.client.service, "ctrl")
	fp, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return 0, err
	}
	defer fp.Close()
	return fp.Write(p)
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
func mkctl(ctl, uid string, client *Client) (*ctlFile, error) {
	buff, err := ioutil.ReadFile(ctl)
	if err != nil {
		return nil, err
	}
	buff = append(buff, []byte("buffer\nopen\nclose\n")...)
	return &ctlFile{
		data:    buff,
		size:    int64(len(buff)),
		off:     0,
		modTime: time.Now(),
		uid:     uid,
		client:  client,
	}, nil
}
