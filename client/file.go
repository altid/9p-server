package main

import(
	"context"
	"io"
	"log"
	"time"

	"github.com/docker/go-p9p"
)

type content struct {
	buff []byte
	err  error
}

func writeFile(s *server, name string, data chan *content, offset int64) error {
	ctx, _ := context.WithTimeout(s.ctx, 5*time.Second)
	tfid := s.nextfid
	s.nextfid++
	iounit, err := walkFile(ctx, s, tfid, name, "write")
	if err != nil {
		return err
	}
	for m := range data {
		if m.err != nil {
			break
		}
		// TODO: Handle large writes here
		n, err := s.session.Write(ctx, tfid, m.buff[:iounit], offset)
		if err != nil {
			return err
		}
		offset += int64(n)
	}
	defer s.session.Clunk(ctx, s.pwdfid)
	return nil
}
func readEvents(s *server) (chan *content, error) {
	ctx, _ := context.WithTimeout(s.ctx, 5*time.Second)
	tfid := s.nextfid
	s.nextfid++
	iounit, err := walkFile(ctx, s, tfid, "event", "read")
	if err != nil {
		return nil, err
	}
	m := make(chan *content)
	go func(m chan *content, iounit uint32, tfid p9p.Fid) {
		defer close(m)
		defer s.session.Clunk(ctx, s.pwdfid)
		var offset int64
		for {
			buff := make([]byte, iounit)
			n, err := s.session.Read(ctx, tfid, buff, offset)
			switch err {
			case io.EOF, context.DeadlineExceeded:
				time.Sleep(300 * time.Millisecond)
				continue
			default:
				log.Print(err)
				return
			}
			if n > 0 {
				m <- &content{
					buff: buff,
					err: err,
				}
			}
			offset += int64(n)
		}
	}(m, iounit, tfid)
	return m, nil
}
		
func readFile(s *server, name string) (chan *content, error) {
	ctx, _ := context.WithTimeout(s.ctx, 5*time.Second)
	tfid := s.nextfid
	s.nextfid++
	iounit, err := walkFile(ctx, s, tfid, name, "read")
	if err != nil {
		return nil, err
	}
	m := make(chan *content)
	go func(m chan *content, iounit uint32, tfid p9p.Fid) {
		defer close(m)
		defer s.session.Clunk(ctx, s.pwdfid)
		var offset int64
		for {
			buff := make([]byte, iounit)
			n, err := s.session.Read(ctx, tfid, buff, offset)
			if err != nil {
				break
			}
			if n > 0 {
				m <- &content{
					buff: buff[:n],
					err: err,
				}
			}
			offset += int64(n)

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

