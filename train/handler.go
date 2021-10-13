package train

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"nns_back/cloud"
	"nns_back/externalAPI"
	"nns_back/log"
	"nns_back/model"
	"nns_back/repository"
	"nns_back/util"
	"strconv"
	"time"
)

const (
	saveTrainedModelFormFileKey = "model"
	trainModelContentType       = "application/zip"
)

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
	TrainNo                    int64           `json:"trainNo"`
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
				TrainNo:                    history.TrainNo,
				Name:                       history.Name,
				Status:                     history.Status,
				Acc:                        history.Acc,
				Loss:                       history.Loss,
				ValAcc:                     history.ValAcc,
				ValLoss:                    history.ValLoss,
				Epochs:                     history.Epochs,
				ResultUrl:                  history.ResultUrl, // saved model url
				TrainDatasetUrl:            history.TrainConfig.TrainDatasetUrl,
				ValidDatasetUrl:            history.TrainConfig.ValidDatasetUrl,
				DatasetShuffle:             history.TrainConfig.DatasetShuffle,
				DatasetLabel:               history.TrainConfig.DatasetLabel,
				DatasetNormalizationUsage:  history.TrainConfig.DatasetNormalizationUsage,
				DatasetNormalizationMethod: history.TrainConfig.DatasetNormalizationMethod,
				ModelContent:               history.TrainConfig.ModelContent,
				ModelConfig:                history.TrainConfig.ModelConfig,
				CreateTime:                 history.TrainConfig.CreateTime,
				UpdateTime:                 history.TrainConfig.UpdateTime,
			})
		}
	}

	util.WriteJson(w, http.StatusOK, resp)
}

func (h *Handler) GetTrainHistoryEpochsHandler(w http.ResponseWriter, r *http.Request) {
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

	epochs, err := h.EpochRepository.FindAll(WithTrainUserId(userId), WithProjectProjectNo(projectNo), WithTrainTrainNo(trainNo))
	if err != nil {
		log.Warnw(
			"Can't query with userId, projectNo, trainNo",
			"error code", util.ErrInvalidQueryParm,
			"error", err,
			"input value", userId, projectNo, trainNo,
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

	// 지금은 각 유저는 한번에 하나의 학습만 가능
	// 이 유저가 현재 학습중인지 확인
	if trainable, err := isTrainable(h.TrainRepository, userId); err != nil {
		log.Errorf("failed to CountCurrentTraining(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	} else if !trainable {
		log.Warnw("current training count is maximum")
		util.WriteError(w, http.StatusBadRequest, util.ErrBadRequest)
		return
	}

	if err := startNewTrain(h.TrainRepository, h.ProjectRepository, h.Fitter, userId, projectNo, body); err != nil {
		log.Error(err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func isTrainable(trainRepository TrainRepository, userId int64) (bool, error){
	trainingCount, err := trainRepository.CountCurrentTraining(userId)
	if err != nil {
		return false, err
	}

	result := false
	if trainingCount < _numberOfTrainingLimit {
		result = true
	}

	return result, nil
}

func startNewTrain(trainRepository TrainRepository, projectRepository repository.ProjectRepository, fitter externalAPI.Fitter, userId int64, projectNo int, body NewTrainHandlerRequestBody) error {
	project, err := projectRepository.SelectProject(repository.ClassifiedByProjectNo(userId, projectNo))
	if err != nil {
		return err
	}

	nextTrainNo, err := trainRepository.FindNextTrainNo(userId)
	if err != nil {
		return err
	}

	newTrain := createNewTrain(userId, nextTrainNo, project, body)
	newTrain.Id, err = saveTrain(trainRepository, newTrain)
	if err != nil {
		return err
	}

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

	return fitRequest(fitter, payload)
}

func createNewTrain(userId int64, nextTrainNo int64, project model.Project, body NewTrainHandlerRequestBody) Train {
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

	return newTrain
}

func saveTrain(trainRepository TrainRepository, train Train) (int64, error) {
	return trainRepository.Insert(train)
}

func fitRequest(fitter externalAPI.Fitter, payload externalAPI.FitRequestBody) error {
	resp, err := fitter.Fit(payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to Fit")
	}
	return nil
}

func (h *Handler) SaveTrainModelHandler(w http.ResponseWriter, r *http.Request) {
	trainId, err := strconv.ParseInt(mux.Vars(r)["trainId"], 10, 64)
	if err != nil {
		log.Warnw(
			"failed to convert trainId to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["trainId"],
		)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
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

	fBytes, err := io.ReadAll(f)
	if err != nil {
		log.Errorf("error")
		return
	}

	url, err := h.AwsS3Uploader.UploadBytes(fBytes, cloud.WithContentType(trainModelContentType), cloud.WithExtension("zip"))
	if err != nil {
		log.Warnw(
			"Failed to save model on S3 bucket",
			"error code", util.ErrInternalServerError,
			"error", err,
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	train, err := h.TrainRepository.Find(WithEpochTrainId(trainId))
	if err != nil {
		log.Warnw(
			"Can't query with train id",
			"error code", util.ErrInvalidQueryParm,
			"error", err,
			"input value", trainId,
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
