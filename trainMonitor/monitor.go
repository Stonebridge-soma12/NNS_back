package trainMonitor

import (
	"bytes"
	"encoding/gob"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
	channel chan []byte
}

func (c *Client) Close() {
	c.conn.Close()
	close(c.channel)
}

func (c *Client) Send(monitor Monitor) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(monitor)
	if err != nil {
		return err
	}

	c.channel <- buf.Bytes()

	return nil
}

func (c *Client) write() {
	for {
		select {
		case msg, ok := <-c.channel:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

type Monitor struct {
	Epoch    Epoch
	TrainLog TrainLog
}


