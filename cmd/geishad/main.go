package main

import "fmt"
import "time"
import "github.com/faiface/beep"
import "github.com/faiface/beep/mp3"
import "github.com/faiface/beep/speaker"
import "bufio"
import "os"

type Song string
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

type Event struct {
	Type string
}

type State struct {
	NowPlaying Song
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
				go func() { events <- Event{"SONG_PLAYED"} }()
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
				case PLAY:
					stream.Play()
				case PAUSE:
					stream.Pause()
				case FWD:
					stream.Seek(true)
				case BWD:
					stream.Seek(false)
				case SKIP:
					// this needs to be ran in a goroutine because when we teardown
					// we send an event to the done channel
					go stream.Teardown()
				}
			}
			// queue-oriented controls
			switch c {
			case CLEAR:
				song_buff = []Song{}
			}
		case <-req:
			p := false
			m := make([]Song, min(10, len(song_buff)))
			copy(m, song_buff)
			if stream != nil {
				p = stream.Paused()
			}
			state <- State{
				NowPlaying: curr,
				Queue:      m,
				Paused:     p,
			}
		case song := <-songs:
			song_buff = append(song_buff, song)
			go func() { events <- Event{"SONG_QUEUED"} }()
			playNext()
		case <-done:
			go func() { events <- Event{"DONE_PLAYING"} }()
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
			if text == "state\n" {
				req <- struct{}{}
				fmt.Println(<-state)
			}
			if text == "pause\n" {
				controls <- PAUSE
			}
			if text == "play\n" {
				controls <- PLAY
			}
			if text == "fwd\n" {
				controls <- FWD
			}
			if text == "bwd\n" {
				controls <- BWD
			}
			if text == "skip\n" {
				controls <- SKIP
			}
			if text == "stop\n" {
				controls <- STOP
			}
			if text == "clear\n" {
				controls <- CLEAR
			}
		}
	}
}
