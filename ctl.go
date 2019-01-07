package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type ctlFile struct {
	data chan []byte
	done chan struct{}
	err chan error
	client *Client
	modTime time.Time
	size int64
	off  int64
	uid string
}

func (f *ctlFile) Read(p []byte) (n int, err error) {
	s, ok := <-f.data
	if ! ok {
		return 0, io.EOF
	}
	n = copy(p, s)
	f.off = int64(n)
	return n, err
}

// Path validation is a kludgy mess currently.
// Rewrite will use new data structure to clean this up greatly
func (f *ctlFile) Write(p []byte) (n int, err error) {
	// Make sure we empty our chan
	<-f.data
	// Validate we're given a byte array that matches our schema
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
		err = ioutil.WriteFile(path.Join(f.client.service, "ctl"), p, 0777)
		f.client.buffer = path.Dir(fp)
	} else if ok, _ = regexp.Match("close *", p); ok {
		fp := string(bytes.TrimLeft(p, "close "))
		fp = strings.Trim(fp, "\n")
		fp = path.Join(f.client.service, fp)
		f.client.buffer = DefaultBuffer(f.client.service)
		err = ioutil.WriteFile(path.Join(f.client.service, "ctl"), p, 0777)
	} else {
		err = ioutil.WriteFile(f.client.service, p, 0777)
	}
	f.off = int64(len(p))
	f.modTime = time.Now().Truncate(time.Hour)
	n = len(p)
	return
}

func (f *ctlFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.off = offset
	case io.SeekCurrent:
		f.off += offset
	case io.SeekEnd:
		f.off = f.size + offset
	}
	// No seeking past EOF
	if f.off >= f.size {
	f.off = f.size
		return 0, io.EOF
	}
	return f.off, nil
}

func (f *ctlFile) Close() error {
	close(f.done)
	return nil
}

func (f *ctlFile) Uid() string { return f.uid }
func (f *ctlFile) Gid() string { return f.uid }

type ctlStat struct {
	name string
	file *ctlFile
}

func (s *ctlStat) Name() string     { return s.name }
func (s *ctlStat) Sys() interface{} { return s.file }

func (s *ctlStat) ModTime() time.Time {
	return s.file.modTime
}

func (s *ctlStat) IsDir() bool {
	return false
}

func (s *ctlStat) Mode() os.FileMode {
	return os.ModeAppend | 0666
}

func (s *ctlStat) Size() int64 {
	return s.file.size
}

// This returns a ready rwc for future reads/writes
func mkctl(ctl, uid string, client *Client) (*ctlFile, error) {
	data := make(chan []byte)
	done := make(chan struct{})
	e := make(chan error)
	final, err := ioutil.ReadFile(ctl)
	if err != nil {
		return nil, err
	}
	// TODO: Revisit what we want the ctlfile to contain
	buff := []byte("quit\nrestart\n")
	final = append(buff, final...)
	go func(b []byte) {
		LOOP:
			for {
				select {
				case <-done:
					break LOOP
				case data <-b:
				}
			}
		close(data)
	}(final)
	return &ctlFile{data: data, done: done, err: e, size: int64(len(final)), off: 0, modTime: time.Now().Truncate(time.Hour), uid: uid, client: client}, nil
}
