package service

import (
	"database/sql"
	"github.com/gorilla/mux"
	"net/http"
	"nns_back/model"
	"strconv"
	"time"
)

const tempUserId int64 = 1

func (e Env) GetProjectListHandler(w http.ResponseWriter, r *http.Request) {
	// implement require ----------------------------
	userId := tempUserId
	// ----------------------------------------------

	count, err := model.SelectProjectCount(e.DB, userId)
	if err != nil {
		e.Logger.Errorw("failed to select project count",
			"error code", ErrInternalServerError,
			"error", err,
			"userId", userId)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	pagination := GetPaginationFromUrl(r, count)

	projectList, err := model.SelectProjectList(e.DB, userId, pagination.Offset(), pagination.Limit())
	if err != nil {
		e.Logger.Errorw("failed to select project list",
			"error code", ErrInternalServerError,
			"error", err,
			"userId", userId,
			"offset", pagination.Offset(),
			"limit", pagination.Limit())
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
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

	writeJson(w, http.StatusOK, resp)
}

type GetProjectListResponseBody struct {
	Projects   []GetProjectListResponseProjectBody `json:"projects"`
	Pagination Pagination                          `json:"pagination"`
}

type GetProjectListResponseProjectBody struct {
	ProjectNo   int       `json:"projectNo"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	LastModify  time.Time `json:"lastModify"`
}

func (e Env) GetProjectHandler(w http.ResponseWriter, r *http.Request) {
	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		e.Logger.Warnw("failed to convert projectNo to int",
			"error code", ErrNotFound,
			"error", err,
			"input value", mux.Vars(r)["projectNo"])
		writeError(w, http.StatusBadRequest, ErrInvalidPathParm)
		return
	}

	// implement require ----------------------------
	userId := tempUserId
	// ----------------------------------------------

	project, err := model.SelectProject(e.DB, userId, projectNo)
	if err != nil {
		if err == sql.ErrNoRows {
			e.Logger.Warnw("result of select project is empty",
				"error code", ErrNotFound,
				"error", err,
				"userId", userId,
				"projectNo", projectNo)
			writeError(w, http.StatusNotFound, ErrNotFound)
			return
		}

		e.Logger.Errorw("failed to select project",
			"error code", ErrInternalServerError,
			"error", err,
			"userId", userId,
			"projectNo", projectNo)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	writeJson(w, http.StatusOK, responseBody{
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
		e.Logger.Warnw("failed to convert projectNo to int",
			"error code", ErrNotFound,
			"error", err,
			"input value", mux.Vars(r)["projectNo"])
		writeError(w, http.StatusBadRequest, ErrInvalidPathParm)
		return
	}

	// implement require ----------------------------
	userId := tempUserId
	// ----------------------------------------------

	project, err := model.SelectProject(e.DB, userId, projectNo)
	if err != nil {
		if err == sql.ErrNoRows {
			e.Logger.Warnw("result of select project is empty",
				"error code", ErrNotFound,
				"error", err,
				"userId", userId,
				"projectNo", projectNo)
			writeError(w, http.StatusNotFound, ErrNotFound)
			return
		}

		e.Logger.Errorw("failed to select project",
			"error code", ErrInternalServerError,
			"error", err,
			"userId", userId,
			"projectNo", projectNo)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	writeJson(w, http.StatusOK, project.Content.Json)
}

func (e Env) GetProjectConfigHandler(w http.ResponseWriter, r *http.Request) {
	projectNo, err := strconv.Atoi(mux.Vars(r)["projectNo"])
	if err != nil {
		e.Logger.Warnw("failed to convert projectNo to int",
			"error code", ErrNotFound,
			"error", err,
			"input value", mux.Vars(r)["projectNo"])
		writeError(w, http.StatusBadRequest, ErrInvalidPathParm)
		return
	}

	// implement require ----------------------------
	userId := tempUserId
	// ----------------------------------------------

	project, err := model.SelectProject(e.DB, userId, projectNo)
	if err != nil {
		if err == sql.ErrNoRows {
			e.Logger.Warnw("result of select project is empty",
				"error code", ErrNotFound,
				"error", err,
				"userId", userId,
				"projectNo", projectNo)
			writeError(w, http.StatusNotFound, ErrNotFound)
			return
		}

		e.Logger.Errorw("failed to select project",
			"error code", ErrInternalServerError,
			"error", err,
			"userId", userId,
			"projectNo", projectNo)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	writeJson(w, http.StatusOK, project.Config.Json)
}

func (e Env) CreateProjectHandler(w http.ResponseWriter, r *http.Request) {

}

type CreateProjectRequestBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (e Env) UpdateProjectHandler(w http.ResponseWriter, r *http.Request) {

}

func (e Env) DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {

}
