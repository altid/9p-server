package main

import (
	"path"
	"io"
	"io/ioutil"
	"os"
)

type dir struct {
	c chan os.FileInfo
	done chan struct{}
}

func mkdir(filepath string) *dir {
	c := make(chan os.FileInfo, 10)
	done := make(chan struct{})
	list, err := ioutil.ReadDir(filepath)
	if err != nil {
		return nil
	}
	ctlfile, err := mkctl(getBase(path.Join(filepath, "ctl")))
	if err != nil {
		return nil
	}
	list = append(list, &ctlStat{name: "ctl", file: ctlfile })
	//list = append(list, &eventStat{}))
	go func() {
		for _, f := range list {
			select {
			case c <- f:
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
		fi = append(fi, s)
	}
	return fi, err
}

func (d *dir) Close() error {
	close(d.done)
	return nil
}