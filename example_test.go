package ubqt

import (
	"testing"

	"github.com/ubqt-systems/srv-lib"
)

type Event struct {
	filename string
	client   string
}

func (e *Event) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return nil
}

func (e *Event) ReadFile(filename string) ([]byte, error) {
	return nil, nil
}

func (e *Event) CloseFile(filename string) error {
	return nil
}

func main() {
	t := ubqt.newSrv()
	t.Loop(&Event{})

}
