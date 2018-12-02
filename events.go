package geisha

type Event string

const (
	// Controls
	EventCtrlPause  = Event("ctrl:pause")
	EventCtrlPlay   = Event("ctrl:play")
	EventCtrlFwd    = Event("ctrl:fwd")
	EventCtrlBwd    = Event("ctrl:bwd")
	EventCtrlSkip   = Event("ctrl:skip")
	EventCtrlStop   = Event("ctrl:stop")
	EventCtrlClear  = Event("ctrl:clear")
	EventCtrlToggle = Event("ctrl:toggle")
	// Playback Mode
	EventModeChange = Event("mode:change")
	// Song
	EventSongPlay = Event("song:play")
	EventSongDone = Event("song:done")
	// Queue Change
	EventQueueChange = Event("queue:change")
)
