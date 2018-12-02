package main

import "time"
import "encoding/json"
import "bufio"
import "os"
import "fmt"
import "net"
import "github.com/eugene-eeo/geisha"

func handleConnection(p *player, conn net.Conn, subs chan func(geisha.Event) error) {
	defer conn.Close()
	if conn.SetDeadline(time.Now().Add(time.Second*2)) != nil {
		return
	}
	r := bufio.NewReader(conn)
	e := json.NewEncoder(conn)
	d := json.NewDecoder(r)
	for {
		req := &geisha.Request{}
		err := d.Decode(req)
		if err != nil {
			break
		}
		if req.Method == geisha.MethodSubscribe {
			// Once we are in subscribe mode, we should never leave subscribe mode.
			conn.SetDeadline(time.Time{})
			done := make(chan struct{})
			subs <- func(e geisha.Event) error {
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
	subscribers := make(chan func(geisha.Event) error)
	go func() {
		subs := [](func(geisha.Event) error){}
		for {
			select {
			case sub := <-subscribers:
				subs = append(subs, sub)
			case evt := <-p.context.events:
				n := len(subs)
				for i := n - 1; i >= 0; i-- {
					if subs[i](evt) != nil {
						subs = append(subs[:i], subs[i+1:]...)
					}
				}
			}
		}
	}()

	ln, err := net.Listen("tcp", "localhost:9912")
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
