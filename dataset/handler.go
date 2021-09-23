package dataset

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"nns_back/cloud"
	"nns_back/util"
	"time"
	"unicode/utf8"
)

type Handler struct {
	Repository Repository
	Logger     *zap.SugaredLogger
	AwsS3Client *cloud.AwsS3Client
}

const _uploadDatasetFormFileKey = "dataset"

const _requestBodyTooLarge = "http: request body too large"

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// maximum upload of 10 MB files
	const maxSize = 10 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	file, _, err := r.FormFile(_uploadDatasetFormFileKey)
	if err != nil {
		// requires handling on big file input
		if err.Error() == _requestBodyTooLarge {
			util.WriteError(w, http.StatusBadRequest, util.ErrFileTooLarge)
			return
		}
		h.Logger.Error(err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}
	defer file.Close()

	url, err := save(h.AwsS3Client, file)
	if err != nil {
		if IsUnsupportedContentTypeError(err) {
			h.Logger.Warn(err)
			util.WriteError(w, http.StatusBadRequest, util.ErrUnSupportedContentType)
			return
		}

		h.Logger.Errorf("failed to save file: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	userID, ok := r.Context().Value("userId").(int64)
	if !ok {
		h.Logger.Errorf("failed to get userId")
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// find last dataset_no
	lastDatasetNo, err := h.Repository.FindNextDatasetNo(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			lastDatasetNo = 0
		} else {
			h.Logger.Errorf("failed to select dataset: %v", err)
			util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
			return
		}
	}

	newDataset := Dataset{
		ID:          0,
		UserID:      userID,
		DatasetNo:   lastDatasetNo + 1,
		URL:         url,
		Name:        sql.NullString{},
		Description: sql.NullString{},
		Public:      sql.NullBool{},
		Status:      UPLOADED,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	if _, err := h.Repository.Insert(newDataset); err != nil {
		h.Logger.Errorf("failed to insert new dataset: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	util.WriteJson(w, http.StatusCreated, util.ResponseBody{"datasetNo": newDataset.DatasetNo})
}

type UpdateFileConfigRequestBody struct {
	DatasetNo   int64  `json:"datasetNo"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
}

func (u *UpdateFileConfigRequestBody) Validate() error {
	if utf8.RuneCountInString(u.Name) > maxDatasetName {
		return fmt.Errorf("dataset name too long")
	}

	if utf8.RuneCountInString(u.Description) > maxDatasetDescription {
		return fmt.Errorf("dataset description too long")
	}

	return nil
}

func (h *Handler) UpdateFileConfig(w http.ResponseWriter, r *http.Request) {
	body := UpdateFileConfigRequestBody{}
	if err := util.BindJson(r.Body, &body); err != nil {
		h.Logger.Errorf("failed to bind request body: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	userID, ok := r.Context().Value("userId").(int64)
	if !ok {
		h.Logger.Errorf("failed to get userId")
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	dataset, err := h.Repository.FindByUserIdAndDatasetNo(userID, body.DatasetNo)
	if err != nil {
		if err == sql.ErrNoRows {
			h.Logger.Warnw("dataset not exist",
				"datasetNo", body.DatasetNo,
				"userId", userID)
			util.WriteError(w, http.StatusBadRequest, util.ErrInvlidDatasetId)
			return
		}

		h.Logger.Errorf("failed to find dataset: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	dataset.Name = sql.NullString{String: body.Name, Valid: true}
	dataset.Description = sql.NullString{String: body.Description, Valid: true}
	dataset.Status = EXIST
	dataset.Public = sql.NullBool{Bool: body.Public, Valid: true}
	dataset.UpdateTime = time.Now()

	if err := h.Repository.Update(dataset.ID, dataset); err != nil {
		h.Logger.Errorf("failed to update dataset: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	util.WriteJson(w, http.StatusOK, nil)
}

type GetListResponseBody struct {
	Datasets   []DatasetDto    `json:"datasets"`
	Pagination util.Pagination `json:"pagination"`
}

type DatasetDto struct {
	DatasetNo   int64     `json:"datasetNo"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Public      bool      `json:"public"`
	CreateTime  time.Time `json:"createTime"`
	UpdateTime  time.Time `json:"updateTime"`
}

func (h *Handler) GetList(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		h.Logger.Errorf("failed to get userId")
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	my := r.URL.Query().Get("my")
	var (
		list  []Dataset
		err   error
		count int64
	)
	if my == "true" {
		count, err = h.Repository.CountPublic()
	} else {
		count, err = h.Repository.CountByUserId(userId)
	}
	if err != nil {
		h.Logger.Errorf("failed to count dataset: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	pagination := util.NewPaginationFromRequest(r, count)

	if my == "true" {
		list, err = h.Repository.FindAllPublic(pagination.Offset(), pagination.Limit())
	} else {
		list, err = h.Repository.FindByUserId(userId, pagination.Offset(), pagination.Limit())
	}
	if err != nil {
		h.Logger.Errorf("failed to find dataset list: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// make response body
	datasets := make([]DatasetDto, 0, len(list))
	for _, val := range list {
		datasets = append(datasets, DatasetDto{
			DatasetNo:   val.DatasetNo,
			Name:        val.Name.String,
			Description: val.Description.String,
			Public:      val.Public.Bool,
			CreateTime:  val.CreateTime,
			UpdateTime:  val.UpdateTime,
		})
	}

	response := GetListResponseBody{
		Datasets:   datasets,
		Pagination: pagination,
	}

	util.WriteJson(w, http.StatusOK, response)
}

func (h *Handler) GetLibraryList(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		h.Logger.Errorf("failed to get userId")
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	cnt, err := h.Repository.CountDatasetLibraryByUserId(userId)
	if err != nil {
		h.Logger.Errorf("CountDatasetLibraryByUserId(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	pagination := util.NewPaginationFromRequest(r, cnt)

	libraryContents, err := h.Repository.FindDatasetFromDatasetLibraryByUserId(userId, pagination.Offset(), pagination.Limit())
	if err != nil {
		h.Logger.Errorf("FindDatasetFromDatasetLibraryByUserId(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// make response body
	datasets := make([]DatasetDto, 0, len(libraryContents))
	for _, val := range libraryContents {
		datasets = append(datasets, DatasetDto{
			DatasetNo:   val.DatasetNo,
			Name:        val.Name.String,
			Description: val.Description.String,
			Public:      val.Public.Bool,
			CreateTime:  val.CreateTime,
			UpdateTime:  val.UpdateTime,
		})
	}

	response := GetListResponseBody{
		Datasets:   datasets,
		Pagination: pagination,
	}

	util.WriteJson(w, http.StatusOK, response)
}

type AddNewDatasetToLibraryRequestBody struct {
	DatasetId int64 `json:"datasetId"`
}

var ErrInvalidDatasetId = errors.New("invalid datasetId")

func (a AddNewDatasetToLibraryRequestBody) Validate() error {
	if a.DatasetId == 0 {
		return ErrInvalidDatasetId
	}

	return nil
}

func (h *Handler) AddNewDatasetToLibrary(w http.ResponseWriter, r *http.Request) {
	body := AddNewDatasetToLibraryRequestBody{}
	if err := util.BindJson(r.Body, body); err != nil {
		h.Logger.Errorf("Failed to bind json body: %v", err)
		util.WriteError(w, http.StatusBadRequest, util.ErrBadRequest)
		return
	}

	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		h.Logger.Errorf("failed to get userId")
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// datasetId validation: is dataset public or uploaded by me
	toAddDataset, err := h.Repository.FindByID(body.DatasetId)
	if err != nil {
		if err == sql.ErrNoRows {
			h.Logger.Warnw("invalid datasetId",
				"requested datasetId", body.DatasetId)
			util.WriteError(w, http.StatusBadRequest, util.ErrBadRequest)
			return
		}

		h.Logger.Errorf("failed to FindByID(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	if !toAddDataset.Public.Bool && toAddDataset.UserID != userId {
		// this dataset is inaccessible
		h.Logger.Warnw("invalid datasetId",
			"requested datasetId", body.DatasetId)
		util.WriteError(w, http.StatusBadRequest, util.ErrBadRequest)
		return
	}

	// check duplicate
	if _, err := h.Repository.FindDatasetFromDatasetLibraryByDatasetId(userId, toAddDataset.ID); err != sql.ErrNoRows {
		if err == nil {
			// already exist
			h.Logger.Warnw("duplicated datasetId",
				"requested datasetId", body.DatasetId)
			util.WriteError(w, http.StatusBadRequest, util.ErrDuplicate)
			return
		} else {
			h.Logger.Errorf("failed to FindDatasetFromDatasetLibraryByDatasetId(): %v", err)
			util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
			return
		}
	}

	if err := h.Repository.AddDatasetToDatasetLibrary(userId, toAddDataset.ID); err != nil {
		h.Logger.Errorf("failed to AddDatasetToDatasetLibrary(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteDatasetFromLibrary(w http.ResponseWriter, r *http.Request) {
	datasetId, _ := util.Atoi64(mux.Vars(r)["datasetId"])

	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		h.Logger.Errorf("failed to get userId")
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// find dataset from library
	toDeleteDatasetFromDatasetLibrary, err := h.Repository.FindDatasetFromDatasetLibraryByDatasetId(userId, datasetId)
	if err != nil {
		if err == sql.ErrNoRows {
			// invalid datasetId: dataset not exist in my library
			h.Logger.Warnw("invalid datasetId",
				"requested datasetId", datasetId)
			util.WriteError(w, http.StatusBadRequest, util.ErrNotFound)
			return
		}

		h.Logger.Errorf("failed to FindDatasetFromDatasetLibraryByDatasetId(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// delete dataset from library
	if err := h.Repository.DeleteDatasetFromDatasetLibrary(userId, toDeleteDatasetFromDatasetLibrary.ID); err != nil {
		h.Logger.Errorf("failed to DeleteDatasetFromDatasetLibrary(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}