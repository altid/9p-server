package ubqtlib_test

import (
	"os"

	"github.com/ubqt-systems/ubqtlib"
)

type myClient struct{}

func (c *myClient) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return nil
}

func (c *myClient) ReadFile(filename string) ([]byte, error) {
	return []byte("hello"), nil
}

func (c *myClient) CloseFile(filename string) error {
	return nil
}

// ExampleServer - Simply return hello on reads to any file
func ExampleServer() {
	t := ubqtlib.NewSrv()
	t.port = ":1234"
	c := &myClient{}
	t.Loop(c)
}
