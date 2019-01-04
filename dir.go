package main

import (
	"log"
	"path"
	"io"
	"io/ioutil"
	"os"
	"time"
)

type dir struct {
	c chan os.FileInfo
	done chan struct{}
}

func mkdir(filepath, uid string, client *Client) *dir {
	c := make(chan os.FileInfo, 10)
	done := make(chan struct{})
	list, err := ioutil.ReadDir(filepath)
	if err != nil {
		return nil
	}
	ctlfile, err := mkctl(getBase(path.Join(filepath, "ctrl")), uid, client)
	if err != nil {
		return nil
	}
	list = append(list, &ctlStat{name: "ctrl", file: ctlfile })
	eventfile, err := mkevent(uid, client)
	if err != nil {
		return nil
	}
	list = append(list, &eventStat{name: "event", file: eventfile})
	go func([]os.FileInfo) {
		for _, f := range list {
			log.Print(f.Name() + "\n")
			select {
			case c <- f:
			case <- done:
				break
			}
		}
		close(c)
	}(list)
	return &dir{ c: c, done: done, }
}

func (d *dir) IsDir() bool {
	return true
}

func (d *dir) ModTime() time.Time {
	return time.Now()
}

func (d *dir) Mode() os.FileMode {
	return os.ModeDir
}

func (d *dir) Name() string {
	return "/"
}

func (d *dir) Size() int64 {
	return 0
}

func (d *dir) Sys() interface{} {
	return nil
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
