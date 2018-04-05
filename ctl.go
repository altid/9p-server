package main

import (
	"io"
	"io/ioutil"
	"os"
	"time"
)

type ctlFile struct {
	data chan []byte
	done chan struct{}
	size int64
	off  int64
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
	// TODO: Read in full input and switch out server commands
	// Pass through all other writes as a normal write to our type, in f.
	return len(p), nil
}

func (f *ctlFile) Seek(offset int64, whence int) (int64, error) {
	if offset > f.size {
		return 0, io.EOF
	}
	switch whence {
	case io.SeekStart:
		f.off = offset
	case io.SeekCurrent:
		f.off += offset
	case io.SeekEnd:
		if offset > 0 {
			return 0, io.EOF
		}
		f.off = f.size + offset
	}
	return f.off, nil
}

func (f *ctlFile) Close() error {
	close(f.done)
	return nil
}

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
	return time.Now().Truncate(time.Hour)
}

func (s *ctlStat) IsDir() bool {
	return false
}

func (s *ctlStat) Mode() os.FileMode {
	return 0644
}

func (s *ctlStat) Size() int64 {
	return s.file.size
}

// This returns a ready rwc for future reads/writes
func mkctl(ctl string) (*ctlFile, error) {
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
	return &ctlFile{data: data, done: done, size: int64(size), off: 0}, nil
}
