package main

import (
	"context"
	"io"
	"time"

	"github.com/docker/go-p9p"
)

type content struct {
	buff []byte
	err  error
}

func writeFile(s *server, name string, data *content) error {
	ctx, _ := context.WithTimeout(s.ctx, 5*time.Second)
	tfid := s.nextfid
	s.nextfid++
	_, err := walkFile(ctx, s, tfid, name, "write")
	if err != nil {
		return err
	}
	defer s.session.Clunk(ctx, s.pwdfid)
	dirstat, err := s.session.Stat(ctx, tfid)
	end := int64(dirstat.Length)
	if err != nil {
		return err
	}
	data.buff = append(data.buff, '\n')
	_, err = s.session.Write(ctx, tfid, data.buff, end)
	if err != nil {
		return err
	}
	return nil
}

func readFile(s *server, name string, uuid uint32) (chan *content, error) {
	tfid := s.nextfid
	s.nextfid++
	iounit, err := walkFile(s.ctx, s, tfid, name, "read")
	if err != nil {
		return nil, err
	}
	polling[uuid] = (name == "feed")
	if uuid > 0 {
		polling[last] = false
	}
	m := make(chan *content)
	go func(m chan *content, iounit uint32, tfid p9p.Fid) {
		defer close(m)
		defer s.session.Clunk(s.ctx, s.pwdfid)
		var offset int64
		for {
			buff := make([]byte, iounit)
			n, err := s.session.Read(s.ctx, tfid, buff, offset)
			offset += int64(n)
			if n > 0 {
				m <- &content{
					buff: buff,
					err:  err,
				}
			}
			if err == io.EOF && polling[uuid] {
				time.Sleep(300 * time.Millisecond)
				err = nil
			}
			if err != nil {
				break
			}
		}
	}(m, iounit, tfid)
	return m, nil
}

func walkFile(ctx context.Context, s *server, tfid p9p.Fid, name string, rw string) (uint32, error) {
	if _, err := s.session.Walk(ctx, s.rootfid, tfid, name); err != nil {
		return 0, err
	}
	defer s.session.Clunk(s.ctx, s.pwdfid)
	var iounit uint32
	var err error
	switch rw {
	case "read":
		_, iounit, err = s.session.Open(ctx, tfid, p9p.OREAD)
	case "write":
		_, iounit, err = s.session.Open(ctx, tfid, p9p.OWRITE)
	default:
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if iounit < 1 {
		msize, _ := s.session.Version()
		iounit = uint32(msize - 24)
	}
	return iounit, nil
}
