package trainMonitor

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	socketReadSize  = 1024
	socketWriteSize = 1024

	epochLogFormat = "%s Epoch=%d Accuracy=%g Loss=%g Val_accuracy=%g Val_Loss=%g Learning_rate=%g"

	trainId = "train_id"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  socketReadSize,
	WriteBufferSize: socketWriteSize,
	CheckOrigin:     defaultCheckOrigin,
}

type Bridge struct {
	clients []*Client
	
	epochRepository    EpochRepository
	trainRepository    TrainRepository
	trainLogRepository TrainLogRepository
}

func NewBridge(epochRepository EpochRepository, trainRepository TrainRepository, trainLogRepository TrainLogRepository) *Bridge {
	bridge := Bridge{
		epochRepository:    epochRepository,
		trainRepository:    trainRepository,
		trainLogRepository: trainLogRepository,
	}

	return &bridge
}

func defaultCheckOrigin(r *http.Request) bool {
	return true
}

func getTrainId(r *http.Request) int64 {
	tid, _ := strconv.ParseInt(r.Header.Get(trainId), 10, 0)

	return tid
}

type Client struct {
	conn      *websocket.Conn
	channel   chan *Monitor
	TrainId   int64
	ProjectId int64
}

func (b *Bridge) NewEpochHandler(w http.ResponseWriter, r *http.Request) {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	tid := getTrainId(r)

	var epoch Epoch
	epoch.TrainId = tid
	err := epoch.Bind(r)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(epoch)

	err = b.epochRepository.Insert(epoch)
	if err != nil {
		log.Println(err)
		return
	}

	train, err := b.trainRepository.Find(WithID(epoch.TrainId))
	if err != nil {
		log.Println(err)
		return
	}
	train.Update(epoch)

	err = b.trainRepository.Update(train)
	if err != nil {
		log.Println(err)
		return
	}

	msg := fmt.Sprintf(
		epochLogFormat,
		currentTime,
		epoch.Epoch,
		epoch.Acc,
		epoch.Loss,
		epoch.ValAcc,
		epoch.ValLoss,
		epoch.LearningRate,
	)

	trainLog := TrainLog{
		TrainId: tid,
		Message: msg,
	}

	err = b.trainLogRepository.Insert(trainLog)
	if err != nil {
		log.Println(err)
		return
	}

	monitor := Monitor{
		Epoch:    epoch,
		TrainLog: trainLog,
	}
	fmt.Println(monitor)

	b.Send(epoch.TrainId, &monitor)
}

func (b *Bridge) TrainReplyHandler(w http.ResponseWriter, r *http.Request) {
	tid := getTrainId(r)

	var trainLog TrainLog

	trainLog.TrainId = tid
	err := trainLog.Bind(r)
	if err != nil {
		log.Println(err)
		return
	}
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	trainLog.Message = currentTime + ": " + trainLog.Message
	err = b.trainLogRepository.Insert(trainLog)
	if err != nil {
		log.Println(err)
	}

	train, err := b.trainRepository.Find(WithID(tid))
	if err != nil {
		log.Println(err)
		return
	}

	if trainLog.Status == 200 {
		train.Status = "FIN"
		err = b.trainRepository.Update(train)
		if err != nil {
			log.Println(err)
			return
		}
	} else if trainLog.Status >= 400 {
		train.Status = "ERR"
		err = b.trainRepository.Update(train)
		if err != nil {
			log.Println(err)
			return
		}
	}

	fmt.Println("Train finished")

	b.Close(tid)
}

func (b *Bridge) ServeMonitorWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		conn:    conn,
		channel: make(chan *Monitor),
	}
	b.clients = append(b.clients, client)

	go client.writeLoop()
}

func (b *Bridge) Close(tid int64) {
	for i, client := range b.clients {
		if client.TrainId == tid {
			b.clients = append(b.clients[:i], b.clients[i+1:]...)
			client.conn.Close()
			close(client.channel)
			break
		}
	}
}

func (b *Bridge) Send(tid int64, monitor *Monitor) {
	for _, client := range b.clients {
		if client.TrainId == tid {
			client.channel <- monitor
			break
		}
	}
}

func (c *Client) writeLoop() {
	for {
		select {
		case msg, ok := <-c.channel:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.write(msg)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func (c *Client) write(monitor *Monitor) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(monitor)
	if err != nil {
		log.Println(err)
		return err
	}

	return c.conn.WriteMessage(websocket.BinaryMessage, buf.Bytes())
}
