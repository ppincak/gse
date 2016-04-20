package socket

import "encoding/json"

var Errors = map[int]string{

}

const (

)

type Error struct {
	ErrorCode 	int    	`json:"errorCode"`
	Message   	string 	`json:"message"`
	Cause     	string 	`json:"cause,omitempty"`
}

func (e Error) Error() string {
	return e.Message + " (" + e.Cause + ")"
}

func (e Error) toJsonString() string {
	b, _ := json.Marshal(e)
	return string(b)
}