package socket

import "encoding/json"

var Errors = map[int]string{
	RoomDoesNotExist: "Room doesnt exist",
}

const (
	RoomDoesNotExist = iota
)

type Error struct {
	ErrorCode 	int    	`json:"errorCode"`
	Message   	string 	`json:"message"`
	Cause     	string 	`json:"cause,omitempty"`
}

func makeError(errorCode int) Error {
	return Error{
		ErrorCode: errorCode,
		Message: Errors[errorCode],
	}
}

func (e Error) Error() string {
	return e.Message + " (" + e.Cause + ")"
}

func (e Error) toJsonString() string {
	b, _ := json.Marshal(e)
	return string(b)
}