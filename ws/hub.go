package ws

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"nns_back/ws/message"
)

type Hub struct {
	clients    []*Client
	nextID     int
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make([]*Client, 0),
		nextID:     0,
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.onConnect(client)
		case client := <-h.unregister:
			h.onDisconnect(client)
		}
	}
}

func (h *Hub) onConnect(client *Client) {
	// make new client
	client.id = h.nextID
	h.nextID++
	h.clients = append(h.clients, client)

	// make list of all users
	users := make([]message.User, 0, len(h.clients))
	for _, c := range h.clients {
		users = append(users, message.User{ID: c.id})
	}

	// notify that a user join
	h.send(message.NewConnected(users), client)
	h.broadcast(message.NewUserJoined(client.id), client)
}

func (h *Hub) onDisconnect(client *Client) {
	client.close()

	// find index of client
	i := -1
	for j, c := range h.clients {
		if c.id == client.id {
			i = j
			break
		}
	}

	// delete client from list
	copy(h.clients[i:], h.clients[i+1:])
	h.clients[len(h.clients)-1] = nil
	h.clients = h.clients[:len(h.clients)-1]

	h.broadcast(message.NewUserLeft(client.id), nil)
}

func (h *Hub) onMessage(data []byte, client *Client) {
	var msg message.Message
	if json.Unmarshal(data, &msg) != nil {
		return
	}

	msg.UserID = client.id
	h.broadcast(msg, client)
}

func (h *Hub) broadcast(message interface{}, ignore *Client) {
	data, _ := json.Marshal(message)
	for _, c := range h.clients {
		if c == ignore {
			continue
		}

		c.outbound <- data
	}
}

func (h *Hub) send(message interface{}, client *Client) {
	data, _ := json.Marshal(message)
	client.outbound <- data
}

var upgrader = websocket.Upgrader{
	//HandshakeTimeout:  0,
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	//WriteBufferPool:   nil,
	//Subprotocols:      nil,
	//Error:             nil,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	//EnableCompression: false,
}

func (h *Hub) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "could not upgrade", http.StatusInternalServerError)
		return
	}

	client := newClient(h, conn)
	h.register <- client

	go client.read()
	go client.write()
}
