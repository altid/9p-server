package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/ubqt-systems/cleanmark"
)

type tabs struct {
	data string
	size int64
	uid  string
}

func (f *tabs) ReadAt(b []byte, off int64) (n int, err error) {
	n = copy(b, f.data[off:])
	if int64(n)+off > f.size {
		return n, io.EOF
	}
	return n, nil
}

func (f *tabs) Close() error { return nil }
func (f *tabs) Uid() string { return f.uid }
func (f *tabs) Gid() string { return f.uid }

type tabsStat struct {
	name string
	file *tabs
}

func (s *tabsStat) Name() string       { return s.name }
func (s *tabsStat) Sys() interface{}   { return s.file }
func (s *tabsStat) ModTime() time.Time { return time.Now() }
func (s *tabsStat) IsDir() bool        { return false }
func (s *tabsStat) Mode() os.FileMode  { return 0644 }
func (s *tabsStat) Size() int64        { return s.file.size }

func mktabs(tab, uid string, cl *client) (*tabs, error) {
	var data string
	tf, err := os.Open(tab)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(tf)
	for scanner.Scan() {
		line := scanner.Text()
		code, ok := cl.tabs[path.Join(cl.service, line)]
		if ! ok {
			code = "grey"
		}
		msg := path.Base(line)
		color, _ := cleanmark.NewColor(code, []byte(msg))
		data += fmt.Sprintf("%s\n", color)
	}
	t := &tabs{
		data: data,
		uid:  uid,
		size: int64(len(data)),
	}
	return t, nil
}
