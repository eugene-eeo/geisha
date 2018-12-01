package main

import "strconv"
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
	res.Status = geisha.StatusOk
	// METHOD_SUBSCRIBE should be handled somewhere else
	switch r.Method {
	case geisha.MethodCtrl:
		res.Status = geisha.StatusErr
		if len(r.Args) == 1 {
			i, err := strconv.Atoi(r.Args[0])
			if err == nil {
				res.Status = geisha.StatusOk
				p.handleControl(geisha.Control(i))
			}
		}

	case geisha.MethodGetState:
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

	case geisha.MethodGetQueue:
		queue := make([]Song, len(p.queue))
		copy(queue, p.queue)
		res.Result = map[string]interface{}{"queue": queue}

	case geisha.MethodPlaySong:
		res.Status = geisha.StatusErr
		if len(r.Args) == 1 {
			res.Status = geisha.StatusOk
			p.queue = append([]Song{Song(r.Args[0])}, p.queue...)
			if p.stream != nil {
				go p.stream.Teardown()
			}
			p.playNext()
		}

	case geisha.MethodNext:
		res.Status = geisha.StatusErr
		if len(r.Args) == 1 {
			res.Status = geisha.StatusOk
			p.queue = append([]Song{Song(r.Args[0])}, p.queue...)
			p.playNext()
		}

	case geisha.MethodEnqueue:
		res.Status = geisha.StatusErr
		if len(r.Args) == 1 {
			res.Status = geisha.StatusOk
			p.queue = append(p.queue, Song(r.Args[0]))
			p.playNext()
		}

	case geisha.MethodShutdown:
		p.context.exit <- struct{}{}
	}
	return res
}

func (p *player) loop() {
	for {
		select {
		case <-p.context.exit:
			break
		case <-p.done:
			p.handleDone()
		case r := <-p.context.requests:
			p.context.response <- p.handleRequest(r)
		}
	}
}
