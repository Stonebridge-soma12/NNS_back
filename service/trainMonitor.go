package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"nns_back/trainMonitor"
	"time"
)

const (
	epochLogFormat = "%s Epoch=%d Accuracy=%g Loss=%g Val_accuracy=%g Val_Loss=%g Learning_rate=%g"
)

func (e Env) NewEpochAndLogHandler(w http.ResponseWriter, r *http.Request) {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	var epoch trainMonitor.Epoch
	epochRepo := trainMonitor.EpochDbRepository{
		DB: e.DB,
	}
	var byteBody []byte
	var err error

	epoch.TrainId = r.Header.Get("train_id")

	_, err = r.Body.Read(byteBody)
	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrMsg(err.Error()))
		return
	}

	err = json.Unmarshal(byteBody, &epoch)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrMsg(err.Error()))
		return
	}

	err = epochRepo.Insert(epoch)
	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrMsg(err.Error()))
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

	trainLog := trainMonitor.TrainLog {
		TrainId: epoch.TrainId,
		Message: msg,
	}
	logRepo := trainMonitor.TrainLogDbRepository{
		DB: e.DB,
	}

	err = logRepo.Insert(trainLog)
	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrMsg(err.Error()))
		return
	}

	monitor := trainMonitor.Monitor {
		Epoch: epoch,
		TrainLog: trainLog,
	}

	monitor.Send()
}
