package main

import "time"
import "strconv"
import "github.com/eugene-eeo/geisha"

var CTRL_EVENT_MAP = map[geisha.Control]geisha.Event{
	geisha.PAUSE:  geisha.EventCtrlPause,
	geisha.PLAY:   geisha.EventCtrlPlay,
	geisha.FWD:    geisha.EventCtrlFwd,
	geisha.BWD:    geisha.EventCtrlBwd,
	geisha.PREV:   geisha.EventCtrlPrev,
	geisha.SKIP:   geisha.EventCtrlSkip,
	geisha.STOP:   geisha.EventCtrlStop,
	geisha.TOGGLE: geisha.EventCtrlToggle,
}

type nextControl int

const (
	nextNatural nextControl = iota
	nextSkip
	nextPrev
	nextNoop
)

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
	done    chan nextControl
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
		queue:  newQueue(false, false),
		done:   make(chan nextControl),
	}
}

func (p *player) broadcast(ev geisha.Event) {
	go func() {
		p.context.events <- ev
	}()
}

func (p *player) play() {
	for p.stream == nil {
		entry := p.queue.current()
		if entry == nil {
			break
		}
		// spin until we can play a song without any errors
		s, err := play(entry.Song, p.done)
		if err != nil {
			p.queue.remove(p.queue.curr)
			continue
		}
		p.broadcast(geisha.EventSongPlay)
		p.stream = s
		break
	}
}

func (p *player) handleDone(i nextControl) {
	p.broadcast(geisha.EventSongDone)
	p.stream = nil
	switch i {
	case nextPrev:
		p.queue.next(-1, true)
	case nextSkip:
		p.queue.next(1, true)
	case nextNatural:
		p.queue.next(1, false)
	case nextNoop:
	}
	p.play()
}

func (p *player) broadcastControlEvent(c geisha.Control) {
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
		case geisha.PREV:
			// this needs to be ran in a goroutine because when we teardown
			// we send an event to p.done
			go p.stream.Teardown(nextPrev)
		case geisha.SKIP:
			go p.stream.Teardown(nextSkip)
		}
	} else {
		switch c {
		case geisha.PREV:
			p.queue.next(-1, true)
			p.play()
		case geisha.SKIP:
			p.queue.next(1, true)
			p.play()
		}
	}
	p.broadcastControlEvent(c)
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

	case geisha.MethodClear:
		p.queue = newQueue(p.queue.loop, p.queue.repeat)
		p.broadcast(geisha.EventQueueChange)

	case geisha.MethodGetState:
		paused := false
		current := -1
		path := Song("")
		elapsed := time.Duration(0 * time.Second)
		total := time.Duration(0 * time.Second)
		if p.stream != nil {
			paused = p.stream.Paused()
			path = p.queue.current().Song
			elapsed, total = p.stream.Progress()
			current = p.queue.current().Id
		}
		res.Result = map[string]interface{}{
			"paused":  paused,
			"elapsed": int(elapsed.Seconds()),
			"total":   int(total.Seconds()),
			"current": current,
			"path":    path,
			"loop":    p.queue.loop,
			"repeat":  p.queue.repeat,
		}

	case geisha.MethodGetQueue:
		curr := -1
		if p.stream != nil {
			curr = p.queue.curr
		}
		queue := make([]*queueEntry, p.queue.len())
		copy(queue, p.queue.q)
		res.Result = map[string]interface{}{
			"queue": queue,
			"curr":  curr,
		}

	case geisha.MethodPlaySong:
		res.Status = geisha.StatusErr
		if len(r.Args) == 1 {
			id, err := strconv.Atoi(r.Args[0])
			if err == nil {
				p.broadcast(geisha.EventQueueChange)
				res.Status = geisha.StatusOk
				if idx := p.queue.find(id); idx >= 0 {
					p.queue.curr = idx
					if p.stream != nil {
						go p.stream.Teardown(nextNoop)
					}
				}
			}
		}

	case geisha.MethodNext:
		p.broadcast(geisha.EventQueueChange)
		for i, song := range r.Args {
			p.queue.insert(p.queue.curr+1+i, Song(song))
		}
		p.play()

	case geisha.MethodEnqueue:
		p.broadcast(geisha.EventQueueChange)
		for _, song := range r.Args {
			p.queue.append(Song(song))
		}
		p.play()

	case geisha.MethodRepeat:
		p.broadcast(geisha.EventModeChange)
		p.queue.repeat = !p.queue.repeat
		p.play()

	case geisha.MethodLoop:
		p.broadcast(geisha.EventModeChange)
		p.queue.loop = !p.queue.loop
		p.play()

	case geisha.MethodSort:
		p.broadcast(geisha.EventQueueChange)
		p.queue.sort()

	case geisha.MethodShuffle:
		p.broadcast(geisha.EventQueueChange)
		p.queue.shuffle()

	case geisha.MethodRemove:
		should_skip := false
		p.broadcast(geisha.EventQueueChange)
		for _, sid := range r.Args {
			id, err := strconv.Atoi(sid)
			if err != nil {
				res.Status = geisha.StatusErr
				break
			}
			if idx := p.queue.find(id); idx >= 0 {
				should_skip = should_skip || idx == p.queue.curr
				p.queue.remove(idx)
			}
		}
		// special case: if we have a currently playing stream
		// and we removed it then p.queue.curr points to the
		// correct song we should play.
		if should_skip && p.stream != nil {
			go p.stream.Teardown(nextNoop)
		}

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
		case i := <-p.done:
			p.handleDone(i)
		case r := <-p.context.requests:
			p.context.response <- p.handleRequest(r)
		}
	}
}
