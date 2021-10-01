package train

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"nns_back/cloud"
	"nns_back/model"
	"nns_back/util"
	"strconv"
)

const saveTrainedModelFormFileKey = "model"

type Handler struct {
	DB          *sqlx.DB
	Repository  TrainRepository
	Logger      *zap.SugaredLogger
	AwsS3Client *cloud.AwsS3Client
}

type GetTrainHistoryListResponseBody struct {
	TrainHistories []GetTrainHistoryListResponseHistoryBody `json:"history"`
}

type GetTrainHistoryListResponseHistoryBody struct {
	Name    string  `db:"name" json:"name"`
	Acc     float64 `db:"acc" json:"acc"`
	Loss    float64 `db:"loss" json:"loss"`
	ValAcc  float64 `db:"val_acc" json:"val_acc"`
	ValLoss float64 `db:"val_loss" json:"val_loss"`
	Epochs  int     `db:"epochs" json:"epochs"`
	Url     string  `db:"url" json:"url"` // saved model url
}

func (h *Handler) GetTrainHistoryListHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		h.Logger.Errorw(
			"failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"),
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		h.Logger.Warnw(
			"failed to convert projectNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"],
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
		return
	}

	trainList, err := h.Repository.FindAll(WithUserIdAndProjectNo(userId, projectNo))
	if err != nil {
		h.Logger.Warnw(
			"Can't query with userId or projectNo",
			"error code", util.ErrInvalidQueryParm,
			"error", err,
			"input value", userId, projectNo,
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInvalidQueryParm)
	}

	resp := GetTrainHistoryListResponseBody{
		TrainHistories: make([]GetTrainHistoryListResponseHistoryBody, 0, len(trainList)),
	}

	for _, history := range trainList {
		if history.Status == TrainStatusFinish || history.Status == TrainStatusTrain {
			resp.TrainHistories = append(resp.TrainHistories, GetTrainHistoryListResponseHistoryBody{
				Name:    history.Name,
				Acc:     history.Acc,
				Loss:    history.Loss,
				ValAcc:  history.ValAcc,
				ValLoss: history.ValLoss,
				Epochs:  history.Epochs,
				Url:     history.Url,
			})
		}
	}

	util.WriteJson(w, http.StatusOK, resp)
}

func (h *Handler) DeleteTrainHistoryHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		h.Logger.Errorw(
			"failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"),
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		h.Logger.Warnw(
			"failed to convert projectNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"],
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
		return
	}

	trainNo, err := strconv.Atoi(mux.Vars(r)["trainNo"])
	if err != nil {
		h.Logger.Warnw(
			"failed to convert trainNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["trainNo"],
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
		return
	}

	err = h.Repository.Delete(WithUserIdAndProjectNoAndTrainNo(userId, projectNo, trainNo))
	if err != nil {
		h.Logger.Warnw(
			"Can't query with userId or projectNo or trainNo",
			"error code", util.ErrInvalidQueryParm,
			"error", err,
			"input value", userId, projectNo, trainNo,
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInvalidQueryParm)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateTrainHistoryHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		h.Logger.Errorw(
			"failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"),
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		h.Logger.Warnw(
			"failed to convert projectNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"],
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
		return
	}

	trainNo, err := strconv.Atoi(mux.Vars(r)["trainNo"])
	if err != nil {
		h.Logger.Warnw(
			"failed to convert trainNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["trainNo"],
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
		return
	}

	type newNameRequestBody struct {
		Name string `json:"name"`
	}

	var reqBody newNameRequestBody
	err = json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		h.Logger.Warnw(
			"failed to bind train",
			"error code", util.ErrInvalidRequestBody,
			"error", err,
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidRequestBody)
		return
	}

	train, err := h.Repository.Find(WithUserIdAndProjectNoAndTrainNo(userId, projectNo, trainNo))
	if err != nil {
		h.Logger.Warnw(
			"Can't query with userId or projectNo or trainNo",
			"error code", util.ErrInvalidQueryParm,
			"error", err,
			"input value", userId, projectNo, trainNo,
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInvalidQueryParm)
		return
	}

	train.Name = reqBody.Name

	err = h.Repository.Update(train)
	if err != nil {
		h.Logger.Warnw(
			"Can't update this record",
			"error code", util.ErrInternalServerError,
			"error", err,
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) NewTrainHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: 여기에 학습요청 만들면 됨
	// 요청시 필요한 내용
	// user_id, train_id

	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		h.Logger.Errorw(
			"failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"),
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		h.Logger.Warnw(
			"failed to convert projectNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"],
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
		return
	}

	project, err := model.SelectProject(h.DB, model.ClassifiedByProjectNo(userId, projectNo))
	if err != nil {
		h.Logger.Errorf("failed to select project: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	newTrain := Train{
		Id:        0,
		TrainNo:   0,
		ProjectId: project.Id,
		Status:    "",
		Acc:       0,
		Loss:      0,
		ValAcc:    0,
		ValLoss:   0,
		Epochs:    0,
		Name:      "",
		Url:       "",
	}
	newTrain.Id, err = h.Repository.Insert(newTrain)
	if err != nil {
		h.Logger.Errorf("failed to Insert train: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// http client
	defaultTransportPointer, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		h.Logger.Errorw("failed to interface conversion",
			"error code", util.ErrInternalServerError,
			"msg", "defaultRoundTripper not an *http.Transport",
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}
	defaultTransport := *defaultTransportPointer // dereference it to get a copy of the struct that the pointer points to
	defaultTransport.MaxIdleConns = 100
	defaultTransport.MaxIdleConnsPerHost = 100
	client := &http.Client{Transport: &defaultTransport}

	// make request body
	payload := make(map[string]interface{})
	payload["user_id"] = userId
	payload["train_id"] = newTrain.Id
	jsonedPayload, err := json.Marshal(payload)
	if err != nil {
		h.Logger.Errorw("failed to json marshal",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// send request
	resp, err := client.Post("http://54.180.153.56:8080/fit", "application/json", bytes.NewBuffer(jsonedPayload))
	if err != nil {
		h.Logger.Errorw("failed to generate python code",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}
	defer resp.Body.Close()

	// response
	if resp.StatusCode != 200 {
		h.Logger.Warnw("failed to fit",
			"status code", resp.StatusCode)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) SaveTrainModelHandler(w http.ResponseWriter, r *http.Request) {
	type saveModelRequestBody struct {
		UserId  int64 `json:"user_id"`
		TrainId int64 `json:"train_id"`
	}

	var reqBody saveModelRequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		h.Logger.Warnw(
			"failed to bind request body",
			"error code", util.ErrInvalidRequestBody,
			"error", err,
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidRequestBody)
		return
	}

	f, fh, err := r.FormFile(saveTrainedModelFormFileKey)
	if err != nil {
		h.Logger.Warnw(
			"Can't find form file with key", saveTrainedModelFormFileKey,
			"error code", util.ErrBadRequest,
			"error", err,
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrBadRequest)
		return
	}
	defer f.Close()

	h.Logger.Debugw("success to retrieve form file",
		"file name", fh.Filename,
		"file size", fh.Size,
		"MIME header", fh.Header)

	url, err := h.AwsS3Client.Put(f)
	if err != nil {
		h.Logger.Warnw(
			"Failed to save model on S3 bucket",
			"error code", util.ErrInternalServerError,
			"error", err,
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	train, err := h.Repository.Find(WithTrainID(reqBody.TrainId))
	if err != nil {
		h.Logger.Warnw(
			"Can't query with train id",
			"error code", util.ErrInvalidQueryParm,
			"error", err,
			"input value", reqBody.TrainId,
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInvalidQueryParm)
		return
	}

	train.Url = url
	err = h.Repository.Update(train)
	if err != nil {
		h.Logger.Warnw(
			"Can't update this record",
			"error code", util.ErrInternalServerError,
			"error", err,
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
