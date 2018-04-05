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
}

func (f *ctlFile) Read(p []byte) (n int, err error) {
	// TODO: Currently blocking indefinitely
	var i int
	for i = 0; i < n; {
		s, ok := <-f.data
		if ! ok {
			return 0, io.EOF
		}
		i+=copy(p, s)
	}	
	return i, err
}

func (f *ctlFile) Write(p []byte) (int, error) {
	// TODO: Read in full input and switch out server commands
	// Pass through all other writes as a normal write to our type, in f.
	return 0, nil
}

func (f *ctlFile) Close() error {
	close(f.done)
	return nil
}

type ctlStat struct {
	name string
	file *ctlFile
	stat os.FileInfo
}

func (s *ctlStat) Name() string     { return s.name }
func (s *ctlStat) Sys() interface{} { return s.file }

func (s *ctlStat) ModTime() time.Time {
	return s.stat.ModTime()
}

func (s *ctlStat) IsDir() bool {
	return false
}

func (s *ctlStat) Mode() os.FileMode {
	return s.stat.Mode()
}

func (s *ctlStat) Size() int64 {
	return s.file.size
}

// This returns a ready rwc for future reads/writes
func mkctl(ctl string) (*ctlFile, error) {
	s, err := os.Stat(ctl)
	if err != nil {
		return nil, err
	}
	data := make(chan []byte)
	done := make(chan struct{})
	// TODO: Add our server-specific ctl data to this []byte
	buff, err := ioutil.ReadFile(ctl)
	if err != nil {
		return nil, err
	}
	// TODO: Add size of our server-specific ctl data to this total
	size := s.Size()
	go func() {
		for {
			select {
			case data <- buff:
			case <- done:
				break
			}
		}
		close(data)
	}()
	return &ctlFile{data: data, done: done, size: size}, nil
}
