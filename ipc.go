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
	STATUS_OK        Status = 0
	STATUS_ERR       Status = 1
	STATUS_EVENT     Status = 2
	METHOD_GET_STATE Method = iota
	METHOD_GET_QUEUE
	METHOD_SET_QUEUE
	METHOD_SUBSCRIBE
	//METHOD_ENQUEUE
	//METHOD_DEQUEUE
	METHOD_CTRL
)
