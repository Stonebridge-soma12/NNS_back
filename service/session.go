package service

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"nns_back/model"
	"regexp"
)

func SetSessionStore(key []byte) sessions.Store {
	return sessions.NewCookieStore(key)
}

const (
	_sessionCookieName = "NNS"
)

type SignUpHandlerRequestBody struct {
	ID string `json:"id"`
	PW string `json:"pw"`
}

const (
	_idMinLen = 2
	_idMaxLen = 50
	_pwMinLen = 8
	_pwMaxLen = 72

	_idRegexp = `[a-zA-Z0-9]{2,50}`
	_pwRegexp = `([0-9a-zA-Z^\w\d\s]|_){8,72}`
)

// TODO: email authentication
func (e Env) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	// bind request body
	reqBody := SignUpHandlerRequestBody{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		e.Logger.Warnw("failed to json decode request body",
			"error code", ErrInvalidRequestBody,
			"error", err)
		writeError(w, http.StatusBadRequest, ErrInvalidRequestBody)
		return
	}

	// check ID validation
	match, err := regexp.MatchString(_idRegexp, reqBody.ID)
	if err != nil {
		e.Logger.Errorw("failed to regexp compile",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}
	if !match {
		e.Logger.Debug(reqBody.ID)
		writeError(w, http.StatusBadRequest, ErrInvalidFormat, col(target, "ID"))
		return
	}

	// check PW validation
	match, err = regexp.MatchString(_pwRegexp, reqBody.PW)

	if err != nil {
		e.Logger.Errorw("failed to regexp compile",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}
	if !match {
		e.Logger.Debug(reqBody.PW)
		writeError(w, http.StatusBadRequest, ErrInvalidFormat, col(target, "password"))
		return
	}

	// check ID duplication
	if _, err := model.SelectUser(e.DB, reqBody.ID); err != sql.ErrNoRows {
		if err != nil {
			// error occur
			e.Logger.Errorw("failed to select user",
				"error code", ErrInternalServerError,
				"error", err)
			writeError(w, http.StatusInternalServerError, ErrInternalServerError)
			return
		}

		// duplicate user id
		e.Logger.Debug(reqBody.ID)
		writeError(w, http.StatusUnprocessableEntity, ErrDuplicate)
		return
	}

	// create user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reqBody.PW), bcrypt.DefaultCost)
	if err != nil {
		e.Logger.Errorw("failed to generate hashed password",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}
	e.Logger.Debugw("password hashed finish",
		"hashed password", hashedPassword,
		"hashed password len", len(hashedPassword))

	user := model.NewUser(reqBody.ID, hashedPassword)
	if _, err := user.Insert(e.DB); err != nil {
		e.Logger.Errorw("failed to insert user",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type LoginHandlerRequestBody struct {
	ID string `json:"id"`
	PW string `json:"pw"`
}

// TODO: Support social login
func (e Env) LoginHandler(w http.ResponseWriter, r *http.Request) {
	reqBody := LoginHandlerRequestBody{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		e.Logger.Warnw("failed to json decode request body",
			"error code", ErrInvalidRequestBody,
			"error", err)
		writeError(w, http.StatusBadRequest, ErrInvalidRequestBody)
		return
	}

	user, err := model.SelectUser(e.DB, reqBody.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			// invalid login id
			e.Logger.Debugw("ID does not exist",
				"error code", ErrInvalidAuthentication,
				"request login ID", reqBody.ID)
			writeError(w, http.StatusUnauthorized, ErrInvalidAuthentication)
			return
		}

		e.Logger.Errorw("failed to select user",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	// bcrypt.CompareHashAndPassword returns nil on success, or an ErrMismatchedHashAndPassword on failure.
	if err := bcrypt.CompareHashAndPassword(user.LoginPw.Bytes, []byte(reqBody.PW)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			// password mismatch
			e.Logger.Debugw("password mismatch",
				"error code", ErrInvalidAuthentication,
				"loginId", user.LoginId,
				"loginPw", user.LoginPw)
			writeError(w, http.StatusUnauthorized, ErrInvalidAuthentication)
			return
		}

		// error
		e.Logger.Errorw("failed to compare password",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	session, err := e.SessionStore.Get(r, _sessionCookieName)
	if err != nil {
		e.Logger.Errorw("failed to session store get",
			"error", err)
	}

	session.Values["authenticated"] = true
	session.Values["userId"] = user.Id
	if err := session.Save(r, w); err != nil {
		e.Logger.Errorw("failed to save session",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}
}

func (e Env) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := e.SessionStore.Get(r, _sessionCookieName)
	if err != nil {
		e.Logger.Errorw("failed to session store get",
			"error", err)
	}

	session.Values["authenticated"] = false
	session.Save(r, w)
}

func (e Env) Secret(w http.ResponseWriter, r *http.Request) {
	session, err := e.SessionStore.Get(r, _sessionCookieName)
	if err != nil {
		e.Logger.Errorw("failed to session store get",
			"error", err)
	}

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	userId := session.Values["userId"].(int64)
	writeJson(w, http.StatusOK, responseBody{
		"userID": userId,
	})
}
