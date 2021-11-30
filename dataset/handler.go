package dataset

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"nns_back/cloud"
	"nns_back/log"
	"nns_back/util"
	"time"
	"unicode/utf8"
)

type Handler struct {
	Repository  Repository
	AwsS3Client *cloud.AwsS3Client
	HttpClient  *http.Client
}

const _uploadDatasetFormFileKey = "dataset"

const _requestBodyTooLarge = "http: request body too large"

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// maximum upload of 10 MB files
	const maxSize = 1000 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	file, _, err := r.FormFile(_uploadDatasetFormFileKey)
	if err != nil {
		// requires handling on big file input
		if err.Error() == _requestBodyTooLarge {
			util.WriteError(w, http.StatusBadRequest, util.ErrFileTooLarge)
			return
		}
		log.Error(err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}
	defer file.Close()

	userID, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorf("failed to get userId")
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// find last dataset_no
	lastDatasetNo, err := h.Repository.FindNextDatasetNo(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			lastDatasetNo = 0
		} else {
			log.Errorf("failed to select dataset: %v", err)
			util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
			return
		}
	}

	newDataset := Dataset{
		ID:          0,
		UserID:      userID,
		DatasetNo:   lastDatasetNo + 1,
		URL:         sql.NullString{},
		OriginURL:   sql.NullString{},
		Name:        sql.NullString{},
		Description: sql.NullString{},
		Public:      sql.NullBool{},
		Status:      UPLOADING,
		ImageId:     sql.NullInt64{},
		Kind:        KindUnknown,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	newDataset.ID, err = h.Repository.Insert(newDataset)
	if err != nil {
		log.Errorf("failed to insert new dataset: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	go uploadAsync(h.AwsS3Client, file, h.Repository, newDataset)

	util.WriteJson(w, http.StatusCreated, util.ResponseBody{"id": newDataset.ID})
}

type UpdateFileConfigRequestBody struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Public      bool      `json:"public"`
	Thumbnail   Thumbnail `json:"thumbnail"`
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
		log.Errorf("failed to bind request body: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	userID, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorf("failed to get userId")
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// validate
	dataset, err := h.Repository.FindByID(body.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnw("dataset not exist",
				"id", body.Id)
			util.WriteError(w, http.StatusBadRequest, util.ErrInvalidDatasetId)
			return
		}

		log.Errorf("failed to find dataset: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	if dataset.UserID != userID {
		// inaccessible object
		log.Warnw("inaccessible dataset id",
			"id", body.Id,
			"userid", userID)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidDatasetId)
		return
	}

	switch dataset.Status {
	case UPLOADING:
		dataset.Status = UPLOADED_D
	case UPLOADED_F:
		dataset.Status = EXIST
	}
	dataset.Name = sql.NullString{String: body.Name, Valid: true}
	dataset.Description = sql.NullString{String: body.Description, Valid: true}
	dataset.Public = sql.NullBool{Bool: body.Public, Valid: true}
	dataset.UpdateTime = time.Now()
	dataset.ImageId = sql.NullInt64{Int64: body.Thumbnail.ImageId, Valid: body.Thumbnail.Valid}

	if err := h.Repository.Update(dataset.ID, dataset); err != nil {
		log.Errorf("failed to update dataset: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	_, err = h.Repository.FindDatasetFromDatasetLibraryByDatasetId(userID, dataset.ID)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("failed to FindDatasetFromDatasetLibraryByDatasetId(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	if err == sql.ErrNoRows {
		// insert into library
		if err := h.Repository.AddDatasetToDatasetLibrary(userID, dataset.ID); err != nil {
			log.Errorf("failed to AddDatasetToDatasetLibrary(): %v", err)
			util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
			return
		}
	}

	util.WriteJson(w, http.StatusOK, nil)
}

type GetListResponseBody struct {
	Datasets   []DatasetDto    `json:"datasets"`
	Pagination util.Pagination `json:"pagination"`
}

type DatasetDto struct {
	Id          int64     `json:"id"`
	DatasetNo   int64     `json:"datasetNo"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Public      bool      `json:"public"`
	CreateTime  time.Time `json:"createTime"`
	UpdateTime  time.Time `json:"updateTime"`
	Usable      bool      `json:"usable"`
	InLibrary   bool      `json:"inLibrary"`
	Thumbnail   Thumbnail `json:"thumbnail"`
	Kind        Kind      `json:"kind"`
	IsUploading bool      `json:"isUploading"`
}

type Thumbnail struct {
	Valid   bool   `json:"valid"`
	ImageId int64  `json:"imageId"`
	Url     string `json:"url"`
}

const (
	_createrName     = "createrName"
	_createrNameLike = "createrNameLike"
	_title           = "title"
	_titleLike       = "titleLike"
)

func (h *Handler) GetList(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorf("failed to get userId")
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	searchType := r.URL.Query().Get("searchType")
	//searchContent := r.URL.Query().Get("searchContent")

	var (
		count    int64
		datasets []Dataset
		err      error
	)

	switch searchType {
	//case _createrName:
	//	count, err = h.Repository.CountPublicByUserName(searchContent)
	//case _createrNameLike:
	//	count, err = h.Repository.CountPublicByUserNameLike(searchContent)
	//case _title:
	//	count, err = h.Repository.CountPublicByTitle(searchContent)
	//case _titleLike:
	//	count, err = h.Repository.CountPublicByTitleLike(searchContent)
	default:
		count, err = h.Repository.CountPublic(userId)
	}
	if err != nil {
		log.Errorf("failed to count dataset: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	pagination := util.NewPaginationFromRequest(r, count)

	switch searchType {
	//case _createrName:
	//	datasets, err = h.Repository.FindAllPublicByUserName(userId, searchContent, pagination.Offset(), pagination.Limit())
	//case _createrNameLike:
	//	datasets, err = h.Repository.FindAllPublicByUserNameLike(userId, searchContent, pagination.Offset(), pagination.Limit())
	//case _title:
	//	datasets, err = h.Repository.FindAllPublicByTitle(userId, searchContent, pagination.Offset(), pagination.Limit())
	//case _titleLike:
	//	datasets, err = h.Repository.FindAllPublicByTitleLike(userId, searchContent, pagination.Offset(), pagination.Limit())
	default:
		datasets, err = h.Repository.FindAllPublic(userId, pagination.Offset(), pagination.Limit())
	}
	if err != nil {
		log.Errorf("failed to find dataset list: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// make response body
	responseDatasetDtos := make([]DatasetDto, 0, len(datasets))
	for _, val := range datasets {
		responseDatasetDtos = append(responseDatasetDtos, DatasetDto{
			Id:          val.ID,
			DatasetNo:   val.DatasetNo,
			Name:        val.Name.String,
			Description: val.Description.String,
			Public:      val.Public.Bool,
			CreateTime:  val.CreateTime,
			UpdateTime:  val.UpdateTime,
			Usable:      val.Usable.Bool,
			InLibrary:   val.InLibrary.Bool,
			Thumbnail: Thumbnail{
				Valid:   val.ImageId.Valid,
				ImageId: val.ImageId.Int64,
				Url:     val.ThumbnailUrl.String,
			},
			Kind:        val.Kind,
			IsUploading: val.Status != EXIST,
		})
	}

	response := GetListResponseBody{
		Datasets:   responseDatasetDtos,
		Pagination: pagination,
	}

	util.WriteJson(w, http.StatusOK, response)
}

func (h *Handler) DeleteDataset(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorf("failed to get userId")
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	datasetId, _ := util.Atoi64(mux.Vars(r)["datasetId"])

	// exist check
	dataset, err := h.Repository.FindByID(datasetId)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnw("invalid datasetId",
				"id", datasetId)
			util.WriteError(w, http.StatusBadRequest, util.ErrInvalidDatasetId)
			return
		}
		log.Errorf("failed to find dataset: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	if dataset.UserID != userId {
		// inaccessible object
		log.Warnw("inaccessible dataset",
			"id", datasetId,
			"userId", userId)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidDatasetId)
		return
	}

	// delete dataset
	if err := h.Repository.Delete(datasetId); err != nil {
		log.Errorf("failed to delete dataset: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// TODO: Add author data to response body
func (h *Handler) GetLibraryList(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorf("failed to get userId")
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	cnt, err := h.Repository.CountDatasetLibraryByUserId(userId)
	if err != nil {
		log.Errorf("CountDatasetLibraryByUserId(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	pagination := util.NewPaginationFromRequest(r, cnt)

	libraryContents, err := h.Repository.FindDatasetFromDatasetLibraryByUserId(userId, pagination.Offset(), pagination.Limit())
	if err != nil {
		log.Errorf("FindDatasetFromDatasetLibraryByUserId(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// make response body
	datasets := make([]DatasetDto, 0, len(libraryContents))
	for _, val := range libraryContents {
		datasets = append(datasets, DatasetDto{
			Id:          val.ID,
			DatasetNo:   val.DatasetNo,
			Name:        val.Name.String,
			Description: val.Description.String,
			Public:      val.Public.Bool,
			CreateTime:  val.CreateTime,
			UpdateTime:  val.UpdateTime,
			Usable:      val.Usable.Bool,
			InLibrary:   val.InLibrary.Bool,
			Thumbnail: Thumbnail{
				Valid:   val.ImageId.Valid,
				ImageId: val.ImageId.Int64,
				Url:     val.ThumbnailUrl.String,
			},
			Kind:        val.Kind,
			IsUploading: val.Status != EXIST,
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
	if err := util.BindJson(r.Body, &body); err != nil {
		log.Errorf("Failed to bind json body: %v", err)
		util.WriteError(w, http.StatusBadRequest, util.ErrBadRequest)
		return
	}

	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorf("failed to get userId")
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// datasetId validation: is dataset public or uploaded by me
	toAddDataset, err := h.Repository.FindByID(body.DatasetId)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnw("invalid datasetId",
				"requested datasetId", body.DatasetId)
			util.WriteError(w, http.StatusBadRequest, util.ErrInvalidDatasetId)
			return
		}

		log.Errorf("failed to FindByID(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	if !toAddDataset.Public.Bool && toAddDataset.UserID != userId {
		// this dataset is inaccessible
		log.Warnw("invalid datasetId",
			"requested datasetId", body.DatasetId)
		util.WriteError(w, http.StatusBadRequest, util.ErrBadRequest)
		return
	}

	// check duplicate
	if _, err := h.Repository.FindDatasetFromDatasetLibraryByDatasetId(userId, toAddDataset.ID); err != sql.ErrNoRows {
		if err == nil {
			// already exist
			log.Warnw("duplicated datasetId",
				"requested datasetId", body.DatasetId)
			util.WriteError(w, http.StatusBadRequest, util.ErrDuplicate)
			return
		} else {
			log.Errorf("failed to FindDatasetFromDatasetLibraryByDatasetId(): %v", err)
			util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
			return
		}
	}

	if err := h.Repository.AddDatasetToDatasetLibrary(userId, toAddDataset.ID); err != nil {
		log.Errorf("failed to AddDatasetToDatasetLibrary(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteDatasetFromLibrary(w http.ResponseWriter, r *http.Request) {
	datasetId, _ := util.Atoi64(mux.Vars(r)["datasetId"])

	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorf("failed to get userId")
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// find dataset from library
	toDeleteDatasetFromDatasetLibrary, err := h.Repository.FindDatasetFromDatasetLibraryByDatasetId(userId, datasetId)
	if err != nil {
		if err == sql.ErrNoRows {
			// invalid datasetId: dataset not exist in my library
			log.Warnw("invalid datasetId",
				"requested datasetId", datasetId)
			util.WriteError(w, http.StatusBadRequest, util.ErrNotFound)
			return
		}

		log.Errorf("failed to FindDatasetFromDatasetLibraryByDatasetId(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// delete dataset from library
	if err := h.Repository.DeleteDatasetFromDatasetLibrary(userId, toDeleteDatasetFromDatasetLibrary.ID); err != nil {
		log.Errorf("failed to DeleteDatasetFromDatasetLibrary(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type DatasetDetailDto struct {
	DatasetNum int        `json:"datasetNum"`
	FeatureNum int        `json:"featureNum"`
	Feature    []string   `json:"feature"`
	Rows       [][]string `json:"rows"`
	Kind       Kind       `json:"kind"`
}

func (h *Handler) GetDatasetDetail(w http.ResponseWriter, r *http.Request) {
	datasetId, _ := util.Atoi64(mux.Vars(r)["datasetId"])

	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorf("failed to get userId")
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	ds, err := h.Repository.FindDatasetFromDatasetLibraryByDatasetId(userId, datasetId)
	if err != nil {
		if err == sql.ErrNoRows {
			// invalid datasetId: dataset not exist in my library
			log.Warnw("invalid datasetId",
				"requested datasetId", datasetId)
			util.WriteError(w, http.StatusBadRequest, util.ErrNotFound)
			return
		}

		log.Errorf("failed to FindDatasetFromDatasetLibraryByDatasetId(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// 아직 업로드 완료되지 않은 데이터셋을 미리보기하려하면 다른 응답 내려줌
	if ds.Status != EXIST {
		util.WriteError(w, http.StatusBadRequest, "Dataset upload not complete yet")
		return
	}

	resp, err := h.HttpClient.Get(ds.URL.String)
	if err != nil {
		log.Errorw("failed to http get",
			"error", err,
			"url", ds.URL)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}
	defer resp.Body.Close()

	csvReader := csv.NewReader(resp.Body)
	records, err := readRecords(csvReader)
	if err != nil {
		log.Errorf("failed to read ReadAll(): %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	responseBody := DatasetDetailDto{
		DatasetNum: len(records) - 1,
		FeatureNum: len(records[0]),
		Feature:    records[0],
		Rows:       records[1:],
		Kind:       ds.Kind,
	}
	util.WriteJson(w, http.StatusOK, responseBody)
}

func readRecords(reader *csv.Reader) (records [][]string, err error) {
	const maxRecordsLen = 100
	for i := 0; i < maxRecordsLen+1; i++ {
		record, err := reader.Read()
		if err == io.EOF {
			return records, nil
		}
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
