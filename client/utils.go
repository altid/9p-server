package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/user"
	"strings"

	"github.com/docker/go-p9p"
	"github.com/ubqt-systems/cleanmark"
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
	handleTab(srv, s)
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

func clean(m *content) []byte {
	l := cleanmark.NewLexer(m.buff)
	var dst bytes.Buffer
	for {
		i := l.Next()
		switch i.ItemType {
		case cleanmark.EOF:
			return dst.Bytes()
		case cleanmark.ColorCode, cleanmark.ImagePath:
			continue
		case cleanmark.UrlLink, cleanmark.ImageLink:
			s := fmt.Sprintf(" (%s) ", i.Data)
			dst.WriteString(s)
		default:
			dst.Write(i.Data)
		}
	}
	return dst.Bytes()
}

func buildCtlMsg(s *server, action, content string) ([]byte, error) {
	var buff bytes.Buffer
	buff.WriteString(action + " ")
	buff.WriteString(s.current + " ")
	buff.WriteString(content)
	return buff.Bytes(), nil
}

func split(in string) (action, content string) {
	token := strings.Fields(in)
	return token[0], strings.Join(token[1:], " ")
}

func tabs(s *server, b []byte, srv string) []byte {
	var dst, last []byte
	l := cleanmark.NewLexer(b)
	for {
		i := l.Next()
		switch i.ItemType {
		case cleanmark.EOF:
			return dst
		case cleanmark.ColorCode:
			switch string(i.Data) {
			case cleanmark.Red:
				dst = append(dst, '!')
			case cleanmark.Blue:
				dst = append(dst, '+')
			case cleanmark.Purple:
				if srv == current {
					dst = append(dst, '*')
				}
				s.current = string(last)
			}
			if srv != current {
				dst = append(dst, srv...)
				dst = append(dst, '/')
			}
			dst = append(dst, last...)
			dst = append(dst, ' ')
		case cleanmark.ColorText:
			last = i.Data
		}
	}
	return dst
}

