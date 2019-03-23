package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

// TODO(halfwit): Return any errors on an error chan here, or nil
func handleCtrl(srv map[string]*server, command string) error {
	if strings.HasPrefix(command, "buffer ") {
		buffer := strings.TrimPrefix(command, "buffer ")
		parts := strings.Split(buffer, "/")
		if _, ok := srv[parts[0]]; ok && len(parts) > 1{
			current = parts[0]
			command = "buffer " + parts[1]
			// we gotta restart the events loop here
		}
	}
	s := srv[current]
	data := &content{
		buff: []byte(command + "\n"),
	}
	return writeFile(s, "ctrl", data)
}

func handleInput(s *server, input string) error {
	log.Println(input)
	data := &content{
		buff: []byte(input + "\n"),
		err: nil,
	}
	return writeFile(s, "input", data)
	
}

func handleMessage(s *server, m *msg) error {
	if m.srv != current {
		return nil
	}
	if m.msg != "document" && m.msg != "feed" && m.msg != "stream" {
		return fmt.Errorf("%s not supported currently", m.msg)
	}
	data, err := readFile(s, m.msg)
	if err != nil {
		return err
	}
	for m := range data {
		if m.err != nil {
			return err
		}
		if _, err := os.Stdout.Write(m.buff); err != nil {
			return err
		}
	}
	return nil
}

func handleTabs(srv map[string]*server) {
	var active string
	r := regexp.MustCompile(`%\[([^\s]+)\]\(([^\s,]+)\)`)
	for name, s := range srv {
		data, err := readFile(s, "tabs")
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
