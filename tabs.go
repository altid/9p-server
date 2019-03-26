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
	data    string
	size    int64
	off     int64
	uid     string
	modtime time.Time
}

func (f *tabs) Read(b []byte) (n int, err error) {
	n = copy(b, f.data[f.off:])
	if int64(n)+f.off >= f.size {
		return n, io.EOF
	}
	return n, nil
}

func (f *tabs) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.off = offset
	case io.SeekCurrent:
		f.off += offset
	case io.SeekEnd:
		f.off = f.size + offset
	}
	if f.off < 0 {
		return 0, fmt.Errorf("No seeking past start of file")
	}
	if f.off > f.size {
		return f.size, fmt.Errorf("No seeking past end of file")
	}
	return f.off, nil
}

func (f *tabs) Close() error {
	f.data = ""
	return nil
}

func (f *tabs) Uid() string { return f.uid }
func (f *tabs) Gid() string { return f.uid }

type tabsStat struct {
	name string
	file *tabs
}

func (s *tabsStat) Name() string       { return s.name }
func (s *tabsStat) Sys() interface{}   { return s.file }
func (s *tabsStat) ModTime() time.Time { return s.file.modtime }
func (s *tabsStat) IsDir() bool        { return false }
func (s *tabsStat) Mode() os.FileMode  { return 0644 }
func (s *tabsStat) Size() int64        { return s.file.size }

// BUGS(halfwit): The current implementation doesn't always correctly mark a given clients' current buffer as purple
// https://github.com/ubqt-systems/9p-server/issues/12
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
		if !ok {
			code = "grey"
		}
		msg := path.Base(line)
		color, _ := cleanmark.NewColor(code, []byte(msg))
		data += fmt.Sprintf("%s\n", color)
	}
	t := &tabs{
		data:    data,
		uid:     uid,
		size:    int64(len(data)),
		modtime: time.Now().Truncate(time.Hour),
	}
	return t, nil
}
