package message

type MessageType int

const (
	MessageTypeConnected MessageType = iota + 1
	MessageTypeUserJoin
	MessageTypeUserLeft
	MessageTypeMessage
)
