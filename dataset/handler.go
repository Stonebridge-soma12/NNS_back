package dataset

import (
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"nns_back/cloud"
	"nns_back/util"
	"time"
	"unicode/utf8"
)

type Handler struct {
	Repository  Repository
	Logger      *zap.SugaredLogger
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

	url, err := h.AwsS3Client.Put(file)
	if err != nil {
		h.Logger.Errorf("failed to put object: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	userID, ok := r.Context().Value("userId").(int64)
	if !ok {
		h.Logger.Errorf("failed to get userId")
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	newDataSet := Dataset{
		UserID:      userID,
		URL:         url,
		Name:        sql.NullString{},
		Description: sql.NullString{},
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	insertedId, err := h.Repository.Insert(newDataSet)
	if err != nil {
		h.Logger.Errorf("failed to insert new dataset: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	util.WriteJson(w, http.StatusCreated, util.ResponseBody{"id": insertedId})
}

type UpdateFileConfigRequestBody struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
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

	dataset, err := h.Repository.FindByID(body.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			h.Logger.Warnw("dataset not exist",
				"id", body.Id)
			util.WriteError(w, http.StatusBadRequest, util.ErrInvlidDatasetId)
			return
		}

		h.Logger.Errorf("failed to find dataset: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	dataset.Name = sql.NullString{String: body.Name, Valid: true}
	dataset.Description = sql.NullString{String: body.Description, Valid: true}
	dataset.UpdateTime = time.Now()

	if err := h.Repository.Update(body.Id, dataset); err != nil {
		h.Logger.Errorf("failed to update dataset: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	util.WriteJson(w, http.StatusOK, nil)
}

type GetListResponseBody struct {
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
	response := make([]GetListResponseBody, 0, len(list))
	for _, val := range list {
		response = append(response, GetListResponseBody{
			DatasetNo:   val.DatasetNo,
			Name:        val.Name.String,
			Description: val.Description.String,
			Public:      val.Public.Bool,
			CreateTime:  val.CreateTime,
			UpdateTime:  val.UpdateTime,
		})
	}

	util.WriteJson(w, http.StatusOK, util.ResponseBody{"datasets": response, "pagination": pagination})
}
