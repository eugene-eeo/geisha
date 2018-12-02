package geisha

import "bufio"
import "encoding/json"
import "errors"
import "io"
import "net"

var useSubscribeMethod error = errors.New("IPC.Request: use IPC.Subscribe for MethodSubscribe")

type IPC struct {
	conn io.ReadWriteCloser
	r    *json.Decoder
	w    *json.Encoder
}

func NewDefaultIPC() (*IPC, error) {
	conn, err := net.Dial("tcp", "localhost:9912")
	if err != nil {
		return nil, err
	}
	return &IPC{
		conn: conn,
		r:    json.NewDecoder(conn),
		w:    json.NewEncoder(conn),
	}, nil
}

func (i *IPC) Close() error {
	return i.conn.Close()
}

func (i *IPC) Subscribe(f func(Event) error) error {
	defer i.conn.Close()
	err := i.w.Encode(Request{Method: MethodSubscribe})
	if err != nil {
		return err
	}
	r := bufio.NewScanner(i.conn)
	for r.Scan() {
		if err := f(Event(r.Text())); err != nil {
			return err
		}
	}
	return nil
}

func (i *IPC) Request(method Method, args []string) (*Response, error) {
	res := &Response{}
	if method == MethodSubscribe {
		return res, useSubscribeMethod
	}
	req := Request{Method: method, Args: args}
	if err := i.w.Encode(req); err != nil {
		return res, err
	}
	if err := i.r.Decode(res); err != nil {
		return res, err
	}
	return res, nil
}
