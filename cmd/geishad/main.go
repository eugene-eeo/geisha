package main

import "fmt"
import "time"
import "github.com/faiface/beep"
import "github.com/faiface/beep/mp3"
import "github.com/faiface/beep/speaker"
import "bufio"
import "os"

type Song string
type Event string
type Control int

const (
	PLAY Control = iota
	PAUSE
	CLEAR
	FWD
	BWD
	SKIP
	STOP
)

type State struct {
	NowPlaying Song
	Progress   float64
	Queue      []Song
	Paused     bool
}

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

func broadcast_event(evt string, events chan Event) {
	go func() {
		events <- Event(evt)
	}()
}

// playerloop is meant to be ran in a separate goroutine.
// It handles playing songs and queuing them.
func playerloop(
	controls chan Control,
	songs chan Song,
	state chan State,
	req chan struct{},
	exit chan struct{},
	events chan Event,
) {
	var stream *Stream = nil
	waiting := true
	song_buff := []Song{}
	curr := Song("")
	done := make(chan bool)
	// called when we are ready to play the next song
	// this was removed from the main loop to prevent excessive CPU usage
	playNext := func() {
		for waiting && len(song_buff) > 0 {
			song := song_buff[0]
			song_buff = song_buff[1:]
			s, err := play(song, done)
			// spin until we can play a song without any errors
			if err == nil {
				broadcast_event("SONG:PLAY", events)
				stream = s
				waiting = false
				curr = song
			}
		}
	}
	for {
		select {
		case <-exit:
			break
		case c := <-controls:
			// these controls only act on the stream
			if stream != nil {
				switch c {
				case STOP:
					stream.SeekRaw(0)
					stream.Pause()
					broadcast_event("CTRL:STOP", events)
				case PLAY:
					stream.Play()
					broadcast_event("CTRL:PLAY", events)
				case PAUSE:
					stream.Pause()
					broadcast_event("CTRL:PAUSE", events)
				case FWD:
					stream.Seek(true)
					broadcast_event("CTRL:FWD", events)
				case BWD:
					stream.Seek(false)
					broadcast_event("CTRL:BWD", events)
				case SKIP:
					// this needs to be ran in a goroutine because when we teardown
					// we send an event to the done channel
					go stream.Teardown()
					broadcast_event("CTRL:SKIP", events)
				}
			}
			// queue-oriented controls
			switch c {
			case CLEAR:
				broadcast_event("CTRL:CLEAR", events)
				song_buff = []Song{}
			}
		case <-req:
			f := 0.0
			p := false
			m := make([]Song, min(10, len(song_buff)))
			copy(m, song_buff)
			if stream != nil {
				p = stream.Paused()
				f = stream.Progress()
			}
			state <- State{
				NowPlaying: curr,
				Progress:   f,
				Queue:      m,
				Paused:     p,
			}
		case song := <-songs:
			song_buff = append(song_buff, song)
			broadcast_event("SONG:QUEUED", events)
			playNext()
		case <-done:
			broadcast_event("SONG:DONE", events)
			curr = Song("")
			stream = nil
			waiting = true
			playNext()
		}
	}
}

func main() {
	controls := make(chan Control)
	songs := make(chan Song)
	state := make(chan State)
	req := make(chan struct{})
	exit := make(chan struct{})
	events := make(chan Event)
	go playerloop(controls, songs, state, req, exit, events)

	r := bufio.NewReader(os.Stdin)

	songs <- Song("mp3/beep-01.mp3")
	songs <- Song("mp3/beep-02.mp3")
	songs <- Song("mp3/beep-03.mp3")
	songs <- Song("mp3/file.mp3")
	songs <- Song("mp3/file.mp3")
	songs <- Song("mp3/file.mp3")
	for {
		select {
		case evt := <-events:
			fmt.Println(evt)
		default:
			fmt.Print("g^ ")
			text, _ := r.ReadString('\n')
			switch text {
			case "state\n":
				req <- struct{}{}
				fmt.Println(<-state)
			case "pause\n":
				controls <- PAUSE
			case "play\n":
				controls <- PLAY
			case "fwd\n":
				controls <- FWD
			case "bwd\n":
				controls <- BWD
			case "skip\n":
				controls <- SKIP
			case "stop\n":
				controls <- STOP
			case "clear\n":
				controls <- CLEAR
			}
		}
	}
}
