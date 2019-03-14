package main

import (
	"bufio"
	"bytes"
	"context"
	"log"
	"net"
	"os"
	"os/user"
	"strings"
	"time"
	
	"github.com/docker/go-p9p"
)

var current string

type msg struct {
	srv string
	msg string
}

type server struct {
	ctx context.Context
	session p9p.Session
	pwd string
	pwdfid p9p.Fid
	rootfid p9p.Fid
	nextfid p9p.Fid
}
	
func attach(srv string, ctx context.Context, event chan *msg) (*server, error) {
	// Docker's lib doesn't do much heavy lifting for us
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	d := net.Dialer{}
	if ! strings.Contains(srv, ":") {
		srv += ":564"
	}
	c, err := d.DialContext(ctx, "tcp", srv)
	if err != nil {
		return nil, err
	}
	session, err := p9p.NewSession(ctx, c)
	if err != nil {
		return nil, err
	}
	s := &server{
		ctx: ctx,
		pwd: "/",
		nextfid: 1,
		session: session,
	}
	if _, err := s.session.Attach(s.ctx, s.nextfid, p9p.NOFID, usr.Username, "/"); err != nil {
		return nil, err
	}
	s.rootfid = s.nextfid
	s.nextfid++
	if _, err := s.session.Walk(s.ctx, s.rootfid, s.nextfid); err != nil {
		return nil, err
	}
	s.pwdfid = s.nextfid
	s.nextfid++
	go sendEvents(s, srv, event)
	return s, nil
}

func readStdin(ctx context.Context) chan string {
	input := make(chan string)
	go func(ctx context.Context, input chan string) {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case input <- scanner.Text():
			}
		}
	}(ctx, input)
	return input
}

func dispatch(srv map[string]*server, events chan *msg, input chan string) {
	for {
		select {
		case i := <-input:
			if i == "/quit" {
				return
			}
			if i[0] == '/' && len(i) > 1 {
				handleCtrl(srv, i[1:])
				continue
			}
			handleInput(srv[current], i)
		
		case event := <-events:
			if event.srv == current {
				err := handleMessage(srv[current], event)
				if err != nil {
					log.Print(err)
				}
			}
		}
	}
}

// BUG(halfwit): Writes are not making it to the underlying service
// We'll have to do quite a lot of debugging to see what is not happening
func handleInput(s *server, input string) error {
	ctx, _ := context.WithTimeout(s.ctx, 5*time.Second)
	targetfid := s.nextfid
	s.nextfid++
	if _, err := s.session.Walk(ctx, s.rootfid, targetfid, "input"); err != nil {
		return err
	}
	defer s.session.Clunk(s.ctx, s.pwdfid)
	_, iounit, err := s.session.Open(ctx, targetfid, p9p.OWRITE)
	if err != nil {
		return err
	}
	if iounit < 1 {
		msize, _ := s.session.Version()
		iounit = uint32(msize - 24)
	}
	buffer := bytes.NewBufferString(input)
	var n int64
	for buffer.Len() > 0 {
		b := buffer.Next(int(iounit))
		offset := len(b)
		_, err := s.session.Write(ctx, targetfid, b, n - 1)
		if err != nil {
			return err
		}
		s.session.Write(ctx, targetfid, []byte("\n"), n)
		n+=int64(offset)
	}
	return err
	
}

func handleCtrl(srv map[string]*server, command string) error {
	log.Println(command)
	if strings.HasPrefix("service ", command) {
		buff := strings.TrimPrefix("service ", command)
      		if srv[buff] != nil {
			current = buff
			handleMessage(srv[buff], &msg{
				srv: buff,
				msg: "document",
			})
		}
		return nil
	}
	s := srv[current]
	ctx, _ := context.WithTimeout(s.ctx, 5*time.Second)
	targetfid := s.nextfid
	s.nextfid++
	if _, err := s.session.Walk(ctx, s.rootfid, targetfid, "ctrl"); err != nil {
		return err
	}
	defer s.session.Clunk(ctx, s.pwdfid)
	_, iounit, err := s.session.Open(ctx, targetfid, p9p.OWRITE)
	if err != nil {
		return err
	}
	if iounit < 1 {
		msize, _ := s.session.Version()
		iounit = uint32(msize - 24)
	}
	buffer := bytes.NewBufferString(command)
	var n int64
	for buffer.Len() > 0 {
		b := buffer.Next(int(iounit))
		offset := len(b)
		_, err := s.session.Write(ctx, targetfid, b, n)
		if err != nil {
			return err
		}
		s.session.Write(ctx, targetfid, []byte("\n"), n)
		s.session.Clunk(ctx, s.pwdfid)
		n+=int64(offset)
	}
	return err
}

func handleMessage(s *server, m *msg) error {
	if m.srv != current || m.msg != "document" {
		return nil
	}
	ctx, _ := context.WithTimeout(s.ctx, 5*time.Second)
	targetfid := s.nextfid
	s.nextfid++
	if _, err := s.session.Walk(ctx, s.rootfid, targetfid, "document"); err != nil {
		return err
	}
	defer s.session.Clunk(s.ctx, s.pwdfid)
	_, iounit, err := s.session.Open(ctx, targetfid, p9p.OREAD)
	if err != nil {
		return err
	}
	if iounit < 1 {
		msize, _ := s.session.Version()
		iounit = uint32(msize - 24)
	}
	b := make([]byte, iounit)
	n, err := s.session.Read(ctx, targetfid, b, 0)
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write(b[:n]); err != nil {
		return err
	}
	os.Stdout.Write([]byte("\n"))
	return nil
}

type eventReader struct {
	ctx context.Context
	session p9p.Session
	tfid p9p.Fid
	offset int64
}

// BUG(halfwit): Currently no events are being read
// It is very likely this is due to the tailing implementation
// not working well with go-p9p
func (e *eventReader) Read(p []byte) (n int, err error) {
	return e.session.Read(e.ctx, e.tfid, p, e.offset)
}

func sendEvents(s *server, name string, event chan *msg) {
	ctx, _ := context.WithTimeout(s.ctx, 5*time.Second)
	targetfid := s.nextfid
	s.nextfid++
	if _, err := s.session.Walk(ctx, s.rootfid, targetfid, "event"); err != nil {
		log.Print(err)
		return 
	}
	defer s.session.Clunk(s.ctx, s.pwdfid)
	_, _, err := s.session.Open(ctx, targetfid, p9p.OREAD)
	if err != nil {
		log.Print(err)
		return
	}
	e := &eventReader{
		session: s.session,
		ctx: s.ctx,
		tfid: targetfid,
	}
	scanner := bufio.NewScanner(e)
	for scanner.Scan() {
		txt := scanner.Text()
		event <- &msg{
			msg: txt,
			srv: name,
		}
		e.offset += int64(len(txt))
	}
	
}

func main() {
	if len(os.Args) <= 1 {
		log.Fatalf("Usage: %s <service> [<service>...]\n", os.Args[0])
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	servlist := make(map[string]*server)
	events := make(chan *msg)
	for _, arg := range os.Args[1:] {
		c, err := attach(arg, ctx, events)
		if err != nil {
			log.Print(err)
			continue
		}
		log.Print("successfully added server")
		servlist[arg] = c
		current = arg
	}
	if len(servlist) < 1 {
		log.Fatal("Unable to connect")
	}
	handleMessage(servlist[current], &msg{
		srv: current,
		msg: "document",
	})
	input := readStdin(ctx)
	dispatch(servlist, events, input)
}
