package geisha

type Status int
type Method int
type Event string
type Control int
type Result interface{}

type Request struct {
	Method Method   `json:"method"`
	Args   []string `json:"args"`
}

type Response struct {
	Status Status `json:"status"`
	Result Result `json:"result"`
}

const (
	StatusOk  Status = 0
	StatusErr Status = 1
)

const (
	PLAY Control = iota
	PAUSE
	TOGGLE
	FWD
	BWD
	PREV
	SKIP
	STOP
)

const (
	MethodGetState Method = iota
	MethodGetQueue
	MethodSubscribe
	MethodPlaySong
	MethodEnqueue
	MethodNext
	MethodCtrl
	MethodSort
	MethodShuffle
	MethodLoop
	MethodRepeat
	MethodClear
	MethodShutdown
)

const (
	// Controls
	EventCtrlPause  = Event("ctrl:pause")
	EventCtrlPlay   = Event("ctrl:play")
	EventCtrlFwd    = Event("ctrl:fwd")
	EventCtrlBwd    = Event("ctrl:bwd")
	EventCtrlSkip   = Event("ctrl:skip")
	EventCtrlPrev   = Event("ctrl:prev")
	EventCtrlStop   = Event("ctrl:stop")
	EventCtrlToggle = Event("ctrl:toggle")
	// Playback Mode
	EventModeChange = Event("mode:change")
	// Song
	EventSongPlay = Event("song:play")
	EventSongDone = Event("song:done")
	// Queue Change
	EventQueueChange = Event("queue:change")
)
