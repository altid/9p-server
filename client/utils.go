package main

import (
	"bufio"
	"context"
	"log"
	"net"
	"os"
	"os/user"
	"strings"

	"github.com/docker/go-p9p"
)

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
		done: make(chan struct{}),
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

func dispatch(srv map[string]*server, events chan *msg, input chan string) {
	for {
		select {
		case i := <-input:
			if i == "/quit" {
				return
			}
			if i == "/tabs" {
				handleTabs(srv)
				continue
			}
			if i[0] == '/' && len(i) > 1 {
				handleCtrl(srv, i[1:])
				continue
			}
			handleInput(srv[current], i)
		
		case event := <-events:
			log.Println(event)
			if event.srv == current {
				err := handleMessage(srv[current], event)
				if err != nil {
					log.Print(err)
				}
			}
		}
	}
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

func sendEvents(s *server, name string, event chan *msg) {
	data, err := readEvents(s)
	if err != nil {
		log.Print(err)
		return
	}
	for m := range data {
		if m.err != nil {
			log.Print(err)
			break
		}
		event <- &msg{
			msg: string(m.buff),
			srv: name,
		}
	}
	log.Print("Ending events loop")
}
