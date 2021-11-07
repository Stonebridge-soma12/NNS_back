package train

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"nns_back/cloud"
	"nns_back/dataset"
	"nns_back/datasetConfig"
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
	Fitter                  externalAPI.Fitter
	ProjectRepository       repository.ProjectRepository
	TrainRepository         TrainRepository
	EpochRepository         EpochRepository
	DatasetRepository       dataset.Repository
	DatasetConfigRepository datasetConfig.Repository
	TrainLogRepository      TrainLogRepository
	AwsS3Uploader           cloud.AwsS3Uploader
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

type trainHistoryEpochListResponseBodyBody struct {
	EpochNo      int     `json:"epochNo"`
	Acc          float64 `json:"acc"`
	Loss         float64 `json:"loss"`
	ValAcc       float64 `json:"valAcc"`
	ValLoss      float64 `json:"valLoss"`
	LearningRate float64 `json:"learningRate"`
}

type trainHistoryEpochListResponseBody struct {
	Epochs []trainHistoryEpochListResponseBodyBody `json:"epochs"`
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

	var resp trainHistoryEpochListResponseBody
	for _, epoch := range epochs {
		resp.Epochs = append(resp.Epochs, trainHistoryEpochListResponseBodyBody{
			EpochNo:      epoch.Epoch,
			Acc:          epoch.Acc,
			Loss:         epoch.Loss,
			ValAcc:       epoch.ValAcc,
			ValLoss:      epoch.ValLoss,
			LearningRate: epoch.LearningRate,
		})
	}

	util.WriteJson(w, http.StatusOK, resp)
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

const _numberOfTrainingLimit = 1

func (h *Handler) NewTrainHandler(w http.ResponseWriter, r *http.Request) {
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

	projectNo, _ := strconv.Atoi(mux.Vars(r)["projectNo"])

	// 지금은 각 유저는 한번에 하나의 학습만 가능
	// 이 유저가 현재 학습중인지 확인
	if trainable, err := isTrainable(h.TrainRepository, userId); err != nil {
		log.Errorf("failed to CountCurrentTraining(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	} else if !trainable {
		log.Warnw("current training count is maximum")
		util.WriteError(w, http.StatusBadRequest, util.ErrAlreadyTrainingToTheMax)
		return
	}

	// get dataset config
	project, err := h.ProjectRepository.SelectProject(repository.ClassifiedByProjectNo(userId, projectNo))
	if err != nil {
		log.Errorf("failed to SelectProject(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	datasetConfigId, err := getDatasetConfigId(project)
	if err != nil {
		log.Warnw("dataset config id is not set in project config")
		util.WriteError(w, http.StatusBadRequest, util.ErrRequiresDatasetConfigSetting)
		return
	}

	datasetConfig, err := h.DatasetConfigRepository.FindByUserIdAndId(userId, datasetConfigId)
	if err != nil {
		log.Errorf("failed to FindByuserIdAndId(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	if err := startNewTrain(h.DatasetRepository, h.TrainRepository, h.Fitter, project, datasetConfig, userId); err != nil {
		log.Error(err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func isTrainable(trainRepository TrainRepository, userId int64) (bool, error) {
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

func getDatasetConfigId(project model.Project) (int64, error) {
	if valid := gjson.GetBytes(project.Config.Json, "dataset_config").Get("valid").Bool(); !valid {
		return 0, errors.New("dataset_config id is not valid")
	}

	return gjson.GetBytes(project.Config.Json, "dataset_config").Get("id").Int(), nil
}

func startNewTrain(datasetRepository dataset.Repository, trainRepository TrainRepository, fitter externalAPI.Fitter, project model.Project, config datasetConfig.DatasetConfig, userId int64) error {
	nextTrainNo, err := trainRepository.FindNextTrainNo(userId)
	if err != nil {
		return errors.Wrapf(err, "FindNextTrainNo(userId: %d)", userId)
	}

	dataset, err := datasetRepository.FindByID(config.DatasetId)
	if err != nil {
		return errors.Wrapf(err, "FindByID(id: %d)", config.DatasetId)
	}

	newTrain := createNewTrain(userId, nextTrainNo, project, dataset, config)
	newTrain.Id, err = saveTrain(trainRepository, newTrain)
	if err != nil {
		return errors.Wrapf(err, "saveTrain(trainRepository: %v, newTrain: %v", trainRepository, newTrain)
	}

	payload := externalAPI.FitRequestBody{
		TrainId: newTrain.Id,
		UserId:  userId,
		Config:  project.Config.Json,
		Content: project.Content.Json,
		DataSet: externalAPI.FitRequestBodyDataSet{
			TrainUri:      newTrain.TrainConfig.TrainDatasetUrl,
			ValidationUri: newTrain.TrainConfig.ValidDatasetUrl.String,
			Shuffle:       newTrain.TrainConfig.DatasetShuffle,
			Label:         newTrain.TrainConfig.DatasetLabel,
			Normalization: externalAPI.FitRequestBodyDataSetNormalization{
				Usage:  newTrain.TrainConfig.DatasetNormalizationUsage,
				Method: newTrain.TrainConfig.DatasetNormalizationMethod.String,
			},
			Kind: string(dataset.Kind),
		},
		ProjectNo: project.ProjectNo,
	}

	if err := fitRequest(fitter, payload); err != nil {
		return errors.Wrapf(err, "fitRequest(fitter: %v, payload: %v", fitter, spew.Sdump(payload))
	}

	newTrain.Status = TrainStatusTrain
	if err := trainRepository.Update(newTrain); err != nil {
		return errors.Wrapf(err, "Update(train: %v)", newTrain)
	}

	return nil
}

func createNewTrain(userId int64, nextTrainNo int64, project model.Project, dataset dataset.Dataset, config datasetConfig.DatasetConfig) Train {
	newTrain := Train{
		//Id:        0,
		UserId:    userId,
		TrainNo:   nextTrainNo,
		ProjectId: project.Id,
		Status:    TrainStatusCreated,
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
			TrainDatasetUrl:           dataset.OriginURL,
			ValidDatasetUrl:           sql.NullString{},
			DatasetShuffle:            config.Shuffle,
			DatasetLabel:              config.Label,
			DatasetNormalizationUsage: config.NormalizationMethod.Valid,
			DatasetNormalizationMethod: sql.NullString{
				String: config.NormalizationMethod.String,
				Valid:  config.NormalizationMethod.Valid,
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
		return errors.Wrapf(err, "Fit(payload: %v)", payload)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("failed to Fit: response status code : %d", resp.StatusCode))
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

	train, err := h.TrainRepository.Find(WithTrainTrainId(trainId))
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

func (h *Handler) GetTrainLogListHandler(w http.ResponseWriter, r *http.Request) {
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

	trainLogs, err := h.TrainLogRepository.FindAll(WithTrainUserId(userId), WithProjectProjectNo(projectNo), WithTrainTrainNo(trainNo))
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

	type trainLogListResponseBody struct {
		TrainLogs	[]TrainLog `json:"trainLogs"`
	}

	var resp trainLogListResponseBody
	for _, trainLog := range trainLogs {
		resp.TrainLogs = append(resp.TrainLogs, trainLog)
	}

	util.WriteJson(w, http.StatusOK, resp)
}
