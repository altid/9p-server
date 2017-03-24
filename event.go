package ubqtlib

import (
	"os"
	"time"
)

//A read on "event" will block indefinitely, recieving lines for each SendEvent the client issues.
//instigating a read will register a listener for messages from SendEvent, and Close()ing the file will unregister
type event struct {
	client string
	mtime  time.Time
	wait   chan string
	off    int64
}

func (e *event) Size() int64        { return e.off }
func (e *event) Name() string       { return "event" }
func (e *event) ModTime() time.Time { return e.mtime }
func (e *event) Mode() os.FileMode  { return 0400 }
func (e *event) IsDir() bool        { return false }
func (e *event) Sys() interface{}   { return nil }

//TODO: We will have to figure out the proper way to do this, with 9p.
//direct_io flag is set
