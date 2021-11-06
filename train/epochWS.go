package train

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
	"nns_back/log"
	"nns_back/util"
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
	clients map[int64]*Client

	epochRepository    EpochRepository
	trainRepository    TrainRepository
	trainLogRepository TrainLogRepository
}

func NewBridge(epochRepository EpochRepository, trainRepository TrainRepository, trainLogRepository TrainLogRepository) *Bridge {
	bridge := Bridge{
		clients:            map[int64]*Client{},
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
	tid, _ := strconv.ParseInt(r.Header.Get("trainId"), 10, 0)

	return tid
}

type Client struct {
	conn    *websocket.Conn
	send    chan *Monitor
	TrainId int64
}

func (b *Bridge) NewEpochHandler(w http.ResponseWriter, r *http.Request) {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	tid, err := strconv.ParseInt(mux.Vars(r)["trainId"], 10, 0)

	log.Debug(tid)

	var epoch Epoch
	epoch.TrainId = tid
	err = epoch.Bind(r)
	if err != nil {
		log.Error(err)
		return
	}

	log.Debug(epoch)

	err = b.epochRepository.Insert(epoch)
	if err != nil {
		log.Error(err)
		return
	}

	train, err := b.trainRepository.Find(WithTrainTrainId(epoch.TrainId))
	if err != nil {
		log.Error(err)
		return
	}
	train.Update(epoch)

	err = b.trainRepository.Update(train)
	if err != nil {
		log.Error(err)
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
		TrainId:    tid,
		Message:    msg,
		StatusCode: 200,
	}

	err = b.trainLogRepository.Insert(trainLog)
	if err != nil {
		log.Error(err)
		return
	}

	monitor := Monitor{
		Epoch:    epoch,
		TrainLog: trainLog,
	}
	log.Debug(monitor)

	b.Send(epoch.TrainId, &monitor)
}

func (b *Bridge) TrainLogHandler(w http.ResponseWriter, r *http.Request) {
	var trainLog TrainLog
	err := trainLog.Bind(r)
	if err != nil {
		log.Debug(err)
		return
	}

	log.Debug(trainLog)

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	trainLog.Message = currentTime + ": " + trainLog.Message
	err = b.trainLogRepository.Insert(trainLog)
	if err != nil {
		log.Debug(err)
	}

	monitor := Monitor{
		TrainLog: trainLog,
	}
	log.Debug(monitor)

	b.Send(trainLog.TrainId, &monitor)
}

func (b *Bridge) TrainReplyHandler(w http.ResponseWriter, r *http.Request) {
	var trainLog TrainLog
	err := trainLog.Bind(r)
	if err != nil {
		log.Error(err)
		return
	}

	log.Debug(trainLog)

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	trainLog.Message = currentTime + ": " + trainLog.Message
	err = b.trainLogRepository.Insert(trainLog)
	if err != nil {
		log.Error(err)
	}

	train, err := b.trainRepository.Find(WithTrainTrainId(trainLog.TrainId))
	if err != nil {
		log.Error(err)
		return
	}

	if trainLog.StatusCode == 200 {
		train.Status = TrainStatusFinish
		err = b.trainRepository.Update(train)
		if err != nil {
			log.Error(err)
			return
		}
	} else if trainLog.StatusCode >= 400 {
		train.Status = TrainStatusError
		err = b.trainRepository.Update(train)
		if err != nil {
			log.Error(err)
			return
		}
	}

	log.Debug("Train finished")

	b.Close(trainLog.TrainId)
}

func (b *Bridge) MonitorWsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}

	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorw("failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"))
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	vars := mux.Vars(r)
	projectNo, _ := strconv.Atoi(vars["projectNo"])
	trainNo, _ := strconv.Atoi(vars["trainNo"])

	log.Debug(userId, projectNo, trainNo)

	train, err := b.trainRepository.Find(WithTrainUserId(userId), WithProjectProjectNo(projectNo), WithTrainTrainNo(trainNo))
	if err != nil {
		log.Error(err)
	}

	log.Debug(train)

	log.Debugf("Train id: %d", train.Id)

	client := &Client{
		conn:    conn,
		send:    make(chan *Monitor),
		TrainId: train.Id,
	}

	b.clients[train.Id] = client

	go client.writePump()
}

func (b *Bridge) Close(tid int64) {
	if client, ok := b.clients[tid]; ok {
		client.conn.Close()
		close(client.send)
		delete(b.clients, tid)
	}
}

func (b *Bridge) Send(tid int64, monitor *Monitor) {
	if client, ok := b.clients[tid]; ok {
		client.send <- monitor
	}
}

func (c *Client) writePump() {
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.write(msg)
			if err != nil {
				log.Error(err)
				return
			}
		}
	}
}

func (c *Client) write(monitor *Monitor) error {
	return c.conn.WriteJSON(monitor)
}
