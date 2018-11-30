package main

import "time"
import "os"
import "github.com/faiface/beep"
import "github.com/faiface/beep/mp3"
import "github.com/faiface/beep/speaker"

type Song string
type Event string

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func play(song Song, done chan bool) (*Stream, error) {
	f, err := os.Open(string(song))
	if err != nil {
		return nil, err
	}
	stream, format, err := mp3.Decode(f)
	if err != nil {
		return nil, err
	}
	ss := newStream(stream, func() {
		_ = f.Close()
		_ = stream.Close()
		done <- true
	})
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(beep.Seq(ss.BeepStream(), beep.Callback(ss.Teardown)))
	return ss, nil
}

func main() {
	p := newPlayer()
	go p.loop()
	server(p)
}
