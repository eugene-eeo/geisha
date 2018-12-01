package main

import "bufio"
import "strconv"
import "sort"
import "os"
import "encoding/json"
import "fmt"
import "net"
import "github.com/eugene-eeo/geisha"
import "github.com/urfave/cli"

func ipc(f func(*cli.Context, *json.Decoder, *json.Encoder) (*geisha.Response, error)) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		conn, err := net.Dial("tcp", "localhost:9912")
		if err != nil {
			return err
		}
		defer conn.Close()
		r := json.NewDecoder(conn)
		w := json.NewEncoder(conn)
		res, err := f(c, r, w)
		if err != nil {
			return err
		}
		if res.Status == geisha.STATUS_OK {
			os.Exit(0)
		}
		os.Exit(1)
		return nil
	}
}

func play(c *cli.Context, r *json.Decoder, w *json.Encoder) (*geisha.Response, error) {
	args := c.Args()
	if len(args) < 1 {
		return nil, fmt.Errorf("geishac: play needs song")
	}
	song := args[0]
	res := &geisha.Response{}
	w.Encode(geisha.Request{
		Method: geisha.METHOD_PLAY_SONG,
		Args:   []string{song},
	})
	err := r.Decode(res)
	return res, err
}

func enqueue(c *cli.Context, r *json.Decoder, w *json.Encoder) (*geisha.Response, error) {
	args := c.Args()
	if len(args) < 1 {
		return nil, fmt.Errorf("geishac: enqueue needs song")
	}
	song := args[0]
	res := &geisha.Response{}
	w.Encode(geisha.Request{
		Method: geisha.METHOD_ENQUEUE,
		Args:   []string{song},
	})
	err := r.Decode(res)
	return res, err
}

func ctrl(ct geisha.Control) func(*cli.Context) error {
	return ipc(func(c *cli.Context, r *json.Decoder, w *json.Encoder) (*geisha.Response, error) {
		w.Encode(geisha.Request{
			Method: geisha.METHOD_CTRL,
			Args:   []string{strconv.Itoa(int(ct))},
		})
		res := &geisha.Response{}
		err := r.Decode(res)
		return res, err
	})
}

func sub(c *cli.Context) error {
	conn, err := net.Dial("tcp", "localhost:9912")
	defer conn.Close()
	if err != nil {
		return err
	}
	s := bufio.NewScanner(bufio.NewReader(conn))
	w := json.NewEncoder(conn)
	w.Encode(geisha.Request{Method: geisha.METHOD_SUBSCRIBE})
	for s.Scan() {
		fmt.Println(s.Text())
	}
	return s.Err()
}

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:   "sub",
			Usage:  "subscribe to events",
			Action: sub,
		},
		{
			Name:   "play",
			Usage:  "play song",
			Action: ipc(play),
		},
		{
			Name:   "enqueue",
			Usage:  "enqueue song",
			Action: ipc(enqueue),
		},
		{
			Name:   "resume",
			Usage:  "resume",
			Action: ctrl(geisha.PLAY),
		},
		{
			Name:   "pause",
			Usage:  "pause",
			Action: ctrl(geisha.PAUSE),
		},
		{
			Name:   "fwd",
			Usage:  "forward",
			Action: ctrl(geisha.FWD),
		},
		{
			Name:   "bwd",
			Usage:  "backwards",
			Action: ctrl(geisha.BWD),
		},
		{
			Name:   "stop",
			Usage:  "stop",
			Action: ctrl(geisha.STOP),
		},
		{
			Name:   "skip",
			Usage:  "skip",
			Action: ctrl(geisha.SKIP),
		},
	}
	sort.Sort(cli.CommandsByName(app.Commands))
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
