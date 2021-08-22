package ws

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Hub struct {
	rooms map[string]*room
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*room),
	}
}

func (h *Hub) WsHandler(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]

	if _, exist := h.rooms[key]; !exist {
		h.rooms[key] = newRoom(key)
		go h.rooms[key].run()
	}

	serveWs(h.rooms[key], w, r)
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

// serveWs handles websocket requests from the peer.
func serveWs(room *room, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{room: room, conn: conn, send: make(chan []byte, 256)}
	client.room.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

