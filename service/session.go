package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"nns_back/log"
	"nns_back/model"
	"nns_back/util"
)

type Auth struct {
	Env

	SessionStore sessions.Store
}

const _sessionCookieName = "NNS"

func SetSessionStore(key []byte) sessions.Store {
	return sessions.NewCookieStore(key)
}

type LoginHandlerRequestBody struct {
	ID string `json:"id"`
	PW string `json:"pw"`
}

// TODO: Support social login
func (a Auth) LoginHandler(w http.ResponseWriter, r *http.Request) {
	reqBody := LoginHandlerRequestBody{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Warnw("failed to json decode request body",
			"error code", util.ErrInvalidRequestBody,
			"error", err)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidRequestBody)
		return
	}

	user, err := model.SelectUser(a.DB, model.ClassifiedByLoginId(reqBody.ID))
	if err != nil {
		if err == sql.ErrNoRows {
			// invalid login id
			log.Debugw("ID does not exist",
				"error code", util.ErrInvalidAuthentication,
				"request login ID", reqBody.ID)
			util.WriteError(w, http.StatusUnauthorized, util.ErrInvalidAuthentication)
			return
		}

		log.Errorw("failed to select user",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// bcrypt.CompareHashAndPassword returns nil on success, or an ErrMismatchedHashAndPassword on failure.
	if err := bcrypt.CompareHashAndPassword(user.LoginPw.Bytes, []byte(reqBody.PW)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			// password mismatch
			log.Debugw("password mismatch",
				"error code", util.ErrInvalidAuthentication,
				"loginId", user.LoginId,
				"loginPw", user.LoginPw)
			util.WriteError(w, http.StatusUnauthorized, util.ErrInvalidAuthentication)
			return
		}

		// error
		log.Errorw("failed to compare password",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	session, err := a.SessionStore.Get(r, _sessionCookieName)
	if err != nil {
		log.Errorw("failed to session store get",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	session.Values["authenticated"] = true
	session.Values["userId"] = user.Id
	session.Options.SameSite = http.SameSiteNoneMode
	session.Options.Secure = true
	session.Options.HttpOnly = true
	if err := session.Save(r, w); err != nil {
		log.Errorw("failed to save session",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a Auth) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := a.SessionStore.Get(r, _sessionCookieName)
	if err != nil {
		log.Errorw("failed to session store get",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	session.Values["authenticated"] = false
	session.Options.SameSite = http.SameSiteNoneMode
	session.Options.Secure = true
	session.Options.HttpOnly = true
	if err := session.Save(r, w); err != nil {
		log.Errorw("failed to save session",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a Auth) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := a.SessionStore.Get(r, _sessionCookieName)
		if err != nil {
			log.Errorw("failed to session store get",
				"error code", util.ErrInternalServerError,
				"error", err)
			util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
			return
		}
		log.Debug(session.Values)
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			util.WriteError(w, http.StatusUnauthorized, util.ErrLoginRequired)
			return
		}

		userId := session.Values["userId"].(int64)
		r = r.WithContext(context.WithValue(r.Context(), "userId", userId))
		next.ServeHTTP(w, r)
	})
}