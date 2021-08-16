package ws

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	id       int
	hub      *Hub
	socket   *websocket.Conn
	outbound chan []byte
}

func newClient(hub *Hub, socket *websocket.Conn) *Client {
	return &Client{
		hub:      hub,
		socket:   socket,
		outbound: make(chan []byte),
	}
}

func (c *Client) read() {
	defer func() {
		c.hub.unregister <- c
	}()

	for {
		_, data, err := c.socket.ReadMessage()
		if err != nil {
			break
		}

		c.hub.onMessage(data, c)
	}
}

func (c *Client) write() {
	for {
		select {
		case data, ok := <-c.outbound:
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.socket.WriteMessage(websocket.TextMessage, data)
		}
	}
}

func (c Client) close() {
	c.socket.Close()
	close(c.outbound)
}
