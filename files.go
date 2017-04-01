package ubqtlib

import (
	//"fmt"
	"io"
	"os"
	"time"
)

type fakefile struct {
	name         string
	client       string
	size         int64
	mode         os.FileMode
	handler      ClientHandler
	atime, mtime time.Time
}

func (f *fakefile) ReadAt(p []byte, off int64) (int, error) {
	c, err := f.handler.ClientRead(f.name, f.client)
	if err != nil {
		return 0, err
	}
	n := copy(p, c[off:])
	return n, err
}

func (f *fakefile) WriteAt(p []byte, off int64) (int, error) {
	return f.handler.ClientWrite(f.name, f.client, p)
}

func (f *fakefile) Size() int64 {
	return f.size
}

func (f *fakefile) Name() string     { return f.name }
func (f *fakefile) Sys() interface{} { return f }

func (f *fakefile) ModTime() time.Time {
	if f.mtime.IsZero() {
		f.mtime = time.Now()
		f.atime = time.Now()
	}
	return f.mtime
}

// We have only one directory, so return that
func (f *fakefile) IsDir() bool {
	return (f.name == "/")
}

// Again, only root directory so we can safely optimize
func (f *fakefile) Mode() os.FileMode {
	switch f.name {
	case "/":
		return os.ModeDir | 0755
	}
	return 0666
}

type dir struct {
	c    chan fakefile
	done chan struct{}
}

func mkdir(fi Client) *dir {
	c := make(chan fakefile, 10)
	done := make(chan struct{})
	go func() {
		for _, file := range fi {
			select {
			case c <- *file:
			case <-done:
				break
			}
		}
		close(c)
	}()
	return &dir{
		c:    c,
		done: done,
	}
}

// This is fine for our needs
func (d *dir) Readdir(n int) ([]os.FileInfo, error) {
	var err error
	fi := make([]os.FileInfo, 0, 10)
	for i := 0; i < n; i++ {
		s, ok := <-d.c
		if !ok {
			err = io.EOF
			break
		}
		fi = append(fi, &s)
	}
	return fi, err
}

func (d *dir) Close() error {
	close(d.done)
	return nil
}
