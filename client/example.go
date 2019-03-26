package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/docker/go-p9p"
)

var (
	scrollback = flag.Uint64("s", 8000, "Characters of scrollback in feed files")
	current    string
	polling    map[uint32]bool
	last       uint32
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
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(0)
	}
	if flag.NArg() < 1 {
		log.Fatalf("Usage: %s <service> [<service>...]\n", flag.Arg(0))
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	servlist := make(map[string]*server)
	for _, arg := range flag.Args() {
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
