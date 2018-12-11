package main

import "strconv"
import "sort"
import "os"
import "fmt"
import "github.com/eugene-eeo/geisha"
import "github.com/urfave/cli"

func ipc(f func(*cli.Context, *geisha.IPC) (*geisha.Response, error)) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		ipc, err := geisha.NewDefaultIPC()
		if err != nil {
			return err
		}
		defer ipc.Close()
		res, err := f(c, ipc)
		if err != nil {
			return err
		}
		if res.Status == geisha.StatusErr {
			os.Exit(1)
		}
		return nil
	}
}

func play(c *cli.Context, ipc *geisha.IPC) (*geisha.Response, error) {
	args := c.Args()
	if len(args) < 1 {
		return nil, fmt.Errorf("geishac: play needs id")
	}
	return ipc.Request(geisha.MethodPlaySong, []string{args[0]})
}

func next(c *cli.Context, ipc *geisha.IPC) (*geisha.Response, error) {
	args := c.Args()
	if len(args) < 1 {
		return nil, fmt.Errorf("geishac: next needs song")
	}
	return ipc.Request(geisha.MethodNext, []string{args[0]})
}

func enqueue(c *cli.Context, ipc *geisha.IPC) (*geisha.Response, error) {
	return ipc.Request(geisha.MethodEnqueue, []string(c.Args()))
}

func remove(c *cli.Context, ipc *geisha.IPC) (*geisha.Response, error) {
	return ipc.Request(geisha.MethodRemove, []string(c.Args()))
}

func get_state(c *cli.Context, ipc *geisha.IPC) (*geisha.Response, error) {
	res, err := ipc.Request(geisha.MethodGetState, nil)
	if err != nil {
		return res, err
	}
	x := res.Result.(map[string]interface{})
	fmt.Println("current:\t", x["current"].(float64))
	fmt.Println("elapsed:\t", x["elapsed"].(float64))
	fmt.Println("total:\t", x["total"].(float64))
	fmt.Println("paused:\t", x["paused"].(bool))
	fmt.Println("loop:\t", x["loop"].(bool))
	fmt.Println("repeat:\t", x["repeat"].(bool))
	return res, nil
}

func get_queue(c *cli.Context, ipc *geisha.IPC) (*geisha.Response, error) {
	res, err := ipc.Request(geisha.MethodGetQueue, nil)
	if err != nil {
		return res, err
	}
	x := res.Result.(map[string]interface{})
	i := int(x["curr"].(float64))
	q := x["queue"].([]interface{})
	for j, song := range q {
		f := ""
		if i == j {
			f = "*"
		}
		entry := song.(map[string]interface{})
		fmt.Printf("%s\t%d\t%s\n", f, int(entry["id"].(float64)), entry["song"].(string))
	}
	return res, err
}

func meth(m geisha.Method) func(*cli.Context) error {
	return ipc(func(c *cli.Context, ipc *geisha.IPC) (*geisha.Response, error) {
		return ipc.Request(m, nil)
	})
}

func ctrl(ct geisha.Control) func(*cli.Context) error {
	return ipc(func(c *cli.Context, ipc *geisha.IPC) (*geisha.Response, error) {
		return ipc.Request(geisha.MethodCtrl, []string{
			strconv.Itoa(int(ct)),
		})
	})
}

func sub(c *cli.Context) error {
	ipc, err := geisha.NewDefaultIPC()
	defer ipc.Close()
	if err != nil {
		return err
	}
	return ipc.Subscribe(func(evt geisha.Event) error {
		fmt.Println(evt)
		return nil
	})
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
			Name:   "get_queue",
			Usage:  "get queue",
			Action: ipc(get_queue),
		},
		{
			Name:   "play",
			Usage:  "play song",
			Action: ipc(play),
		},
		{
			Name:   "next",
			Usage:  "next song",
			Action: ipc(next),
		},
		{
			Name:   "enqueue",
			Usage:  "enqueue song(s)",
			Action: ipc(enqueue),
		},
		{
			Name:   "remove",
			Usage:  "remove song(s)",
			Action: ipc(remove),
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
			Name:   "prev",
			Usage:  "prev",
			Action: ctrl(geisha.PREV),
		},
		{
			Name:   "skip",
			Usage:  "skip",
			Action: ctrl(geisha.SKIP),
		},
		{
			Name:   "toggle",
			Usage:  "toggle",
			Action: ctrl(geisha.TOGGLE),
		},
		{
			Name:   "clear",
			Usage:  "clear",
			Action: meth(geisha.MethodClear),
		},
		{
			Name:   "shuffle",
			Usage:  "shuffle queue",
			Action: meth(geisha.MethodShuffle),
		},
		{
			Name:   "repeat",
			Usage:  "toggle repeat",
			Action: meth(geisha.MethodRepeat),
		},
		{
			Name:   "sort",
			Usage:  "sort queue",
			Action: meth(geisha.MethodSort),
		},
		{
			Name:   "loop",
			Usage:  "toggle loop",
			Action: meth(geisha.MethodLoop),
		},
		{
			Name:   "get_state",
			Usage:  "get server state",
			Action: ipc(get_state),
		},
	}
	sort.Sort(cli.CommandsByName(app.Commands))
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
