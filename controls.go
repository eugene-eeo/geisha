package geisha

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
