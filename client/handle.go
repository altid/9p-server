package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
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
				//TODO: Scrub out color, url, image
				if _, err := os.Stdout.Write(m.buff); err != nil {
					return
				}
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
		if _, err := os.Stdout.Write(m.buff); err != nil {
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
		if _, err := os.Stdout.Write(m.buff); err != nil {
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
	var buffer []byte
	for m := range data {
		buffer = append(buffer, m.buff...)
	}
	reader := bufio.NewReader(bytes.NewReader(buffer))
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
	var active string
	r := regexp.MustCompile(`%\[([^\s]+)\]\(([^\s,]+)\)`)
	for name, s := range srv {
		data, err := readFile(s, "tabs", 0)
		if err != nil {
			log.Print(err)
			continue
		}
		var buffers []byte
		for m := range data {
			buffers = append(buffers, m.buff...)
		}
		reader := bufio.NewReader(bytes.NewReader(buffers))
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			matches := r.FindAllStringSubmatch(line, -1)
			if matches[0][2] == "blue" {
				fmt.Printf("+")
			}
			if matches[0][2] == "red" {
				fmt.Printf("!")
			}
			if name == current {
				if matches[0][2] == "purple" {
					active = matches[0][1]
				}
				fmt.Printf("%s ", matches[0][1])
				continue
			}
			fmt.Printf("%s/%s ", name, matches[0][1])
		}
	}
	fmt.Printf("\nCurrent: %s\n", active)
}
