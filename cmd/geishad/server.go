package main

import "encoding/json"
import "bufio"
import "os"
import "fmt"
import "net"
import "github.com/eugene-eeo/geisha"

func handleConnection(p *player, conn net.Conn, subs chan func(Event) error) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	e := json.NewEncoder(conn)
	d := json.NewDecoder(r)
	for {
		req := &geisha.Request{}
		err := d.Decode(req)
		if err != nil {
			break
		}
		if req.Method == geisha.METHOD_SUBSCRIBE {
			done := make(chan struct{})
			subs <- func(e Event) error {
				x := []byte(e)
				x = append(x, '\n')
				_, err := conn.Write(x)
				if err != nil {
					done <- struct{}{}
				}
				return err
			}
			<-done
			break
		} else {
			p.context.requests <- req
			res := <-p.context.response
			if e.Encode(res) != nil {
				break
			}
		}
	}
}

func server(p *player) {
	fmt.Println(geisha.METHOD_GET_STATE)
	fmt.Println(geisha.METHOD_GET_QUEUE)
	fmt.Println(geisha.METHOD_SET_QUEUE)
	fmt.Println(geisha.METHOD_SUBSCRIBE)
	fmt.Println(geisha.METHOD_CTRL)

	subscribers := make(chan func(Event) error)
	go func() {
		subs := [](func(Event) error){}
		for {
			select {
			case sub := <-subscribers:
				subs = append(subs, sub)
			case evt := <-p.context.events:
				fmt.Println(evt)
				n := len(subs)
				for i := n - 1; i >= 0; i-- {
					if subs[i](evt) != nil {
						subs = append(subs[:i], subs[i+1:]...)
					}
				}
			}
		}
	}()

	ln, err := net.Listen("tcp", ":9912")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for {
		conn, err := ln.Accept()
		if err == nil {
			go handleConnection(p, conn, subscribers)
		}
	}
}
