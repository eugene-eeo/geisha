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
	return r.Err()
}

func (i *IPC) Request(method Method, args []string) (*Response, error) {
	var msg json.RawMessage
	res := &Response{Result: &msg}
	if method == MethodSubscribe {
		return nil, useSubscribeMethod
	}
	req := Request{Method: method, Args: args}
	if err := i.w.Encode(req); err != nil {
		return nil, err
	}
	if err := i.r.Decode(res); err != nil {
		return nil, err
	}
	err := convertResponseType(method, msg, res)
	return res, err
}

func convertResponseType(m Method, msg json.RawMessage, res *Response) error {
	switch m {
	case MethodGetState:
		r := GetStateResponse{}
		err := json.Unmarshal(msg, &r)
		res.Result = r
		return err
	case MethodGetQueue:
		r := GetQueueResponse{}
		err := json.Unmarshal(msg, &r)
		res.Result = r
		return err
	default:
		res.Result = nil
		return nil
	}
}

type GetStateResponse struct {
	Elapsed  int    `json:"elapsed"`
	Total    int    `json:"total"`
	Current  int    `json:"current"`
	Path     string `json:"path"`
	Paused   bool   `json:"paused"`
	Loop     bool   `json:"loop"`
	Repeat   bool   `json:"repeat"`
	Shuffled bool   `json:"shuffled"`
}

type QueueEntry struct {
	Id   int    `json:"id"`
	Song string `json:"song"`
}

type GetQueueResponse struct {
	Current int          `json:"current"`
	Queue   []QueueEntry `json:"queue"`
}
