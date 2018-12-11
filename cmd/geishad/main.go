package main

import "time"
import "os"
import "github.com/faiface/beep"
import "github.com/faiface/beep/mp3"
import "github.com/faiface/beep/speaker"

type Song string

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func play(song Song, done chan nextControl) (*Stream, error) {
	f, err := os.Open(string(song))
	if err != nil {
		return nil, err
	}
	stream, format, err := mp3.Decode(f)
	if err != nil {
		return nil, err
	}
	ss := newStream(stream, format, func(i nextControl) {
		_ = f.Close()
		_ = stream.Close()
		done <- i
	})
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/5))
	speaker.Play(beep.Seq(ss.BeepStream(), beep.Callback(func() {
		ss.Teardown(nextNatural)
	})))
	return ss, nil
}

func main() {
	p := newPlayer()
	go p.listen()
	server(p)
}
