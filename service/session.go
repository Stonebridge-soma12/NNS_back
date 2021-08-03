package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"nns_back/model"
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
		a.Logger.Warnw("failed to json decode request body",
			"error code", ErrInvalidRequestBody,
			"error", err)
		writeError(w, http.StatusBadRequest, ErrInvalidRequestBody)
		return
	}

	user, err := model.SelectUser(a.DB, model.ClassifiedByLoginId(reqBody.ID))
	if err != nil {
		if err == sql.ErrNoRows {
			// invalid login id
			a.Logger.Debugw("ID does not exist",
				"error code", ErrInvalidAuthentication,
				"request login ID", reqBody.ID)
			writeError(w, http.StatusUnauthorized, ErrInvalidAuthentication)
			return
		}

		a.Logger.Errorw("failed to select user",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	// bcrypt.CompareHashAndPassword returns nil on success, or an ErrMismatchedHashAndPassword on failure.
	if err := bcrypt.CompareHashAndPassword(user.LoginPw.Bytes, []byte(reqBody.PW)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			// password mismatch
			a.Logger.Debugw("password mismatch",
				"error code", ErrInvalidAuthentication,
				"loginId", user.LoginId,
				"loginPw", user.LoginPw)
			writeError(w, http.StatusUnauthorized, ErrInvalidAuthentication)
			return
		}

		// error
		a.Logger.Errorw("failed to compare password",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	session, err := a.SessionStore.Get(r, _sessionCookieName)
	if err != nil {
		a.Logger.Errorw("failed to session store get",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	session.Values["authenticated"] = true
	session.Values["userId"] = user.Id
	session.Options.SameSite = http.SameSiteNoneMode
	session.Options.Secure = true
	session.Options.HttpOnly = true
	if err := session.Save(r, w); err != nil {
		a.Logger.Errorw("failed to save session",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a Auth) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := a.SessionStore.Get(r, _sessionCookieName)
	if err != nil {
		a.Logger.Errorw("failed to session store get",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	session.Values["authenticated"] = false
	session.Options.SameSite = http.SameSiteNoneMode
	session.Options.Secure = true
	session.Options.HttpOnly = true
	if err := session.Save(r, w); err != nil {
		a.Logger.Errorw("failed to save session",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a Auth) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := a.SessionStore.Get(r, _sessionCookieName)
		if err != nil {
			a.Logger.Errorw("failed to session store get",
				"error code", ErrInternalServerError,
				"error", err)
			writeError(w, http.StatusInternalServerError, ErrInternalServerError)
			return
		}
		a.Logger.Debug(session.Values)
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			writeError(w, http.StatusUnauthorized, ErrLoginRequired)
			return
		}

		userId := session.Values["userId"].(int64)
		r = r.WithContext(context.WithValue(r.Context(), "userId", userId))
		next.ServeHTTP(w, r)
	})
}