package socket

type CommandType int;

const(
	CreateRoom 	CommandType = iota
	DestroyRoom
	CloseClient
)

type Command struct {
	CmdType 	CommandType
	Data    	interface{}
	OutputC		chan interface{}
}