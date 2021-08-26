package ws

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"nns_back/model"
)

type Hub struct {
	rooms map[string]*room
	DB    *sqlx.DB
}

func NewHub(db *sqlx.DB) *Hub {
	return &Hub{
		rooms: make(map[string]*room),
		DB:    db,
	}
}

func (h *Hub) WsHandler(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Println("failed to conversion interface to int64")
		http.Error(w, "failed to conversion interface to int64", http.StatusInternalServerError)
		return
	}

	user, err := model.SelectUser(h.DB, model.ClassifiedById(userId))
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to select user", http.StatusInternalServerError)
		return
	}

	project, err := model.SelectProject(h.DB, model.ClassifiedByShareKey(key))
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			log.Println(err)
			http.Error(w, "failed to select project", http.StatusInternalServerError)
			return
		}
	}

	projectContent := make(map[string]interface{})
	if err := json.Unmarshal(project.Content.Json, &projectContent); err != nil {
		log.Println(err)
		http.Error(w, "failed to unmarshal project content", http.StatusInternalServerError)
		return
	}

	if _, exist := h.rooms[key]; !exist {
		h.rooms[key] = newRoom(key, projectContent)
		go h.rooms[key].run()
	}

	serveWs(h.rooms[key], user.Name, w, r)
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
func serveWs(room *room, clientName string, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{Name: clientName, room: room, conn: conn, send: make(chan []byte, 256)}
	client.room.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
