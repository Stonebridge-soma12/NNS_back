package message

type Message struct {
	TYpe    MessageType `json:"tYpe"`
	UserID  int         `json:"userId"`
	Content string      `json:"content"`
}