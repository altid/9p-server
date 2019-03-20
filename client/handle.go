package main

import (
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
			buffer = "buffer " + parts[1]
		}
	}
	s := srv[current]
	data := &content{
		buff: []byte(command),
	}
	return writeFile(s, "ctrl", data, 0) 
}

func handleInput(s *server, input string) error {
	data := &content{
		buff: []byte(input),
		err: nil,
	}
	return writeFile(s, "input", data, 0)
	
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

// TODO halfwit: Strip the color tokens off, and match the purple buffer of current 
// So we can output `current buffer: <buffer>`
// https://github.com/ubqt-systems/9p-server/issues/10
func handleTabs(srv map[string]*server) {
	var active string
	for name, s := range srv {
		data, err := readFile(s, "tabs")
		if err != nil {
			log.Print(err)
			continue
		}
		for m := range data {
			matches := parseTabFileChunk(string(m.buff))
			if len(matches) < 1 {
				continue
			}
			for _, match := range matches {
				if len(match) != 3 {
					continue
				}
				if name == current {
					if match[2] == "purple" {
						active = match[1]
						continue
					}
					fmt.Printf("%s ", match[1])
					continue
				}
				fmt.Printf("%s/%s ", name, match[1])
			}
		}
	}
	fmt.Printf("\nCurrent: %s\n", active)
}

func parseTabFileChunk(tab string) [][]string {
	r := regexp.MustCompile(`%\[([^\s]+)\]\(([^\s,]+)\)`)
	matches := r.FindAllStringSubmatch(tab, -1)
	return matches
}
