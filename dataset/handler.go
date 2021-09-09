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

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// maximum upload of 10 MB files
	const maxSize = 10 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	file, _, err := r.FormFile(_uploadDatasetFormFileKey)
	if err != nil {
		// requires handling on big file input
		if err.Error() == "http: request body too large" {
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
		h.Logger.Errorf("failed to get userId: %v", err)
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
