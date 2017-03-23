package ubqtlib

type event struct {
	data chan []byte
	done chan struct{}
}

func (e *event) Close() error {
	close(e.done)
	return nil
}

func (e *event) Read(b []byte) (n int, err error) {
	e.data = make(chan []byte)
	e.done = make(chan struct{})
	for {
		select {
		case buf := <-e.data:
			n += copy(b, buf)
		case <-e.done:
			goto end
		}
		close(e.data)
	}
end:
	return n, nil
}

func (u *Srv) appendEvent(name string) {
	u.Lock()
	c := make(chan []byte)
	defer u.Unlock()
	u.event[name] = c
}

func (u *Srv) readEvent(name string) *event {
	u.appendEvent(name)
	return &event{data: u.event[name]}
}
