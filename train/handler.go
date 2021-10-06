package train

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"net/http"
	"nns_back/cloud"
	"nns_back/externalAPI"
	"nns_back/log"
	"nns_back/repository"
	"nns_back/util"
	"strconv"
	"time"
)

const saveTrainedModelFormFileKey = "model"

type Handler struct {
	Fitter            externalAPI.Fitter
	ProjectRepository repository.ProjectRepository
	TrainRepository   TrainRepository
	EpochRepository   EpochRepository
	AwsS3Uploader     cloud.AwsS3Uploader
}

type GetTrainHistoryListResponseBody struct {
	TrainHistories []GetTrainHistoryListResponseHistoryBody `json:"history"`
}

type GetTrainHistoryListResponseHistoryBody struct {
	Name                       string          `json:"name"`
	Status                     string          `json:"status"`
	Acc                        float64         `json:"acc"`
	Loss                       float64         `json:"loss"`
	ValAcc                     float64         `json:"valAcc"`
	ValLoss                    float64         `json:"valLoss"`
	Epochs                     int             `json:"epochs"`
	ResultUrl                  string          `json:"resultUrl"` // saved model url
	TrainDatasetUrl            string          `json:"trainDatasetUrl"`
	ValidDatasetUrl            sql.NullString  `json:"validDatasetUrl"`
	DatasetShuffle             bool            `json:"datasetShuffle"`
	DatasetLabel               string          `json:"datasetLabel"`
	DatasetNormalizationUsage  bool            `json:"datasetNormalizationUsage"`
	DatasetNormalizationMethod sql.NullString  `json:"datasetNormalizationMethod"`
	ModelContent               json.RawMessage `json:"modelContent"`
	ModelConfig                json.RawMessage `json:"modelConfig"`
	CreateTime                 time.Time       `json:"createTime"`
	UpdateTime                 time.Time       `json:"updateTime"`
}

func (h *Handler) GetTrainHistoryListHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorw(
			"failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"),
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		log.Warnw(
			"failed to convert projectNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"],
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
		return
	}

	trainList, err := h.TrainRepository.FindAll(WithProjectUserId(userId), WithProjectProjectNo(projectNo), WithoutTrainStatusDel(), WithPagenation(0, 100))
	if err != nil {
		log.Warnw(
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
				Name: history.Name,
				Status: history.Status,
				Acc: history.Acc,
				Loss: history.Loss,
				ValAcc: history.ValAcc,
				ValLoss: history.ValLoss,
				Epochs: history.Epochs,
				ResultUrl: history.ResultUrl,// saved model url
				TrainDatasetUrl: history.TrainConfig.TrainDatasetUrl,
				ValidDatasetUrl: history.TrainConfig.ValidDatasetUrl,
				DatasetShuffle: history.TrainConfig.DatasetShuffle,
				DatasetLabel: history.TrainConfig.DatasetLabel,
				DatasetNormalizationUsage: history.TrainConfig.DatasetNormalizationUsage,
				DatasetNormalizationMethod: history.TrainConfig.DatasetNormalizationMethod,
				ModelContent: history.TrainConfig.ModelContent,
				ModelConfig: history.TrainConfig.ModelConfig,
				CreateTime: history.TrainConfig.CreateTime,
				UpdateTime: history.TrainConfig.UpdateTime,
			})
		}
	}

	util.WriteJson(w, http.StatusOK, resp)
}

func (h *Handler) GetRainHistoryEpochsHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorw(
			"failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"),
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		log.Warnw(
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
		log.Warnw(
			"failed to convert trainNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["trainNo"],
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
		return
	}

	train, err := h.TrainRepository.Find(WithTrainUserId(userId), WithProjectProjectNo(projectNo), WithTrainTrainNo(trainNo))
	if err != nil {
		log.Warnw(
			"Can't query with userId or projectNo or trainNo",
			"error code", util.ErrInvalidQueryParm,
			"error", err,
			"input value", userId, projectNo, trainNo,
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInvalidQueryParm)
		return
	}

	epochs, err := h.EpochRepository.FindAll(WithTrainID(train.Id))
	if err != nil {
		log.Warnw(
			"Can't query with train id",
			"error code", util.ErrInvalidQueryParm,
			"error", err,
			"input value", train.Id,
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInvalidQueryParm)
		return
	}

	util.WriteJson(w, http.StatusOK, epochs)
}

func (h *Handler) DeleteTrainHistoryHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorw(
			"failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"),
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		log.Warnw(
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
		log.Warnw(
			"failed to convert trainNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["trainNo"],
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
		return
	}

	err = h.TrainRepository.Delete(WithProjectUserId(userId), WithProjectProjectNo(projectNo), WithTrainTrainNo(trainNo))
	if err != nil {
		log.Warnw(
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
		log.Errorw(
			"failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"),
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		log.Warnw(
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
		log.Warnw(
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
		log.Warnw(
			"failed to bind train",
			"error code", util.ErrInvalidRequestBody,
			"error", err,
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidRequestBody)
		return
	}

	train, err := h.TrainRepository.Find(WithTrainUserId(userId), WithProjectProjectNo(projectNo), WithTrainTrainNo(trainNo))
	if err != nil {
		log.Warnw(
			"Can't query with userId or projectNo or trainNo",
			"error code", util.ErrInvalidQueryParm,
			"error", err,
			"input value", userId, projectNo, trainNo,
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInvalidQueryParm)
		return
	}

	train.Name = reqBody.Name

	err = h.TrainRepository.Update(train)
	if err != nil {
		log.Warnw(
			"Can't update this record",
			"error code", util.ErrInternalServerError,
			"error", err,
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type NewTrainHandlerRequestBody struct {
	Dataset struct {
		TrainUrl      string `json:"trainUrl"`
		ValidateUrl   string `json:"validateUrl"`
		Shuffle       bool   `json:"shuffle"`
		Label         string `json:"label"`
		Normalization struct {
			Usage  bool   `json:"usage"`
			Method string `json:"method"`
		} `json:"normalization"`
	} `json:"dataset"`
}

func (n NewTrainHandlerRequestBody) Validate() error {
	if n.Dataset.TrainUrl == "" {
		return errors.New("trainUrl is required")
	}

	return nil
}

const _numberOfTrainingLimit = 1

func (h *Handler) NewTrainHandler(w http.ResponseWriter, r *http.Request) {
	body := NewTrainHandlerRequestBody{}
	if err := util.BindJson(r.Body, &body); err != nil {
		log.Warnw("failed to bind json",
			"error", err)
		util.WriteError(w, http.StatusBadRequest, util.ErrBadRequest)
		return
	}

	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorw(
			"failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"),
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// 지금은 각 유저는 한번에 하나의 학습만 가능
	// 이 유저가 현재 학습중인지 확인
	trainingCount, err := h.TrainRepository.CountCurrentTraining(userId)
	if err != nil {
		log.Errorf("failed to CountCurrentTraining(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	if trainingCount >= _numberOfTrainingLimit {
		// 지금은 새로운 학습 요청을 할 수 없음
		log.Warnw("current training count is maximum",
			"trainingCount", trainingCount)
		util.WriteError(w, http.StatusBadRequest, util.ErrBadRequest)
		return
	}

	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		log.Warnw(
			"failed to convert projectNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"],
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
		return
	}

	project, err := h.ProjectRepository.SelectProject(repository.ClassifiedByProjectNo(userId, projectNo))
	if err != nil {
		log.Errorf("failed to select project: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	nextTrainNo, err := h.TrainRepository.FindNextTrainNo(userId)
	if err != nil {
		log.Errorf("failed to FindNextTrainNo: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	newTrain := Train{
		//Id:        0,
		UserId:    userId,
		TrainNo:   nextTrainNo,
		ProjectId: project.Id,
		Status:    TrainStatusTrain,
		//Acc:       0,
		//Loss:      0,
		//ValAcc:    0,
		//ValLoss:   0,
		//Epochs:    0,
		//Name:      "",
		//ResultUrl: "",
		TrainConfig: TrainConfig{
			//Id:              0,
			//TrainId:         0,
			TrainDatasetUrl: body.Dataset.TrainUrl,
			ValidDatasetUrl: sql.NullString{
				String: body.Dataset.ValidateUrl,
				Valid:  body.Dataset.ValidateUrl != "",
			},
			DatasetShuffle:            body.Dataset.Shuffle,
			DatasetLabel:              body.Dataset.Label,
			DatasetNormalizationUsage: body.Dataset.Normalization.Usage,
			DatasetNormalizationMethod: sql.NullString{
				String: body.Dataset.Normalization.Method,
				Valid:  body.Dataset.Normalization.Method != "",
			},
			ModelContent: project.Content.Json,
			ModelConfig:  project.Config.Json,
		},
	}

	newTrain.Id, err = h.TrainRepository.Insert(newTrain)
	if err != nil {
		log.Errorf("failed to Insert train: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// make request body
	payload := externalAPI.FitRequestBody{
		TrainId: newTrain.Id,
		UserId:  userId,
		Config:  project.Config.Json,
		Content: project.Content.Json,
		DataSet: externalAPI.FitRequestBodyDataSet{
			TrainUri:      body.Dataset.TrainUrl,
			ValidationUri: body.Dataset.ValidateUrl,
			Shuffle:       body.Dataset.Shuffle,
			Label:         body.Dataset.Label,
			Normalization: externalAPI.FitRequestBodyDataSetNormalization{
				Usage:  body.Dataset.Normalization.Usage,
				Method: body.Dataset.Normalization.Method,
			},
		},
	}

	// send request
	resp, err := h.Fitter.Fit(payload)
	if err != nil {
		log.Errorw("failed to generate python code",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}
	defer resp.Body.Close()

	// response
	if resp.StatusCode != 200 {
		log.Warnw("failed to fit",
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
		log.Warnw(
			"failed to bind request body",
			"error code", util.ErrInvalidRequestBody,
			"error", err,
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidRequestBody)
		return
	}

	f, fh, err := r.FormFile(saveTrainedModelFormFileKey)
	if err != nil {
		log.Warnw(
			"Can't find form file with key", saveTrainedModelFormFileKey,
			"error code", util.ErrBadRequest,
			"error", err,
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrBadRequest)
		return
	}
	defer f.Close()

	log.Debugw("success to retrieve form file",
		"file name", fh.Filename,
		"file size", fh.Size,
		"MIME header", fh.Header)

	url, err := h.AwsS3Uploader.UploadFile(f)
	if err != nil {
		log.Warnw(
			"Failed to save model on S3 bucket",
			"error code", util.ErrInternalServerError,
			"error", err,
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	train, err := h.TrainRepository.Find(WithTrainTrainId(reqBody.TrainId))
	if err != nil {
		log.Warnw(
			"Can't query with train id",
			"error code", util.ErrInvalidQueryParm,
			"error", err,
			"input value", reqBody.TrainId,
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInvalidQueryParm)
		return
	}

	train.ResultUrl = url
	err = h.TrainRepository.Update(train)
	if err != nil {
		log.Warnw(
			"Can't update this record",
			"error code", util.ErrInternalServerError,
			"error", err,
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
