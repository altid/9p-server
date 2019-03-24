package main

import (
	"context"
	"log"
	"os"

	"github.com/docker/go-p9p"
)

var (
	current string
	polling map[uint32]bool
	last uint32
)

type msg struct {
	srv string
	msg string
}

type server struct {
	ctx     context.Context
	session p9p.Session
	pwd     string
	pwdfid  p9p.Fid
	rootfid p9p.Fid
	nextfid p9p.Fid
	done    chan struct{}
}

func init() {
	polling = make(map[uint32]bool)
	polling[0] = false
}

func main() {
	if len(os.Args) <= 1 {
		log.Fatalf("Usage: %s <service> [<service>...]\n", os.Args[0])
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	servlist := make(map[string]*server)
	for _, arg := range os.Args[1:] {
		c, err := attach(arg, ctx)
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
	handleMessage(servlist[current])
	input := readStdin(ctx)
	dispatch(servlist, input)
}
