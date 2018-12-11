package main

import "time"
import "encoding/json"
import "os"
import "fmt"
import "net"
import "github.com/eugene-eeo/geisha"

type subscriber func(geisha.Event) error

func handleConnection(p *player, conn net.Conn, subs chan subscriber) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(time.Second * 2))
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)
	for {
		req := &geisha.Request{}
		err := decoder.Decode(req)
		if err != nil {
			return
		}
		if req.Method == geisha.MethodSubscribe {
			conn.SetDeadline(time.Time{})
			done := make(chan struct{})
			subs <- func(e geisha.Event) error {
				x := append([]byte(e), '\n')
				_, err := conn.Write(x)
				if err != nil {
					done <- struct{}{}
				}
				return err
			}
			<-done
			return
		} else {
			p.context.requests <- req
			if encoder.Encode(<-p.context.response) != nil {
				return
			}
		}
	}
}

func server(p *player) {
	subscribers := make(chan subscriber)
	go func() {
		subs := []subscriber{}
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
		if conn, err := ln.Accept(); err == nil {
			go handleConnection(p, conn, subscribers)
		}
	}
}
