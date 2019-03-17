package main

import (
	"fmt"
	"log"
	"os"
)

func handleCtrl(srv map[string]*server, command string) error {
	// TODO: Switch on command and send the right commands
	// Make sure we close(done) on services is we remove them
	s := srv[current]
	data := make(chan *content)
	defer close(data)
	err := writeFile(s, "document", data, 0)
	data <- &content{
		buff: []byte(command),
	}
	return err
}

func handleInput(s *server, input string) error {
	data := make(chan *content)
	defer close(data)
	err := writeFile(s, "input", data, 0)
	data <- &content{
		buff: []byte(input),
		err: nil,
	}
	return err
	
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
	defer os.Stdout.Write([]byte("\n"))
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
	// var active string
	for name, s := range srv {
		data, err := readFile(s, "tabs")
		if err != nil {
			log.Print(err)
			continue
		}
		for m := range data {
			if name == current {
				// tabName := uncolor(?)
				// if strings.HasSuffix(m.buff, "(purple)") {
				//	active = tabName
				//}
				fmt.Printf("%s ", m.buff)
				continue
			}
			fmt.Printf("%s/%s ", name, m.buff)
		}
	}
	//fmt.Printf("Current buffer: %s\n", active
}
