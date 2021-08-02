package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"nns_back/model"
	"regexp"
	"time"
	"unicode"
	"unicode/utf8"
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
)

// TODO: move to environment variable
const defaultUserProfileImage = "https://s3.ap-northeast-2.amazonaws.com/image.nns/default_profile.png"

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
	if err := validatePassword(reqBody.PW); err != nil {
		e.Logger.Debug(reqBody.PW)
		writeError(w, http.StatusBadRequest, ErrInvalidFormat, col(target, "password"))
		return
	}

	// check ID duplication
	if _, err := model.SelectUser(e.DB, model.ClassifiedByLoginId(reqBody.ID)); err != sql.ErrNoRows {
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

func validatePassword(password string) error {
	length := utf8.RuneCountInString(password)
	if length < _pwMinLen {
		return fmt.Errorf("password must be at least %d characters", _pwMinLen)
	}
	if length > _pwMaxLen {
		return fmt.Errorf("password must be %d characters or less", _pwMaxLen)
	}

next:
	for name, classes := range map[string][]*unicode.RangeTable{
		"alphabet": {unicode.Upper, unicode.Lower, unicode.Title},
		"numeric":  {unicode.Number, unicode.Digit},
		"special":  {unicode.Space, unicode.Symbol, unicode.Punct, unicode.Mark},
	} {
		for _, character := range password {
			if unicode.IsOneOf(classes, character) {
				continue next
			}
		}
		return fmt.Errorf("password must have at least one %s character", name)
	}
	return nil
}

type GetUserHandlerResponseBody struct {
	Name         string    `json:"name"`
	ProfileImage string    `json:"profileImage"`
	Description  string    `json:"description"`
	Email        string    `json:"email"`
	WebSite      string    `json:"webSite"`
	CreateTime   time.Time `json:"createTime"`
	UpdateTime   time.Time `json:"updateTime"`
}

// TODO: Email verification, WebSite url syntax validation
func (e Env) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		e.Logger.Errorw("failed to conversion interface to int64",
			"error code", ErrInternalServerError,
			"context value", r.Context().Value("userId"))
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	user, err := model.SelectUser(e.DB, model.ClassifiedById(userId))
	if err != nil {
		e.Logger.Errorw("failed to select user",
			"error code", ErrInternalServerError,
			"error", err,
			"userId", userId)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	profileImage := defaultUserProfileImage
	if user.ProfileImage.Valid {
		image, err := model.SelectImage(e.DB, userId, user.ProfileImage.Int64)
		if err != nil {
			e.Logger.Errorw("failed to select image",
				"error code", ErrInternalServerError,
				"error", err)
			writeError(w, http.StatusInternalServerError, ErrInternalServerError)
			return
		}

		profileImage = image.Url
	}

	resp := GetUserHandlerResponseBody{
		Name:         user.Name,
		ProfileImage: profileImage,
		Description:  user.Description.String,
		Email:        user.Email.String,
		WebSite:      user.WebSite.String,
		CreateTime:   user.CreateTime,
		UpdateTime:   user.UpdateTime,
	}
	writeJson(w, http.StatusOK, resp)
}

type UpdateUserHandlerRequestBody struct {
	ProfileImage int64  `json:"profileImage"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Email        string `json:"email"`
	WebSite      string `json:"webSite"`
}

func (e Env) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		e.Logger.Errorw("failed to conversion interface to int64",
			"error code", ErrInternalServerError,
			"context value", r.Context().Value("userId"))
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	reqBody := UpdateUserHandlerRequestBody{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		e.Logger.Warnw("failed to json decode request body",
			"error code", ErrInvalidRequestBody,
			"error", err)
		writeError(w, http.StatusBadRequest, ErrInvalidRequestBody)
		return
	}

	user, err := model.SelectUser(e.DB, model.ClassifiedById(userId))
	if err != nil {
		e.Logger.Errorw("failed to select user",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	user.Name = reqBody.Name
	user.Description = sql.NullString{
		String: reqBody.Description,
		Valid:  reqBody.Description != "",
	}
	user.Email = sql.NullString{
		String: reqBody.Email,
		Valid:  reqBody.Email != "",
	}
	user.WebSite = sql.NullString{
		String: reqBody.WebSite,
		Valid:  reqBody.WebSite != "",
	}

	if reqBody.ProfileImage != 0 {
		if _, err := model.SelectImage(e.DB, userId, reqBody.ProfileImage); err != nil {
			if err == sql.ErrNoRows {
				e.Logger.Warnw("invalid image id",
					"request image ID", reqBody.ProfileImage,
					"request user ID", userId)
				writeError(w, http.StatusBadRequest, ErrInvalidImageId)
				return
			}

			e.Logger.Errorw("failed to select image",
				"error code", ErrInternalServerError,
				"error", err)
			writeError(w, http.StatusInternalServerError, ErrInternalServerError)
			return
		}
	}

	user.ProfileImage = sql.NullInt64{
		Int64: reqBody.ProfileImage,
		Valid: reqBody.ProfileImage != 0,
	}

	if err := user.Update(e.DB); err != nil {
		e.Logger.Errorw("failed to update user",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type UpdateUserPasswordHandlerRequestBody struct {
	PW string `json:"pw"`
}

func (e Env) UpdateUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		e.Logger.Errorw("failed to conversion interface to int64",
			"error code", ErrInternalServerError,
			"context value", r.Context().Value("userId"))
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	reqBody := UpdateUserPasswordHandlerRequestBody{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		e.Logger.Warnw("failed to json decode request body",
			"error code", ErrInvalidRequestBody,
			"error", err)
		writeError(w, http.StatusBadRequest, ErrInvalidRequestBody)
		return
	}

	// check PW validation
	if err := validatePassword(reqBody.PW); err != nil {
		e.Logger.Debug(reqBody.PW)
		writeError(w, http.StatusBadRequest, ErrInvalidFormat, col(target, "password"))
		return
	}

	// hash password
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

	user, err := model.SelectUser(e.DB, model.ClassifiedById(userId))
	if err != nil {
		e.Logger.Errorw("failed to select user",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	user.LoginPw = model.NullBytes{
		Bytes: hashedPassword,
		Valid: true,
	}

	if err := user.Update(e.DB); err != nil {
		e.Logger.Errorw("failed to update user",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a Auth) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		a.Logger.Errorw("failed to conversion interface to int64",
			"error code", ErrInternalServerError,
			"context value", r.Context().Value("userId"))
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	user, err := model.SelectUser(a.DB, model.ClassifiedById(userId))
	if err != nil {
		a.Logger.Errorw("failed to select user",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	if err := user.Delete(a.DB); err != nil {
		a.Logger.Errorw("failed to delete user",
			"error code", ErrInternalServerError,
			"error", err)
		writeError(w, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	a.LogoutHandler(w, r)
}
