package geisha

type Control int

const (
	PLAY Control = iota
	PAUSE
	TOGGLE
	CLEAR
	FWD
	BWD
	SKIP
	STOP
)
