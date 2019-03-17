package main

import (
	"context"
	"log"
	"os"
	
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
	done chan struct{}
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
		servlist[arg] = c
		current = arg
	}
	if len(servlist) < 1 {
		log.Fatal("Unable to connect")
	}
	tries := []string{
		"document",
		"feed",
		"stream",
	}
	for _, i := range tries {
		err := handleMessage(servlist[current], &msg{
			srv: current,
			msg: i,
		})
		if err == nil {
			break
		}
	}
	input := readStdin(ctx)
	dispatch(servlist, events, input)
}
