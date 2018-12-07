package main

import "github.com/faiface/beep"
import "github.com/faiface/beep/speaker"

type Stream struct {
	stream  beep.StreamSeeker
	stopped bool
	ctrl    *beep.Ctrl
	stop    func(nextControl)
}

func newStream(s beep.StreamSeeker, stop func(nextControl)) *Stream {
	ctrl := &beep.Ctrl{Streamer: s}
	return &Stream{
		stream:  s,
		stopped: false,
		ctrl:    ctrl,
		stop:    stop,
	}
}

func (s *Stream) Toggle() {
	speaker.Lock()
	s.ctrl.Paused = !s.ctrl.Paused
	speaker.Unlock()
}

func (s *Stream) Pause() {
	speaker.Lock()
	s.ctrl.Paused = true
	speaker.Unlock()
}

func (s *Stream) Play() {
	speaker.Lock()
	s.ctrl.Paused = false
	speaker.Unlock()
}

func (s *Stream) Seek(fwd bool) {
	speaker.Lock()
	total := s.stream.Len()
	pcg := float64(s.stream.Position()) / float64(total)
	dir := +0.02
	if !fwd {
		dir = -0.02
	}
	pos := int((pcg + dir) * float64(total))
	s.stream.Seek(max(0, min(pos, total)))
	speaker.Unlock()
}

func (s *Stream) BeepStream() beep.Streamer {
	return s.ctrl
}

func (s *Stream) Paused() bool {
	return s.ctrl.Paused
}

func (s *Stream) SeekRaw(i int) {
	speaker.Lock()
	s.stream.Seek(i)
	speaker.Unlock()
}

func (s *Stream) Progress() float64 {
	speaker.Lock()
	defer speaker.Unlock()
	return float64(s.stream.Position()) / float64(s.stream.Len())
}

func (s *Stream) Teardown(b nextControl) {
	if !s.stopped {
		s.stopped = true
		s.stop(b)
	}
}
