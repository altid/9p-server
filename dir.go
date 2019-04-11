package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

type dir struct {
	name string
	c    chan os.FileInfo
	done chan struct{}
}

func mkdir(fp, uid string, cl *client) *dir {
	c := make(chan os.FileInfo, 10)
	done := make(chan struct{})
	list, err := ioutil.ReadDir(fp)
	if err != nil {
		return nil
	}
	ctlfile, err := mkctl(getBase(path.Join(fp, "ctrl")), uid, cl)
	if err != nil {
		log.Println(err)
		return nil
	}
	list = append(list, &ctlStat{name: "ctrl", file: ctlfile})
	eventfile, err := mkevent(uid, cl)
	if err != nil {
		log.Println(err)
		return nil
	}
	list = append(list, &eventStat{name: "event", file: eventfile})
	tabsfile, err := mktabs(getBase(path.Join(fp, "tabs")), uid, cl)
	if err != nil {
		log.Println(err)
		return nil
	}
	list = append(list, &tabsStat{name: "tabs", file: tabsfile})
	go func([]os.FileInfo) {
		for _, f := range list {
			select {
			case c <- f:
			case <-done:
				break
			}
		}
		close(c)
	}(list)
	return &dir{
		c:    c,
		done: done,
		name: fp,
	}
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
	return d.name
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
