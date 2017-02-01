package ubqtlib_test

import (
	"github.com/ubqt-systems/ubqtlib"
)

type myClient struct{}

func (c *myClient) ClientWrite(filename string, client string, data []byte) (int, error) {
	return len(data), nil
}

func (c *myClient) ClientRead(filename string, client string) ([]byte, error) {
	return []byte("hello"), nil
}

func (c *myClient) ClientClose(filename string, client string) error {
	return nil
}

// ExampleServer - Simply return hello on reads to any file
func ExampleServer() {
	t := ubqtlib.NewSrv()
	t.SetPort(":1234")
	c := &myClient{}
	err := t.Loop(c)
	if err != nil {
		// do something
	}
}
