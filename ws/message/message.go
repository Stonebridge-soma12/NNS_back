package message

type Message struct {
	Type    MessageType `json:"type"`
	UserID  int         `json:"userId"`
	Content string      `json:"content"`
	X       int         `json:"x"`
	Y       int         `json:"y"`
}
