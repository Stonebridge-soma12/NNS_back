package service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"nns_back/log"
	"nns_back/model"
	"nns_back/util"
	"strconv"
	"time"
	"unicode/utf8"
)

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

func (e Env) GetProjectListHandler(w http.ResponseWriter, r *http.Request) {
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
		sortOrder    model.ProjectSortOrder
		filterType   model.ProjectFilterType
		filterString string
	)

	switch r.URL.Query().Get("sort") {
	case "createTimeAsc":
		sortOrder = model.OrderByCreateTimeAsc
	case "createTimeDesc":
		sortOrder = model.OrderByCreateTimeDesc
	case "updateTimeAsc":
		sortOrder = model.OrderByUpdateTimeAsc
	case "updateTimeDesc":
		sortOrder = model.OrderByUpdateTimeDesc
	default:
		sortOrder = model.OrderByCreateTimeAsc
	}

	switch r.URL.Query().Get("filterType") {
	case "name":
		filterType = model.FilterByName
	case "nameLike":
		filterType = model.FilterByNameLike
	case "description":
		filterType = model.FilterByDescription
	case "descriptionLike":
		filterType = model.FilterByDescriptionLike
	case "nameOrDescription":
		filterType = model.FilterByNameOrDescription
	case "nameOrDescriptionLike":
		filterType = model.FilterByNameOrDescriptionLike
	default:
		filterType = model.FilterByNone
	}

	filterString = r.URL.Query().Get("filterString")

	count, err := model.SelectProjectCount(e.DB, model.ClassifiedByUserId(userId),
		model.OrderBy(sortOrder),
		model.WithFilter(filterType, filterString))
	if err != nil {
		log.Errorw("failed to select project count",
			"error code", util.ErrInternalServerError,
			"error", err,
			"userId", userId)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	pagination := util.NewPaginationFromRequest(r, int64(count))

	projectList, err := model.SelectProjectList(e.DB, model.ClassifiedByUserId(userId), pagination.Offset(), pagination.Limit(),
		model.OrderBy(sortOrder),
		model.WithFilter(filterType, filterString))
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

func (e Env) GetProjectHandler(w http.ResponseWriter, r *http.Request) {
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

	project, err := model.SelectProject(e.DB, model.ClassifiedByProjectNo(userId, projectNo))
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

func (e Env) GetProjectContentHandler(w http.ResponseWriter, r *http.Request) {
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

	project, err := model.SelectProject(e.DB, model.ClassifiedByProjectNo(userId, projectNo))
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

func (e Env) GetProjectConfigHandler(w http.ResponseWriter, r *http.Request) {
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

	project, err := model.SelectProject(e.DB, model.ClassifiedByProjectNo(userId, projectNo))
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

func (e Env) CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
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
	if _, err := model.SelectProject(e.DB, model.ClassifiedByProjectName(userId, reqBody.Name)); err != sql.ErrNoRows {
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
	itemCount, err := model.SelectProjectCount(e.DB, model.ClassifiedByUserId(userId), model.WithStatus(util.StatusNONE))
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
	if _, err := project.Insert(e.DB); err != nil {
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
func (e Env) UpdateProjectInfoHandler(w http.ResponseWriter, r *http.Request) {
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
	project, err := model.SelectProject(e.DB, model.ClassifiedByProjectNo(userId, projectNo))
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
	if _, err := model.SelectProject(e.DB, model.ClassifiedByProjectName(userId, reqBody.Name), model.WithExcludeProjectId(project.Id)); err != sql.ErrNoRows {
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
	if err := project.Update(e.DB); err != nil {
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
func (e Env) UpdateProjectContentHandler(w http.ResponseWriter, r *http.Request) {
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
	project, err := model.SelectProject(e.DB, model.ClassifiedByProjectNo(userId, projectNo))
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
	if err := project.Update(e.DB); err != nil {
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
func (e Env) UpdateProjectConfigHandler(w http.ResponseWriter, r *http.Request) {
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
	project, err := model.SelectProject(e.DB, model.ClassifiedByProjectNo(userId, projectNo))
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
	if err := project.Update(e.DB); err != nil {
		log.Errorw("failed to update project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"project", project)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (e Env) DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
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
	project, err := model.SelectProject(e.DB, model.ClassifiedByProjectNo(userId, projectNo))
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

	if err := project.Delete(e.DB); err != nil {
		log.Errorw("failed to delete project",
			"error code", util.ErrInternalServerError,
			"error", err,
			"project", project)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (e Env) GetPythonCodeHandler(w http.ResponseWriter, r *http.Request) {
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

	project, err := model.SelectProject(e.DB, model.ClassifiedByProjectNo(userId, projectNo))
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

	// http client
	defaultTransportPointer, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		log.Errorw("failed to interface conversion",
			"error code", util.ErrInternalServerError,
			"msg", "defaultRoundTripper not an *http.Transport",
		)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}
	defaultTransport := *defaultTransportPointer // dereference it to get a copy of the struct that the pointer points to
	defaultTransport.MaxIdleConns = 100
	defaultTransport.MaxIdleConnsPerHost = 100
	client := &http.Client{Transport: &defaultTransport}

	// make request body
	payload := make(map[string]interface{})
	payload["content"] = project.Content.Json
	payload["config"] = project.Config.Json
	jsonedPayload, err := json.Marshal(payload)
	if err != nil {
		log.Errorw("failed to json marshal",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// send request
	resp, err := client.Post("http://54.180.153.56:8080/make-python", "application/json", bytes.NewBuffer(jsonedPayload))
	if err != nil {
		log.Errorw("failed to generate python code",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}
	defer resp.Body.Close()

	// response
	w.Header().Set("Content-Disposition", "attachment; filename=model.py")
	w.Header().Set("Content-Type", "text/x-python; charset=utf-8")
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Errorw("failed to copy file",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}
}

func (e Env) GenerateShareKeyHandler(w http.ResponseWriter, r *http.Request) {
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

	project, err := model.SelectProject(e.DB, model.ClassifiedByProjectNo(userId, projectNo))
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
	if err := project.Update(e.DB); err != nil {
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
