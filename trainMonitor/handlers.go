package trainMonitor

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

const (
	socketReadSize    = 1024
	socketWriteSize   = 1024
	clientChannelSize = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  socketReadSize,
	WriteBufferSize: socketWriteSize,
	CheckOrigin:     defaultCheckOrigin,
}

func defaultCheckOrigin(r *http.Request) bool {
	return true
}

func SendingMonitor(monitor Monitor, w http.ResponseWriter, r *http.Request) error {
	var trainLog TrainLog
	err := trainLog.Bind(r)
	if err != nil {
		return err
	}
	serveMonitor(monitor, w, r)

	return nil
}

func serveMonitor(monitor Monitor, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := Client{
		conn:    conn,
		channel: make(chan []byte, clientChannelSize),
	}
	err = client.Send(monitor)
	if err != nil {
		log.Println(err)
		return
	}

	go client.write()
}
