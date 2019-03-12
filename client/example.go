package main


// TODO: We'll have to switch to dockec/gop9p
// Lionkov's is broken.
import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/lionkov/go9p/p/clnt"
	"github.com/lionkov/go9p/p"
)

var current string

type msg struct {
	srv string
	msg string
}

func attach(srv string, ctx context.Context) (*clnt.Clnt, error) {
	user := p.OsUsers.Uid2User(os.Geteuid())
	d := net.Dialer{}
	c, err := d.DialContext(ctx, "tcp", srv)
	if err != nil {
		return nil, err
	}
	return clnt.MountConn(c, "", 8192, user)
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
				log.Println(input)
			}
		}
	}(ctx, input)
	return input
}

func dispatch(srv map[string]*clnt.Clnt, events chan *msg, input chan string) {
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
		case e := <-events:
			if e.srv == current {
				handleMessage(srv[current], e)
			}
		}
	}
}

func handleInput(srv *clnt.Clnt, input string) {
	f, err := srv.FOpen("/input", p.OWRITE)
	if err != nil {
		log.Print(err)
		return
	}
	defer f.Close()
	f.Write([]byte(input))
}

func handleCtrl(srv map[string]*clnt.Clnt, command string) {
	if strings.HasPrefix(current, command) {
		buff := strings.TrimPrefix(current, command)
      		if srv[buff] != nil {
			current = buff
			handleMessage(srv[buff], &msg{
				srv: buff,
				msg: "document",
			})
		}
	}
}

func handleMessage(srv *clnt.Clnt, m *msg) {
	if m.srv == current && m.msg == "document" {
		f, err := srv.FOpen("/document", p.OREAD)
		if err != nil {
			log.Print(err)
			return
		}
		defer f.Close()
		data, err := ioutil.ReadAll(f)
		if err != nil {
			log.Print(err)
			return
		}
		fmt.Println(data)
	}
}

func main() {
	if len(os.Args) <= 1 {
		log.Fatalf("Usage: %s <service> [<service>...]\n", os.Args[0])
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	servlist := make(map[string]*clnt.Clnt)
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
	for _, conn := range servlist {
		conn.Unmount()
	}
}
