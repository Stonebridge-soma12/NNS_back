package service

import (
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"nns_back/externalAPI"
	"nns_back/log"
	"nns_back/model"
	"nns_back/repository"
	"nns_back/util"
	"strconv"
	"time"
	"unicode/utf8"
)

type ProjectHandler struct {
	ProjectRepository repository.ProjectRepository
	CodeConverter     externalAPI.CodeConverter
}

type GetProjectListResponseBody struct {
	Projects   []GetProjectListResponseProjectBody `json:"projects"`
	Pagination util.Pagination                     `json:"pagination"`
}

type GetProjectListResponseProjectBody struct {
	ProjectNo   int       `json:"projectNo"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	LastModify  time.Time `json:"lastModify"`
}

func (h *ProjectHandler) GetProjectListHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorw("failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"))
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// query params
	var (
		sortOrder    repository.ProjectSortOrder
		filterType   repository.ProjectFilterType
		filterString string
	)

	switch r.URL.Query().Get("sort") {
	case "createTimeAsc":
		sortOrder = repository.OrderByCreateTimeAsc
	case "createTimeDesc":
		sortOrder = repository.OrderByCreateTimeDesc
	case "updateTimeAsc":
		sortOrder = repository.OrderByUpdateTimeAsc
	case "updateTimeDesc":
		sortOrder = repository.OrderByUpdateTimeDesc
	default:
		sortOrder = repository.OrderByCreateTimeAsc
	}

	switch r.URL.Query().Get("filterType") {
	case "name":
		filterType = repository.FilterByName
	case "nameLike":
		filterType = repository.FilterByNameLike
	case "description":
		filterType = repository.FilterByDescription
	case "descriptionLike":
		filterType = repository.FilterByDescriptionLike
	case "nameOrDescription":
		filterType = repository.FilterByNameOrDescription
	case "nameOrDescriptionLike":
		filterType = repository.FilterByNameOrDescriptionLike
	default:
		filterType = repository.FilterByNone
	}

	filterString = r.URL.Query().Get("filterString")

	count, err := h.ProjectRepository.SelectProjectCount(repository.ClassifiedByUserId(userId), repository.OrderBy(sortOrder), repository.WithFilter(filterType, filterString))
	if err != nil {
		log.Errorw("failed to select project count",
			"error code", util.ErrInternalServerError,
			"error", err,
			"userId", userId)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	pagination := util.NewPaginationFromRequest(r, int64(count))

	projectList, err := h.ProjectRepository.SelectProjectList(repository.ClassifiedByUserId(userId), pagination.Offset(), pagination.Limit(), repository.OrderBy(sortOrder), repository.WithFilter(filterType, filterString))
	if err != nil {
		log.Errorw("failed to select project list",
			"error code", util.ErrInternalServerError,
			"error", err,
			"userId", userId,
			"offset", pagination.Offset(),
			"limit", pagination.Limit())
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
	}

	resp := GetProjectListResponseBody{
		Projects:   make([]GetProjectListResponseProjectBody, 0, len(projectList)),
		Pagination: pagination,
	}
	for _, project := range projectList {
		resp.Projects = append(resp.Projects, GetProjectListResponseProjectBody{
			ProjectNo:   project.ProjectNo,
			Name:        project.Name,
			Description: project.Description,
			LastModify:  project.UpdateTime,
		})
	}

	util.WriteJson(w, http.StatusOK, resp)
}

func (h *ProjectHandler) GetProjectHandler(w http.ResponseWriter, r *http.Request) {
	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		log.Warnw("failed to convert projectNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"])
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
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

	project, err := h.ProjectRepository.SelectProject(repository.ClassifiedByProjectNo(userId, projectNo))
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnw("result of select project is empty",
				"error code", util.ErrNotFound,
				"error", err,
				"userId", userId,
				"projectNo", projectNo)
			util.WriteError(w, http.StatusNotFound, util.ErrNotFound)
			return
		}

		log.Errorw("failed to select project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"userId", userId,
			"projectNo", projectNo)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	util.WriteJson(w, http.StatusOK, util.ResponseBody{
		"projectNo":   project.ProjectNo,
		"name":        project.Name,
		"description": project.Description,
		"lastModify":  project.UpdateTime,
		"content":     project.Content.Json,
		"config":      project.Config.Json,
	})
}

func (h *ProjectHandler) GetProjectContentHandler(w http.ResponseWriter, r *http.Request) {
	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		log.Warnw("failed to convert projectNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"])
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
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

	project, err := h.ProjectRepository.SelectProject(repository.ClassifiedByProjectNo(userId, projectNo))
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnw("result of select project is empty",
				"error code", util.ErrNotFound,
				"error", err,
				"userId", userId,
				"projectNo", projectNo)
			util.WriteError(w, http.StatusNotFound, util.ErrNotFound)
			return
		}

		log.Errorw("failed to select project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"userId", userId,
			"projectNo", projectNo)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	util.WriteJson(w, http.StatusOK, project.Content.Json)
}

func (h *ProjectHandler) GetProjectConfigHandler(w http.ResponseWriter, r *http.Request) {
	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		log.Warnw("failed to convert projectNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"])
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
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

	project, err := h.ProjectRepository.SelectProject(repository.ClassifiedByProjectNo(userId, projectNo))
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnw("result of select project is empty",
				"error code", util.ErrNotFound,
				"error", err,
				"userId", userId,
				"projectNo", projectNo)
			util.WriteError(w, http.StatusNotFound, util.ErrNotFound)
			return
		}

		log.Errorw("failed to select project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"userId", userId,
			"projectNo", projectNo)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	util.WriteJson(w, http.StatusOK, project.Config.Json)
}

type CreateProjectRequestBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c CreateProjectRequestBody) Validate() error {
	if err := checkProjectNameLength(c.Name); err != nil {
		return err
	}

	if err := checkProjectDescriptionLength(c.Description); err != nil {
		return err
	}

	return nil
}

const (
	_maximumProjectNameLength        = 45
	_maximumProjectDescriptionLength = 2000
)

func checkProjectNameLength(projectName string) error {
	if utf8.RuneCountInString(projectName) > _maximumProjectNameLength {
		return errors.New("project name too long")
	}
	return nil
}

func checkProjectDescriptionLength(projectDescription string) error {
	if utf8.RuneCountInString(projectDescription) > _maximumProjectDescriptionLength {
		return errors.New("project description too long")
	}
	return nil
}

type CreateProjectResponseBody struct {
	ProjectNo int `json:"projectNo"`
}

func (h *ProjectHandler) CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	reqBody := CreateProjectRequestBody{}
	if err := util.BindJson(r.Body, &reqBody); err != nil {
		log.Warnw("failed to bind request body to json",
			"error code", util.ErrInvalidRequestBody,
			"error", err)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidRequestBody)
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

	// check project name duplicate
	if _, err := h.ProjectRepository.SelectProject(repository.ClassifiedByProjectName(userId, reqBody.Name)); err != sql.ErrNoRows {
		if err != nil {
			log.Errorw("failed to select project with name",
				"error code", util.ErrInternalServerError,
				"error", err,
				"userId", userId,
				"projectName", reqBody.Name)
			util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
			return
		}

		log.Debugw("failed to insert new project (duplicated)",
			"error code", util.ErrDuplicate,
			"error", err,
			"projectName", reqBody.Name)
		util.WriteError(w, http.StatusUnprocessableEntity, util.ErrDuplicate)
		return
	}

	// get exist project count
	itemCount, err := h.ProjectRepository.SelectProjectCount(repository.ClassifiedByUserId(userId), repository.WithStatus(util.StatusNONE))
	if err != nil {
		log.Errorw("failed to select project count",
			"error code", util.ErrInternalServerError,
			"error", err,
			"userId", userId)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// set new project number
	newProjectNo := itemCount + 1

	// create new project and save to database
	project := model.NewProject(userId, newProjectNo, reqBody.Name, reqBody.Description)
	if _, err := h.ProjectRepository.Insert(project); err != nil {
		log.Errorw("failed to insert new project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"project", project)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	util.WriteJson(w, http.StatusCreated, CreateProjectResponseBody{newProjectNo})
}

type UpdateProjectInfoRequestBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (u UpdateProjectInfoRequestBody) Validate() error {
	if err := checkProjectNameLength(u.Name); err != nil {
		return err
	}

	if err := checkProjectDescriptionLength(u.Description); err != nil {
		return err
	}

	return nil
}

// UpdateProjectInfoHandler update project name, description
func (h *ProjectHandler) UpdateProjectInfoHandler(w http.ResponseWriter, r *http.Request) {
	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		log.Warnw("failed to convert projectNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"])
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
		return
	}

	reqBody := UpdateProjectInfoRequestBody{}
	if err := util.BindJson(r.Body, &reqBody); err != nil {
		log.Warnw("failed to bind request body to json",
			"error code", util.ErrInvalidRequestBody,
			"error", err)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidRequestBody)
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

	// get project
	project, err := h.ProjectRepository.SelectProject(repository.ClassifiedByProjectNo(userId, projectNo))
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnw("result of select project is empty",
				"error code", util.ErrNotFound,
				"error", err,
				"userId", userId,
				"projectNo", projectNo)
			util.WriteError(w, http.StatusNotFound, util.ErrNotFound)
			return
		}

		log.Errorw("failed to select project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"userId", userId,
			"projectNo", projectNo)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// check project name duplicate
	if _, err := h.ProjectRepository.SelectProject(repository.ClassifiedByProjectName(userId, reqBody.Name), repository.WithExcludeProjectId(project.Id)); err != sql.ErrNoRows {
		if err != nil {
			log.Errorw("failed to select project with name",
				"error code", util.ErrInternalServerError,
				"error", err,
				"userId", userId,
				"projectName", reqBody.Name)
			util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
			return
		}

		log.Debugw("failed to insert new project (duplicated)",
			"error code", util.ErrDuplicate,
			"error", err,
			"projectName", reqBody.Name)
		util.WriteError(w, http.StatusUnprocessableEntity, util.ErrDuplicate)
		return
	}

	// update project
	project.Name = reqBody.Name
	project.Description = reqBody.Description
	if err := h.ProjectRepository.Update(project); err != nil {
		log.Errorw("failed to update project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"project", project)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateProjectContentHandler update project content
func (h *ProjectHandler) UpdateProjectContentHandler(w http.ResponseWriter, r *http.Request) {
	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		log.Warnw("failed to convert projectNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"])
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
		return
	}

	// check request body json Unmarshalable
	reqBodyUnmarshaled := make(map[string]interface{})
	if err := json.NewDecoder(r.Body).Decode(&reqBodyUnmarshaled); err != nil {
		log.Warnw("failed to json decode request body",
			"error code", util.ErrInvalidRequestBody,
			"error", err)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidRequestBody)
		return
	}

	reqBodyBytes, err := json.Marshal(reqBodyUnmarshaled)
	if err != nil {
		log.Errorw("failed to json marshal reqBodyUnmarshaled",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
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

	// get project
	project, err := h.ProjectRepository.SelectProject(repository.ClassifiedByProjectNo(userId, projectNo))
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnw("result of select project is empty",
				"error code", util.ErrNotFound,
				"error", err,
				"userId", userId,
				"projectNo", projectNo)
			util.WriteError(w, http.StatusNotFound, util.ErrNotFound)
			return
		}

		log.Errorw("failed to select project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"userId", userId,
			"projectNo", projectNo)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// update project
	project.Content.Json = reqBodyBytes
	if err := h.ProjectRepository.Update(project); err != nil {
		log.Errorw("failed to update project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"project", project)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateProjectConfigHandler update project config
func (h *ProjectHandler) UpdateProjectConfigHandler(w http.ResponseWriter, r *http.Request) {
	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		log.Warnw("failed to convert projectNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"])
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
		return
	}

	// check request body json Unmarshalable
	reqBodyUnmarshaled := make(map[string]interface{})
	if err := json.NewDecoder(r.Body).Decode(&reqBodyUnmarshaled); err != nil {
		log.Warnw("failed to json decode request body",
			"error code", util.ErrInvalidRequestBody,
			"error", err)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidRequestBody)
		return
	}

	reqBodyBytes, err := json.Marshal(reqBodyUnmarshaled)
	if err != nil {
		log.Errorw("failed to json marshal reqBodyUnmarshaled",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
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

	// get project
	project, err := h.ProjectRepository.SelectProject(repository.ClassifiedByProjectNo(userId, projectNo))
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnw("result of select project is empty",
				"error code", util.ErrNotFound,
				"error", err,
				"userId", userId,
				"projectNo", projectNo)
			util.WriteError(w, http.StatusNotFound, util.ErrNotFound)
			return
		}

		log.Errorw("failed to select project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"userId", userId,
			"projectNo", projectNo)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// update project
	project.Config.Json = reqBodyBytes
	if err := h.ProjectRepository.Update(project); err != nil {
		log.Errorw("failed to update project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"project", project)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		log.Warnw("failed to convert projectNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"])
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
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

	// get project
	project, err := h.ProjectRepository.SelectProject(repository.ClassifiedByProjectNo(userId, projectNo))
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnw("result of select project is empty",
				"error code", util.ErrNotFound,
				"error", err,
				"userId", userId,
				"projectNo", projectNo)
			util.WriteError(w, http.StatusNotFound, util.ErrNotFound)
			return
		}

		log.Errorw("failed to select project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"userId", userId,
			"projectNo", projectNo)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	if err := h.ProjectRepository.Delete(project); err != nil {
		log.Errorw("failed to delete project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"project", project)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) GetPythonCodeHandler(w http.ResponseWriter, r *http.Request) {
	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		log.Warnw("failed to convert projectNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"])
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
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

	project, err := h.ProjectRepository.SelectProject(repository.ClassifiedByProjectNo(userId, projectNo))
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnw("result of select project is empty",
				"error code", util.ErrNotFound,
				"error", err,
				"userId", userId,
				"projectNo", projectNo)
			util.WriteError(w, http.StatusNotFound, util.ErrNotFound)
			return
		}

		log.Errorw("failed to select project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"userId", userId,
			"projectNo", projectNo)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// make request body
	payload := externalAPI.CodeConvertRequestBody{
		Content: project.Content.Json,
		Config:  project.Config.Json,
	}

	// send request
	resp, err := h.CodeConverter.CodeConvert(payload)
	if err != nil {
		log.Errorw("failed to generate python code",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}
	defer resp.Body.Close()

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("failed to copy file: %v", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}
	
	util.WriteJson(w, http.StatusOK, util.ResponseBody{"code":string(responseBytes)})
}

func (h *ProjectHandler) GenerateShareKeyHandler(w http.ResponseWriter, r *http.Request) {
	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		log.Warnw("failed to convert projectNo to int",
			"error code", util.ErrInvalidPathParm,
			"error", err,
			"input value", mux.Vars(r)["projectNo"])
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidPathParm)
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

	project, err := h.ProjectRepository.SelectProject(repository.ClassifiedByProjectNo(userId, projectNo))
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnw("result of select project is empty",
				"error code", util.ErrNotFound,
				"error", err,
				"userId", userId,
				"projectNo", projectNo)
			util.WriteError(w, http.StatusNotFound, util.ErrNotFound)
			return
		}

		log.Errorw("failed to select project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"userId", userId,
			"projectNo", projectNo)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	project.ShareKey = sql.NullString{
		String: uuid.New().String(),
		Valid:  true,
	}
	if err := h.ProjectRepository.Update(project); err != nil {
		log.Errorw("failed to update project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"userId", userId,
			"projectNo", projectNo)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	util.WriteJson(w, http.StatusOK, util.ResponseBody{"key": project.ShareKey.String})
}
