package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"runtime"

	"aqwari.net/net/styx"
	"bitbucket.org/mischief/libauth"
	"github.com/google/uuid"
)

type client struct {
	buffer  string
	service string
	event   chan string
	done    chan struct{}
	tabs    map[string]string
	polling bool
}

type server struct {
	c       map[uuid.UUID]*client
	l       net.Listener
	service string
}

func newServer(addr, service string) (*server, error) {
	// Bit of magic here - if we get a good port # then we know addr is fine
	// If we don't, we know addr is only the host name
	// So we just tag on the port and start the listeners
	_, port, _ := net.SplitHostPort(addr)
	if port == "" {
		port = "564"
		addr = net.JoinHostPort(addr, port)
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	srv := &server{
		c: make(map[uuid.UUID]*client),

		l:       l,
		service: path.Base(service),
	}
	if *useTLS {
		var certificate tls.Certificate
		switch runtime.GOOS {
		case "plan9":
			c, err := ioutil.ReadFile(*cert)
			if err != nil {
				log.Fatal(err)
			}
			keys, err := libauth.Listkeys()
			for _, k := range keys {
				pk := x509.MarshalPKCS1PublicKey(&k)
				certificate, err = tls.X509KeyPair(c, pk)
				if err == nil {
					break
				}
			}
		default:
			certificate, err = tls.LoadX509KeyPair(*cert, *key)
			if err != nil {
				log.Fatal(err)
			}
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{
				certificate,
			},
			ServerName: addr,
		}
		tlsConfig.BuildNameToCertificate()
		srv.l = tls.NewListener(l, tlsConfig)
	}
	return srv, nil
}

func (srv *server) newClient(service string) (*client, uuid.UUID) {
	cid := uuid.New()
	buffer := defaultBuffer(service)
	ch := make(chan string)
	dn := make(chan struct{})
	tabs := make(map[string]string)
	tabs[buffer] = "purple"
	srv.c[cid] = &client{
		buffer:  buffer,
		service: service,
		event:   ch,
		done:    dn,
		tabs:    tabs,
		polling: false,
	}
	return srv.c[cid], cid
}

func walkTo(c *client, req string, uid string) (os.FileInfo, string, error) {
	fp := path.Join(c.buffer, req)
	switch req {
	case "/":
		stat, err := os.Lstat(c.buffer)
		return stat, fp, err
	case "/ctrl":
		clientCtl := getBase(fp)
		ctlfile, err := mkctl(clientCtl, uid, c)
		if err != nil {
			log.Print(err)
			return nil, fp, err
		}
		cs := &ctlStat{
			name: "ctrl",
			file: ctlfile,
		}
		return cs, clientCtl, nil
	case "/event":
		clientEvent := getBase(fp)
		eventfile, err := mkevent(uid, c)
		if err != nil {
			log.Print(err)
			return nil, fp, err
		}
		es := &eventStat{
			name: "event",
			file: eventfile,
		}
		return es, clientEvent, nil
	case "/tabs":
		clientTabs := getBase(fp)
		tabsfile, err := mktabs(clientTabs, uid, c)
		if err != nil {
			log.Print(err)
			return nil, fp, err
		}
		ts := &tabsStat{
			name: "tabs",
			file: tabsfile,
		}
		return ts, clientTabs, nil
	default:
		stat, err := os.Lstat(fp)
		// If we have an error here, try to get a base-level stat.
		if err != nil {
			clientFp := getBase(fp)
			stat, err := os.Lstat(clientFp)
			return stat, clientFp, err
		}
		return stat, fp, nil
	}
}

// Called when a client connects
func (srv server) Serve9P(s *styx.Session) {
	client, uuid := srv.newClient(path.Join(*inpath, srv.service))
	defer delete(srv.c, uuid)
	defer close(client.done)
	defer close(client.event)
	for s.Next() {
		req := s.Request()
		stat, fp, err := walkTo(client, req.Path(), s.User)
		if err != nil {
			req.Rerror("%s", err)
			continue
		}
		switch t := req.(type) {
		case styx.Twalk:
			t.Rwalk(stat, nil)
		case styx.Tstat:
			t.Rstat(stat, nil)
		case styx.Topen:
			switch t.Path() {
			case "/":
				t.Ropen(mkdir(fp, s.User, client), nil)
			case "/event":
				t.Ropen(mkevent(s.User, client))
			case "/ctrl":
				t.Ropen(mkctl(fp, s.User, client))
			case "/tabs":
				t.Ropen(mktabs(fp, s.User, client))
			case "/input":
				t.Ropen(os.OpenFile(fp, os.O_RDWR, 0755))
			default:
				if stat.IsDir() {
					t.Ropen(ioutil.ReadDir(fp))
					continue
				}
				t.Ropen(os.OpenFile(fp, t.Flag, stat.Mode()))
			}
		case styx.Tutimes:
			switch t.Path() {
			case "/", "/event", "/ctrl", "/tabs":
				t.Rutimes(nil)
			default:
				t.Rutimes(os.Chtimes(fp, t.Atime, t.Mtime))
			}
		case styx.Ttruncate:
			switch t.Path() {
			case "/", "/event", "/ctrl", "/tabs":
				t.Rtruncate(nil)
			default:
				t.Rtruncate(os.Truncate(fp, t.Size))
			}
		// When clients are done with a notification, they delete it. Allow this
		case styx.Tremove:
			switch t.Path() {
			case "/notification", "/notify":
				t.Rremove(os.Remove(fp))
			default:
				t.Rerror("%s", "permission denied")
			}
		}
	}
}
