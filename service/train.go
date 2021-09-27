package service

import (
	"github.com/gorilla/mux"
	"net/http"
	"nns_back/trainMonitor"
	"strconv"
)

type GetTrainHistoryListResponseBody struct {
	TrainHistories []GetTrainHistoryListResponseTrainBody `json:"history"`
}

type GetTrainHistoryListResponseTrainBody struct {
	Name    string  `db:"name" json:"name"`
	Acc     float64 `db:"acc" json:"acc"`
	Loss    float64 `db:"loss" json:"loss"`
	ValAcc  float64 `db:"val_acc" json:"val_acc"`
	ValLoss float64 `db:"val_loss" json:"val_loss"`
	Epochs  int     `db:"epochs" json:"epochs"`
	Url     string  `db:"url" json:"url"` // saved model url
}

func (e Env) GetTrainHistoryList(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		e.Logger.Errorw(
			"failed to conversion interface to int64",
			"error code", ErrInternalServerError,
			"context value", r.Context().Value("userId"),
			)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		e.Logger.Warnw(
			"failed to convert projectNo to int",
			"error code", ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"],
			)
		writeError(w, http.StatusBadRequest, ErrInvalidPathParm)
		return
	}

	repo := trainMonitor.TrainDbRepository{
		DB: e.DB,
	}

	trainList, err := repo.FindAll(trainMonitor.WithUserIdAndProjectNo(userId, projectNo))
	if err != nil {
		e.Logger.Warnw(
			"Can't query with userId or projectNo",
			"error code", ErrInvalidQueryParm,
			"error", err,
			"input value", userId, projectNo,
		)
		writeError(w, http.StatusInternalServerError, ErrInvalidQueryParm)
	}

	resp := GetTrainHistoryListResponseBody{
		TrainHistories: make([]GetTrainHistoryListResponseTrainBody, 0, len(trainList)),
	}

	for _, history := range trainList {
		resp.TrainHistories = append(resp.TrainHistories, GetTrainHistoryListResponseTrainBody{
			Name: history.Name,
			Acc: history.Acc,
			Loss: history.Loss,
			ValAcc: history.ValAcc,
			ValLoss: history.ValLoss,
			Epochs: history.Epochs,
			Url: history.Url,
		})
	}

	writeJson(w, http.StatusOK, resp)
}