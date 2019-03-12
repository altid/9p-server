package main

import (
	"bufio"
	"bytes"
	"context"
	"log"
	"net"
	"os"
	"os/user"
	//"strings"
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
	
func attach(srv string, ctx context.Context) (*server, error) {
	// Docker's lib doesn't do much heavy lifting for us
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	d := net.Dialer{}
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
	go tailEventsFile(s)
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
			if i[0] == '/' {
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
	d, err := s.session.Stat(ctx, targetfid)
	if err != nil {
		return err
	}
	buffer := bytes.NewBufferString(input)
	n := int64(d.Length) //potentially narrowing
	for buffer.Len() > 0 {
		b := buffer.Next(int(iounit))
		offset := len(b)
		_, err := s.session.Write(ctx, targetfid, b, n)
		if err != nil {
			return err
		}
		n+=int64(offset)
	}
	return err
	
}

func handleCtrl(srv map[string]*server, command string) {
	/*if strings.HasPrefix(current, command) {
		buff := strings.TrimPrefix(current, command)
      		if srv[buff] != nil {
			current = buff
			handleMessage(srv[buff], &msg{
				srv: buff,
				msg: "document",
			})
		}
	}*/
	// Handle buffer changes, etc; then send the message on down to 9p-server.
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

func tailEventsFile(s *server) {
	// 
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
		c, err := attach(arg, ctx)
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
