package main

import "os"
import "time"
import "strconv"
import "github.com/eugene-eeo/geisha"

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
	p.broadcast(geisha.CtrlToEvent(c))
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
		if p.stream != nil {
			go p.stream.Teardown(nextNoop)
		}

	case geisha.MethodGetState:
		paused := false
		current := &queueEntry{}
		elapsed := time.Duration(0)
		total := time.Duration(0)
		if p.stream != nil {
			paused = p.stream.Paused()
			current = p.queue.current()
			elapsed, total = p.stream.Progress()
		}
		res.Result = geisha.GetStateResponse{
			Elapsed:  int(elapsed.Seconds()),
			Total:    int(total.Seconds()),
			Current:  current.Id,
			Path:     string(current.Song),
			Paused:   paused,
			Loop:     p.queue.loop,
			Repeat:   p.queue.repeat,
			Shuffled: p.queue.shuffled,
		}

	case geisha.MethodGetQueue:
		curr := -1
		if p.stream != nil {
			curr = p.queue.current().Id
		}
		res.Result = map[string]interface{}{
			"queue":   p.queue.q,
			"current": curr,
		}

	case geisha.MethodPlaySong:
		res.Status = geisha.StatusErr
		if len(r.Args) == 1 {
			id, err := strconv.Atoi(r.Args[0])
			if err == nil {
				if idx := p.queue.find(id); idx >= 0 {
					res.Status = geisha.StatusOk
					p.broadcast(geisha.EventQueueChange)
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
			os.Exit(0)
		case i := <-p.done:
			p.handleDone(i)
		case r := <-p.context.requests:
			p.context.response <- p.handleRequest(r)
		}
	}
}
