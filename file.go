package main

import (
	"io"
	"io/ioutil"
	"os"
)

type dir struct {
	c	  chan os.FileInfo
	close chan struct {}
}

func mkdir(path string) *dir {
	c := make(chan stat, 10)
	done := make(chan struct{})
	list := ioutil.ReadDir(path)
	// TODO: Investigate how to generate stats for each type
	ctl := &ctlFile{ ... }
	event := &eventFile{ ... }
	append(list, ctl.Stat())
	append(list, event.Stat())
	go func() {
		for f := range list {
			select {
			case c <- f
			case <- done:
				break
			}
		}
		close(c)
	}()
	return &dir{ c: c, done: done, }
}

// Listen for os.FileInfo members to come in from mkdir
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