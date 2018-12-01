package geisha

//import "encoding/json"

type Status int
type Method int
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
	MethodGetState Method = iota
	MethodGetQueue
	MethodSubscribe
	MethodPlaySong
	MethodEnqueue
	MethodNext
	MethodCtrl
	MethodSort
	MethodShuffle
	MethodShutdown
)
