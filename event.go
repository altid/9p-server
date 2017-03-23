package ubqtlib

import "io"

type event struct {
	data chan []byte
	done chan struct{}
	name string
	u    *Srv
}

func (e *event) Close() error {
	close(e.done)
	return nil
}

// On read, the file will want to get back stuff into its buffer; which we copy
func (e *event) Read(b []byte) (n int, err error) {
	buf, ok := <-e.data
	if !ok {
		return 0, io.EOF
	}
	n += copy(b, buf)
	return n, err
}

func (u *Srv) readEvent(name string) *event {
	data := make(chan []byte, 10)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case buf := <-u.event:
				data <- buf
			case <-done:
				break
			}
		}
		close(data)
	}()
	return &event{data: data, done: done}
}
