package main

import (
	"bytes"
	//"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
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

// Path validation is a kludgy mess currently.
// Rewrite will use new data structure to clean this up greatly
func (f *ctlFile) WriteAt(p []byte, off int64) (n int, err error) {
	// TODO: We need to send event for all new buffer items.
	if ok, _ := regexp.Match("buffer *", p); ok {
		fp := string(bytes.TrimLeft(p, "buffer "))
		fp = strings.Trim(fp, "\n")
		fp = path.Join(f.client.service, fp)
		if _, err = os.Lstat(fp); err != nil {
			return
		}
		f.client.buffer = fp
	} else if ok, _ = regexp.Match("open *", p); ok {
		// Switch buffer to request, and send write to underlying ctl
		fp := string(bytes.TrimLeft(p, "open "))
		fp = strings.Trim(fp, "\n")
		fp = path.Join(f.client.service, fp)
		err = ioutil.WriteFile(path.Join(f.client.service, "ctrl"), p, 0644)
		f.client.buffer = path.Dir(fp)
	} else if ok, _ = regexp.Match("close *", p); ok {
		fp := string(bytes.TrimLeft(p, "close "))
		fp = strings.Trim(fp, "\n")
		fp = path.Join(f.client.service, fp)
		f.client.buffer = DefaultBuffer(f.client.service)
		err = ioutil.WriteFile(path.Join(f.client.service, "ctrl"), p, 0644)
	} else {
		err = ioutil.WriteFile(f.client.service, p, 0644)
	}
	f.modTime = time.Now().Truncate(time.Hour)
	n = len(p)
	return
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
