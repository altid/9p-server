package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
)

// TODO(halfwit): Return any errors on an error chan here, or nil
func handleCtrl(srv map[string]*server, command string) error {
	if strings.HasPrefix(command, "buffer ") {
		buffer := strings.TrimPrefix(command, "buffer ")
		parts := strings.Split(buffer, "/")
		if _, ok := srv[parts[0]]; ok && len(parts) > 1 {
			current = parts[0]
			command = "buffer " + strings.Join(parts[1:], "/")
		}
		defer handleMessage(srv[current])
	}
	s := srv[current]
	data := &content{
		buff: []byte(command),
		err:  nil,
	}
	return writeFile(s, "ctrl", data)
}

func handleInput(s *server, input string) error {
	data := &content{
		buff: []byte(input),
		err:  nil,
	}
	return writeFile(s, "input", data)

}

func handleMessage(s *server) {
	id := uuid.New().ID()
	polling[id] = false
	for _, i := range []string{
		"document",
		"feed",
		"stream",
	} {
		go func(i string, id uint32) {
			data, err := readFile(s, i, id)
			if err != nil {
				return
			}
			last = id
			for m := range data {
				if m.err != nil {
					return
				}
				os.Stdout.Write(clean(m))
			}
		}(i, id)
	}
}

func handleStatus(srv *server) error {
	data, err := readFile(srv, "status", 0)
	if err != nil {
		return err
	}
	for m := range data {
		if _, err := os.Stdout.Write(clean(m)); err != nil {
			return err
		}
	}
	return nil
}

func handleTitle(srv *server) error {
	data, err := readFile(srv, "title", 0)
	if err != nil {
		return err
	}
	for m := range data {
		if _, err := os.Stdout.Write(clean(m)); err != nil {
			return err
		}
	}
	return nil
}

func handleSide(srv *server) error {
	data, err := readFile(srv, "sidebar", 0)
	if err != nil {
		return err
	}
	var buffer strings.Builder
	for m := range data {
		buffer.Write(clean(m))
	}
	reader := bufio.NewReader(strings.NewReader(buffer.String()))
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		fmt.Printf("%s ", line)
	}
	fmt.Println()
	return nil
}

func handleTabs(srv map[string]*server) {
	for name, s := range srv {
		data, err := readFile(s, "tabs", 0)
		if err != nil {
			log.Print(err)
			continue
		}
		for m := range data {
			fmt.Fprintf(os.Stdout, "%s\n", tabs(m.buff, name))	
		}
	}
}
