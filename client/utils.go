package main

import (
	"bufio"
	"context"
	"net"
	"os"
	"os/user"
	"strings"

	"github.com/docker/go-p9p"
)

func attach(srv string, ctx context.Context) (*server, error) {
	// Docker's lib doesn't do much heavy lifting for us
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	d := net.Dialer{}
	if !strings.Contains(srv, ":") {
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
		ctx:     ctx,
		pwd:     "/",
		nextfid: 1,
		session: session,
		done:    make(chan struct{}),
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
	return s, nil
}

func dispatch(srv map[string]*server, input chan string) {
	for i := range input {
		switch i {
		case "/quit":
			return
		case "/tabs":
			handleTabs(srv)
		case "/status":
			handleStatus(srv[current])
		case "/sidebar":
			handleSide(srv[current])
		case "/title":
			handleTitle(srv[current])
		default:
			if i[0] == '/' && len(i) > 1 {
				handleCtrl(srv, i[1:])
				continue
			}
			handleInput(srv[current], i)
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
