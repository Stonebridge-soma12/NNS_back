package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"nns_back/model"
	"regexp"
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
