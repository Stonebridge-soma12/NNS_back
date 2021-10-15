package datasetConfig

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"net/http"
	"nns_back/log"
	"nns_back/repository"
	"nns_back/util"
	"strconv"
)

type handler struct {
	datasetConfigRepository Repository
	projectRepository       repository.ProjectRepository
}

func NewHandler(projectRepository repository.ProjectRepository, datasetConfigRepository Repository) *handler {
	return &handler{
		projectRepository:       projectRepository,
		datasetConfigRepository: datasetConfigRepository,
	}
}

type DatasetConfigDto struct {
	DatasetId     int64                         `json:"datasetId"`
	Name          string                        `json:"name"`
	Shuffle       bool                          `json:"shuffle"`
	Label         string                        `json:"label"`
	Normalization DatasetConfigNormalizationDto `json:"normalization"`
}

type DatasetConfigNormalizationDto struct {
	Usage  bool   `json:"usage"`
	Method string `json:"method"`
}

func (d DatasetConfigDto) Validate() error {
	if d.Name == "" {
		return errors.New("name is required")
	}

	if d.Label == "" {
		return errors.New("label is required")
	}

	return nil
}

type GetDatasetConfigListResponseBody struct {
	DatasetConfigs []DatasetConfigDto `json:"datasetConfigs"`
	Pagination     util.Pagination    `json:"pagination"`
}

func (h *handler) GetDatasetConfigList(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorw("failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"))
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	projectNo, _ := strconv.Atoi(mux.Vars(r)["projectNo"])
	project, err := h.projectRepository.SelectProject(repository.ClassifiedByProjectNo(userId, projectNo))
	if err != nil {
		log.Errorf("failed to SelectProject(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	count, err := h.datasetConfigRepository.CountByProjectId(project.Id)
	if err != nil {
		log.Errorf("failed to CountByProjectId(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	pagination := util.NewPaginationFromRequest(r, count)

	datasetConfigList, err := h.datasetConfigRepository.FindAllByProjectId(project.Id, pagination.Offset(), pagination.Limit())
	if err != nil {
		log.Errorf("failed to FindAllByProjectId(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	responseBody := GetDatasetConfigListResponseBody{
		DatasetConfigs: make([]DatasetConfigDto, 0, len(datasetConfigList)),
		Pagination:     pagination,
	}

	for _, datasetConfig := range datasetConfigList {
		responseBody.DatasetConfigs = append(responseBody.DatasetConfigs, DatasetConfigDto{
			DatasetId: datasetConfig.DatasetId,
			Name:      datasetConfig.Name,
			Shuffle:   datasetConfig.Shuffle,
			Label:     datasetConfig.Label,
			Normalization: DatasetConfigNormalizationDto{
				Usage:  datasetConfig.NormalizationMethod.Valid,
				Method: datasetConfig.NormalizationMethod.String,
			},
		})
	}

	util.WriteJson(w, http.StatusOK, responseBody)
}

func (h *handler) GetDatasetConfig(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorw("failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"))
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	datasetConfigId, _ := util.Atoi64(mux.Vars(r)["datasetConfigId"])
	datasetConfig, err := h.datasetConfigRepository.FindByUserIdAndId(userId, datasetConfigId)
	if err != nil {
		log.Errorf("failed to FindByUserIdAndId(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	responseBody := DatasetConfigDto{
		DatasetId: datasetConfig.DatasetId,
		Name:      datasetConfig.Name,
		Shuffle:   datasetConfig.Shuffle,
		Label:     datasetConfig.Label,
		Normalization: DatasetConfigNormalizationDto{
			Usage:  datasetConfig.NormalizationMethod.Valid,
			Method: datasetConfig.NormalizationMethod.String,
		},
	}

	util.WriteJson(w, http.StatusOK, responseBody)
}

func (h *handler) CreateDatasetConfig(w http.ResponseWriter, r *http.Request) {
	var requestBody DatasetConfigDto
	if err := util.BindJson(r.Body, &requestBody); err != nil {
		log.Warnw("failed to bind json",
			"error", err)
		util.WriteError(w, http.StatusBadRequest, util.ErrBadRequest)
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

	projectNo, _ := strconv.Atoi(mux.Vars(r)["projectNo"])
	project, err := h.projectRepository.SelectProject(repository.ClassifiedByProjectNo(userId, projectNo))
	if err != nil {
		log.Errorf("failed to SelectProject(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	newDatasetConfig := DatasetConfig{
		ProjectId: project.Id,
		DatasetId: requestBody.DatasetId,
		Name:      requestBody.Name,
		Shuffle:   requestBody.Shuffle,
		NormalizationMethod: sql.NullString{
			Valid:  requestBody.Normalization.Usage,
			String: requestBody.Normalization.Method,
		},
		Label:  requestBody.Label,
		Status: util.StatusEXIST,
	}

	newDatasetConfig.Id, err = h.datasetConfigRepository.Insert(newDatasetConfig)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == util.MysqlErrDupEntry {
				log.Warnw("duplicate entity",
					"requested dataset config name", newDatasetConfig.Name)
				util.WriteError(w, http.StatusBadRequest, util.ErrDuplicate)
				return
			}
		}

		log.Errorf("failed to Insert(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *handler) UpdateDatasetConfig(w http.ResponseWriter, r *http.Request) {
	var requestBody DatasetConfigDto
	if err := util.BindJson(r.Body, &requestBody); err != nil {
		log.Warnw("failed to bind json",
			"error", err)
		util.WriteError(w, http.StatusBadRequest, util.ErrBadRequest)
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

	datasetConfigId, _ := util.Atoi64(mux.Vars(r)["datasetConfigId"])
	datasetConfig, err := h.datasetConfigRepository.FindByUserIdAndId(userId, datasetConfigId)
	if err != nil {
		log.Errorf("failed to FindByUserIdAndId(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	datasetConfig.DatasetId = requestBody.DatasetId
	datasetConfig.Name = requestBody.Name
	datasetConfig.Shuffle = requestBody.Shuffle
	datasetConfig.NormalizationMethod.Valid = requestBody.Normalization.Usage
	datasetConfig.NormalizationMethod.String = requestBody.Normalization.Method
	datasetConfig.Label = requestBody.Label

	if err := h.datasetConfigRepository.Update(datasetConfig); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == util.MysqlErrDupEntry {
				log.Warnw("duplicate entity",
					"requested dataset config name", requestBody.Name)
				util.WriteError(w, http.StatusBadRequest, util.ErrDuplicate)
				return
			}
		}

		log.Errorf("failed to Update(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) DeleteDatasetConfig(w http.ResponseWriter, r *http.Request) {
	
}
