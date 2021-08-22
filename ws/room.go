package ws

// room maintains the set of active clients and broadcasts messages to the
// clients.
type room struct {
	id string

	// ID of next registered client
	nextID int

	// Registered clients.
	clients []*Client

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newRoom(id string) *room {
	return &room{
		id:         id,
		nextID:     0,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make([]*Client, 0),
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
	r.clients = append(r.clients, client)
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
}

func (r *room) broadcast(message []byte, ignore *Client) {
	for _, c := range r.clients {
		if ignore != nil && c.id == ignore.id {
			continue
		}

		c.send <- message
	}
}
