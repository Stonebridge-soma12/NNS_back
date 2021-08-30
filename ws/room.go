package ws

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"log"
	"nns_back/ws/message"
)

// room maintains the set of active clients and broadcasts messages to the
// clients.
type room struct {
	id string

	projectContent map[string]interface{}

	// ID of next registered client
	nextID int

	// Registered clients.
	clients []*Client

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newRoom(id string, projectContent map[string]interface{}) *room {
	return &room{
		id:             id,
		projectContent: projectContent,
		nextID:         0,
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		clients:        make([]*Client, 0),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.register:
			r.onRegister(client)
		case client := <-r.unregister:
			r.onUnregister(client)
		}
	}
}

//func (r *room) close() {
//	for _, c := range r.clients {
//		c.close()
//	}
//
//	close(r.register)
//	close(r.unregister)
//}

func (r *room) onRegister(client *Client) {
	r.nextID++
	client.id = r.nextID
	client.Color = generateColor()
	r.clients = append(r.clients, client)

	users := make([]message.User, 0, len(r.clients))
	for _, c := range r.clients {
		users = append(users, message.User{
			ID:    c.id,
			Name:  c.Name,
			Color: c.Color,
		})
	}

	// notify that a user joined
	if msg, err := json.Marshal(message.NewUserList(users)); err == nil {
		r.broadcast(msg, nil)
	} else {
		log.Println(err)
	}

	// notify current project to recent joined user
	if msg, err := json.Marshal(
		message.NewUserCreate(message.User{client.id, client.Name, client.Color}, r.projectContent)); err == nil {
		r.send(msg, client)
	} else {
		log.Println(err)
	}
}

func (r *room) onUnregister(client *Client) {
	client.close()

	// find index of client
	index := -1
	for idx, c := range r.clients {
		if c.id == client.id {
			index = idx
			break
		}
	}

	// delete client from list
	copy(r.clients[index:], r.clients[index+1:])
	r.clients[len(r.clients)-1] = nil
	r.clients = r.clients[:len(r.clients)-1]

	//// close the r if client not exist
	//if len(r.clients) == 0 {
	//	r.close()
	//}

	// notify that a user left
	users := make([]message.User, 0, len(r.clients))
	for _, c := range r.clients {
		users = append(users, message.User{
			ID:    c.id,
			Name:  c.Name,
			Color: c.Color,
		})
	}

	if msg, err := json.Marshal(message.NewUserList(users)); err == nil {
		r.broadcast(msg, nil)
	} else {
		log.Println(err)
	}
}

func (r *room) onMessage(data []byte, reader *Client) {
	messageType := gjson.GetBytes(data, message.MessageTypeJsonTag).String()

	switch message.MessageType(messageType) {
	case message.TypeUserCreate:
		// invalid message type
		// server to client only

	case message.TypeUserList:
		// invalid message type
		// server to client only

	case message.TypeCursorMove:
		r.broadcast(data, reader)

	case message.TypeBlockCreate:
		body := message.BlockCreate{}
		if err := json.Unmarshal(data, &body); err != nil {
			log.Println()
		}

		elements := r.projectContent["flowState"].(map[string]interface{})["elements"].([]interface{})
		elements = append(elements, body.Block)
		r.projectContent["flowState"].(map[string]interface{})["elements"] = elements

		r.broadcast(data, reader)

	case message.TypeBlockRemove:
		r.broadcast(data, reader)

	case message.TypeBlockMove:
		r.broadcast(data, reader)

	case message.TypeBlockChange:
		r.broadcast(data, reader)

	case message.TypeEdgeCreate:
		r.broadcast(data, reader)

	case message.TypeEdgeRemove:
		r.broadcast(data, reader)

	default:
		// invalid message type
		r.broadcast(data, reader)
	}
}

func (r *room) broadcast(message []byte, ignore *Client) {
	for _, c := range r.clients {
		if ignore != nil && c.id == ignore.id {
			continue
		}

		c.send <- message
	}
}

func (r *room) send(message []byte, client *Client) {
	client.send <- message
}
