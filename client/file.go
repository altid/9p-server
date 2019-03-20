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

func writeFile(s *server, name string, data *content, offset int64) error {
	ctx, _ := context.WithTimeout(s.ctx, 5*time.Second)
	tfid := s.nextfid
	s.nextfid++
	//iounit, err := walkFile(ctx, s, tfid, name, "write")
	_, err := walkFile(ctx, s, tfid, name, "write")
	if err != nil {
		return err
	}
	defer s.session.Clunk(ctx, s.pwdfid)
	//while len(data.buff - offset) > iounit {
		//n, err := s.session.Write(ctx, tfid, data.buff[:iounit], offset)
		//offset += int64(n)
	//}
	_, err = s.session.Write(ctx, tfid, data.buff, offset)
	if err != nil {
		return err
	}
	return nil
}

// Hold our offset and try to read every 300ms
// read will EOF, so sleep and do it again
func readEvents(s *server) (chan *content, error) {
	ctx, _ := context.WithTimeout(s.ctx, 5*time.Second)
	tfid := s.nextfid
	s.nextfid++
	iounit, err := walkFile(ctx, s, tfid, "event", "read")
	if err != nil {
		return nil, err
	}
	m := make(chan *content)
	go func(m chan *content, tfid p9p.Fid) {
		defer close(m)
		defer s.session.Clunk(ctx, s.pwdfid)
		var offset int64
		b := make([]byte, iounit)
		for {
			n, err := s.session.Read(s.ctx, tfid, b, offset)
			switch err {
			case io.EOF:
				time.Sleep(500 * time.Millisecond)
			case context.DeadlineExceeded:
				log.Println(err)
			default:
				log.Print(err)
				return
			}
			log.Println(b)
			if n > 0 {
				m <- &content{
					buff: b,
					err: nil,
				}
				offset += int64(n)
			}
		}
	}(m, tfid)
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
					buff: buff,
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

