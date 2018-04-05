package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type ctlFile struct {
	data chan []byte
	done chan struct{}
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

func (f *ctlFile) Write(p []byte) (int, error) {
	// Make sure we empty our chan
	<-f.data	
	data := string(p[:])
	switch {
	case strings.HasPrefix(data, "test"):
		fmt.Println("test found")
	}
	f.off = f.size
	f.modTime = time.Now().Truncate(time.Hour)
	return len(data), nil
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
	if f.off > f.size {
		f.off = f.size
	}
	return f.off, nil
}

func (f *ctlFile) Close() error {
	close(f.done)
	return nil
}

func (f *ctlFile) Uid() string { return f.uid }
func (f *ctlFile) Gid() string { return f.uid }
func (f *ctlFile) Muid() string { return f.uid }

func (f *ctlFile) Stat() os.FileInfo {
	return &ctlStat{ name: "ctl", file: f, }
}

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
func mkctl(ctl, uid string) (*ctlFile, error) {
	data := make(chan []byte)
	done := make(chan struct{})
	// TODO: Add our server-specific ctl data to this []byte
	final, err := ioutil.ReadFile(ctl)
	if err != nil {
		return nil, err
	}
	buff := []byte("Our initial server data\n")
	final = append(buff, final...)
	// TODO: Add size of our server-specific ctl data to this total
	size := len(final)
	go func() {
		LOOP:
		for {
			select {
			case data <- final:
			case <- done:
				break LOOP
			}
		}
		close(data)
	}()
	return &ctlFile{data: data, done: done, size: int64(size), off: 0, modTime: time.Now().Truncate(time.Hour), uid: uid}, nil
}
