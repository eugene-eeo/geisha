package main

import "strconv"
import "fmt"
import "github.com/eugene-eeo/geisha"

var CTRL_EVENT_MAP = map[geisha.Control]string{
	geisha.PAUSE: "CTRL:PAUSE",
	geisha.PLAY:  "CTRL:PLAY",
	geisha.FWD:   "CTRL:FWD",
	geisha.BWD:   "CTRL:BWD",
	geisha.SKIP:  "CTRL:SKIP",
	geisha.STOP:  "CTRL:STOP",
	geisha.CLEAR: "CTRL:CLEAR",
}

type playerContext struct {
	songs    chan Song
	response chan *geisha.Response
	requests chan *geisha.Request
	exit     chan struct{}
	events   chan Event
}

type player struct {
	context playerContext
	waiting bool
	stream  *Stream
	queue   []Song
	curr    Song
	done    chan bool
}

func newPlayer() *player {
	return &player{
		context: playerContext{
			songs:    make(chan Song),
			response: make(chan *geisha.Response),
			requests: make(chan *geisha.Request),
			exit:     make(chan struct{}),
			events:   make(chan Event),
		},
		waiting: true,
		stream:  nil,
		queue:   []Song{},
		curr:    Song(""),
		done:    make(chan bool),
	}
}

func (p *player) broadcast(ev string) {
	go func() {
		p.context.events <- Event(ev)
	}()
}

func (p *player) playNext() {
	for p.waiting && len(p.queue) > 0 {
		song := p.queue[0]
		p.queue = p.queue[1:]
		s, err := play(song, p.done)
		// spin until we can play a song without any errors
		if err == nil {
			p.broadcast("SONG:PLAY")
			p.stream = s
			p.waiting = false
			p.curr = song
		}
	}
}

func (p *player) handleNewSong(song Song) {
	p.queue = append(p.queue, song)
	p.broadcast("SONG:QUEUED")
	p.playNext()
}

func (p *player) handleDone() {
	p.broadcast("SONG:DONE")
	p.stream = nil
	p.waiting = true
	p.curr = Song("")
	p.playNext()
}

func (p *player) broadcastControl(c geisha.Control) {
	ev, ok := CTRL_EVENT_MAP[c]
	if ok {
		p.broadcast(ev)
	}
}

func (p *player) handleControl(c geisha.Control) {
	if p.stream != nil {
		switch c {
		case geisha.STOP:
			p.stream.SeekRaw(0)
			p.stream.Pause()
		case geisha.PLAY:
			p.stream.Play()
		case geisha.PAUSE:
			p.stream.Pause()
		case geisha.FWD:
			p.stream.Seek(true)
		case geisha.BWD:
			p.stream.Seek(false)
		case geisha.SKIP:
			// this needs to be ran in a goroutine because when we teardown
			// we send an event to p.done
			go p.stream.Teardown()
		}
	}
	switch c {
	case geisha.CLEAR:
		p.queue = []Song{}
	}
	p.broadcastControl(c)
}

func (p *player) handleRequest(r *geisha.Request) *geisha.Response {
	res := &geisha.Response{}
	res.Status = geisha.STATUS_OK
	// METHOD_SUBSCRIBE should be handled somewhere else
	switch r.Method {
	case geisha.METHOD_CTRL:
		res.Status = geisha.STATUS_ERR
		if len(r.Args) == 1 {
			i, err := strconv.Atoi(r.Args[0])
			if err == nil {
				res.Status = geisha.STATUS_OK
				p.handleControl(geisha.Control(i))
			}
		}

	case geisha.METHOD_GET_STATE:
		queue := make([]Song, min(10, len(p.queue)))
		copy(queue, p.queue)
		paused := false
		progress := 0.0
		if p.stream != nil {
			paused = p.stream.Paused()
			progress = p.stream.Progress()
		}
		res.Result = map[string]interface{}{
			"queue":    queue,
			"paused":   paused,
			"progress": progress,
			"current":  p.curr,
		}

	case geisha.METHOD_GET_QUEUE:
		fmt.Println("get_queue")
		queue := make([]Song, len(p.queue))
		copy(queue, p.queue)
		res.Result = map[string]interface{}{"queue": queue}

	case geisha.METHOD_SET_QUEUE:
		fmt.Println("set_queue")
		p.queue = make([]Song, len(r.Args))
		for i, song := range r.Args {
			p.queue[i] = Song(song)
		}
		p.playNext()
	}
	return res
}

func (p *player) loop() {
	for {
		select {
		case <-p.context.exit:
			break
		case song := <-p.context.songs:
			p.handleNewSong(song)
		case <-p.done:
			p.handleDone()
		case r := <-p.context.requests:
			p.context.response <- p.handleRequest(r)
		}
	}
}
