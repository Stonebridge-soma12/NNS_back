package ws

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"nns_back/log"
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
		log.Error(err)
	}

	// notify current project to recent joined user
	if msg, err := json.Marshal(
		message.NewUserCreate(message.User{client.id, client.Name, client.Color}, r.projectContent)); err == nil {
		r.send(msg, client)
	} else {
		log.Error(err)
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
		log.Error(err)
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
			log.Error(err)
		}

		elements := r.projectContent["flowState"].(map[string]interface{})["elements"].([]interface{})
		elements = append(elements, body.Block)

		r.projectContent["flowState"].(map[string]interface{})["elements"] = elements

		r.broadcast(data, reader)

	case message.TypeBlockRemove:
		body := message.BlockRemove{}
		if err := json.Unmarshal(data, &body); err != nil {
			log.Error(err)
		}

		elements := r.projectContent["flowState"].(map[string]interface{})["elements"].([]interface{})
		var index int // to remove element index from elements
		for idx, element := range elements {
			if element.(map[string]interface{})["id"] == body.BlockID {
				index = idx
				break
			}
		}

		// delete element from elements
		copy(elements[index:], elements[index+1:])
		elements[len(elements)-1] = nil
		elements = elements[:len(elements)-1]

		r.projectContent["flowState"].(map[string]interface{})["elements"] = elements

		r.broadcast(data, reader)

	case message.TypeBlockMove:
		body := message.BlockMove{}
		if err := json.Unmarshal(data, &body); err != nil {
			log.Error(err)
		}

		elements := r.projectContent["flowState"].(map[string]interface{})["elements"].([]interface{})
		for _, element := range elements {
			if element.(map[string]interface{})["id"] == body.BlockID {
				element.(map[string]interface{})["position"] = body.Position
			}
		}

		r.projectContent["flowState"].(map[string]interface{})["elements"] = elements

		r.broadcast(data, reader)

	case message.TypeBlockConfigChange:
		body := message.BlockConfigChange{}
		if err := json.Unmarshal(data, &body); err != nil {
			log.Error(err)
		}

		elements := r.projectContent["flowState"].(map[string]interface{})["elements"].([]interface{})
		for _, element := range elements {
			if element.(map[string]interface{})["id"] == body.BlockID {
				element.(map[string]interface{})["data"].(map[string]interface{})["param"].(map[string]interface{})[body.Config.Name] = body.Config.Value
				break
			}
		}
		r.projectContent["flowState"].(map[string]interface{})["elements"] = elements

		r.broadcast(data, reader)

	case message.TypeBlockLabelChange:
		body := message.BlockLabelChange{}
		if err := json.Unmarshal(data, &body); err != nil {
			log.Error(err)
		}

		elements := r.projectContent["flowState"].(map[string]interface{})["elements"].([]interface{})
		for _, element := range elements {
			if element.(map[string]interface{})["id"] == body.BlockID {
				element.(map[string]interface{})["data"].(map[string]interface{})["label"] = body.Data
				break
			}
		}
		r.projectContent["flowState"].(map[string]interface{})["elements"] = elements
		r.broadcast(data, reader)

	case message.TypeEdgeCreate:
		body := message.EdgeCreate{}
		if err := json.Unmarshal(data, &body); err != nil {
			log.Error(err)
		}

		r.projectContent["flowState"].(map[string]interface{})["elements"] = body.Elements
		r.broadcast(data, reader)

	case message.TypeEdgeUpdate:
		body := message.EdgeUpdate{}
		if err := json.Unmarshal(data, &body); err != nil {
			log.Error(err)
		}

		r.projectContent["flowState"].(map[string]interface{})["elements"] = body.Elements
		r.broadcast(data, reader)

	case message.TypeChat:
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
