package dataset

import (
	"database/sql"
	"go.uber.org/zap"
	"net/http"
	"nns_back/cloud"
	"nns_back/util"
	"time"
)

type Handler struct {
	Repository  Repository
	Logger      *zap.SugaredLogger
	AwsS3Client *cloud.AwsS3Client
}

const _uploadDatasetFormFileKey = "dataset"

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {

	// maximum upload of 10 MB files
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.Logger.Error(err)
		return
	}

	file, _, err := r.FormFile(_uploadDatasetFormFileKey)
	if err != nil {
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