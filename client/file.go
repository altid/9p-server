package main

import(
	"context"
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
	//iounit, err := walkFile(ctx, s, tfid, name, "write")
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
	//while len(data.buff - offset) > iounit {
		//n, err := s.session.Write(ctx, tfid, data.buff[:iounit], end)
		//offset += int64(n)
	//}
	data.buff = append(data.buff, '\n')
	_, err = s.session.Write(ctx, tfid, data.buff, end)
	if err != nil {
		return err
	}
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
	go func(m chan *content, tfid p9p.Fid) {
		defer close(m)
		defer s.session.Clunk(ctx, s.pwdfid)
		var offset int64
		var n int = 1
		for ;; offset += int64(n) {
			b := make([]byte, iounit)
			n, err = s.session.Read(s.ctx, tfid, b, offset)
			if err != nil {
				time.Sleep(500 * time.Millisecond)
				d, _ := s.session.Stat(s.ctx, tfid)
				offset = int64(d.Length)
			}
			if n > 0 {
				m <- &content{
					buff: b,
					err: nil,
				}

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

