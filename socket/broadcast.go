package socket

// TODO namespace and room should implement this
type Broadcastable interface {
	BroadCast(string, interface{})
}