package main

import "strconv"
import "github.com/eugene-eeo/geisha"

var CTRL_EVENT_MAP = map[geisha.Control]geisha.Event{
	geisha.PAUSE:  geisha.EventCtrlPause,
	geisha.PLAY:   geisha.EventCtrlPlay,
	geisha.FWD:    geisha.EventCtrlFwd,
	geisha.BWD:    geisha.EventCtrlBwd,
	geisha.SKIP:   geisha.EventCtrlSkip,
	geisha.STOP:   geisha.EventCtrlStop,
	geisha.CLEAR:  geisha.EventCtrlClear,
	geisha.TOGGLE: geisha.EventCtrlToggle,
}

type playerContext struct {
	response chan *geisha.Response
	requests chan *geisha.Request
	exit     chan struct{}
	events   chan geisha.Event
}

type player struct {
	context playerContext
	stream  *Stream
	queue   *queue
	done    chan bool
	loop    bool
	repeat  bool
}

func newPlayer() *player {
	return &player{
		context: playerContext{
			response: make(chan *geisha.Response),
			requests: make(chan *geisha.Request),
			exit:     make(chan struct{}),
			events:   make(chan geisha.Event),
		},
		stream: nil,
		queue:  newQueue(),
		done:   make(chan bool),
	}
}

func (p *player) broadcast(ev geisha.Event) {
	go func() {
		p.context.events <- ev
	}()
}

func (p *player) playNext() {
	for p.stream == nil {
		song := p.queue.next(p.repeat, p.loop)
		if song == Song("") {
			break
		}
		// spin until we can play a song without any errors
		s, err := play(song, p.done)
		if err != nil {
			p.queue.remove()
			continue
		}
		p.broadcast(geisha.EventSongPlay)
		p.stream = s
		break
	}
}

func (p *player) handleDone() {
	p.broadcast(geisha.EventSongDone)
	p.stream = nil
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
		case geisha.TOGGLE:
			p.stream.Toggle()
		case geisha.SKIP:
			// this needs to be ran in a goroutine because when we teardown
			// we send an event to p.done
			go p.stream.Teardown()
		}
	}
	switch c {
	case geisha.CLEAR:
		p.queue = newQueue()
	}
	p.broadcastControl(c)
}

func (p *player) handleRequest(r *geisha.Request) *geisha.Response {
	res := &geisha.Response{}
	res.Status = geisha.StatusOk
	// MethodSubscribe should be handled somewhere else
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
		paused := false
		progress := 0.0
		if p.stream != nil {
			paused = p.stream.Paused()
			progress = p.stream.Progress()
		}
		res.Result = map[string]interface{}{
			"paused":   paused,
			"progress": progress,
			"current":  p.queue.peek(),
			"loop":     p.loop,
			"repeat":   p.repeat,
		}

	case geisha.MethodGetQueue:
		queue := make([]Song, p.queue.len())
		copy(queue, p.queue.q)
		res.Result = map[string]interface{}{
			"queue": queue,
			"curr":  p.queue.curr,
		}

	case geisha.MethodPlaySong:
		res.Status = geisha.StatusErr
		if len(r.Args) == 1 {
			p.broadcast(geisha.EventQueueChange)
			res.Status = geisha.StatusOk
			p.queue.append(Song(r.Args[0]))
			if p.stream != nil {
				go p.stream.Teardown()
			}
			p.playNext()
		}

	case geisha.MethodNext:
		res.Status = geisha.StatusErr
		if len(r.Args) == 1 {
			p.broadcast(geisha.EventQueueChange)
			res.Status = geisha.StatusOk
			p.queue.prepend(Song(r.Args[0]))
			p.playNext()
		}

	case geisha.MethodEnqueue:
		res.Status = geisha.StatusErr
		if len(r.Args) == 1 {
			p.broadcast(geisha.EventQueueChange)
			res.Status = geisha.StatusOk
			p.queue.append(Song(r.Args[0]))
			p.playNext()
		}

	case geisha.MethodRepeat:
		p.broadcast(geisha.EventModeChange)
		p.repeat = !p.repeat
		p.playNext()

	case geisha.MethodLoop:
		p.broadcast(geisha.EventModeChange)
		p.loop = !p.loop
		p.playNext()

	case geisha.MethodSort:
		p.broadcast(geisha.EventQueueChange)
		p.queue.sort()

	case geisha.MethodShuffle:
		p.broadcast(geisha.EventQueueChange)
		p.queue.shuffle()

	case geisha.MethodShutdown:
		go func() { p.context.exit <- struct{}{} }()
	}
	return res
}

func (p *player) listen() {
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
