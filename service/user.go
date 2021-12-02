package service

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"nns_back/dataset"
	"nns_back/datasetConfig"
	"nns_back/log"
	"nns_back/model"
	"nns_back/repository"
	"nns_back/util"
	"regexp"
	"time"
	"unicode"
	"unicode/utf8"
)

type userHandler struct {
	UserRepository          repository.UserRepository
	ImageRepository         repository.ImageRepository
	ProjectRepository       repository.ProjectRepository
	DatasetRepository       dataset.Repository
	DatasetConfigRepository datasetConfig.Repository
	SessionService          SessionService
}

func NewUserHandler(
	userRepository repository.UserRepository,
	imageRepository repository.ImageRepository,
	projectRepository repository.ProjectRepository,
	datasetRepository dataset.Repository,
	datasetConfigRepository datasetConfig.Repository,
	sessionService SessionService) *userHandler {
	return &userHandler{
		UserRepository:          userRepository,
		ImageRepository:         imageRepository,
		ProjectRepository:       projectRepository,
		DatasetRepository:       datasetRepository,
		DatasetConfigRepository: datasetConfigRepository,
		SessionService:          sessionService,
	}
}

type SignUpHandlerRequestBody struct {
	ID string `json:"id"`
	PW string `json:"pw"`
}

func (s SignUpHandlerRequestBody) Validate() error {
	// check ID validation
	if err := validateID(s.ID); err != nil {
		return errors.Wrap(err, "invalid ID")
	}

	// check PW validation
	if err := validatePassword(s.PW); err != nil {
		return errors.Wrap(err, "invalid password")
	}

	return nil
}

const (
	_idMinLen = 2
	_idMaxLen = 50
	_pwMinLen = 8
	_pwMaxLen = 72

	_idRegexp = `[a-zA-Z0-9]{2,50}`
)

func validateID(id string) error {
	match, err := regexp.MatchString(_idRegexp, id)
	if err != nil {
		return err
	}
	if !match {
		return errors.New("ID regexp doesn't match")
	}
	return nil
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

// TODO: move to environment variable
const defaultUserProfileImage = "https://s3.ap-northeast-2.amazonaws.com/image.nns/default_profile.png"

// TODO: email authentication
func (h *userHandler) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	// bind request body
	reqBody := SignUpHandlerRequestBody{}
	if err := util.BindJson(r.Body, &reqBody); err != nil {
		log.Warnw("failed to bind request body to json",
			"error code", util.ErrInvalidRequestBody,
			"error", err)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidRequestBody)
		return
	}

	// check ID duplication
	if _, err := h.UserRepository.SelectUser(repository.ClassifiedByLoginId(reqBody.ID)); err != sql.ErrNoRows {
		if err != nil {
			// error occur
			log.Errorw("failed to select user",
				"error code", util.ErrInternalServerError,
				"error", err)
			util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
			return
		}

		// duplicate user id
		log.Debug(reqBody.ID)
		util.WriteError(w, http.StatusUnprocessableEntity, util.ErrDuplicate)
		return
	}

	// create user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reqBody.PW), bcrypt.DefaultCost)
	if err != nil {
		log.Errorw("failed to generate hashed password",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}
	log.Debugw("password hashed finish",
		"hashed password", hashedPassword,
		"hashed password len", len(hashedPassword))

	user := model.NewUser(reqBody.ID, hashedPassword)
	user.Id, err = h.UserRepository.Insert(user)
	if err != nil {
		log.Errorw("failed to insert user",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	// create sample project (MNIST)
	if err := CreateAutoCreatedSampleProject(user.Id, h.ProjectRepository, h.DatasetRepository, h.DatasetConfigRepository); err != nil {
		log.Errorw("failed to create sample project",
			"error", err)
	}

	w.WriteHeader(http.StatusCreated)
}

type GetUserHandlerResponseBody struct {
	Name         string `json:"name"`
	ProfileImage struct {
		Id  int64  `json:"id"`
		Url string `json:"url"`
	} `json:"profileImage"`
	Description string    `json:"description"`
	Email       string    `json:"email"`
	WebSite     string    `json:"webSite"`
	CreateTime  time.Time `json:"createTime"`
	UpdateTime  time.Time `json:"updateTime"`
}

// TODO: Email verification, WebSite url syntax validation
func (h *userHandler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorw("failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"))
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	user, err := h.UserRepository.SelectUser(repository.ClassifiedById(userId))
	if err != nil {
		log.Errorw("failed to select user",
			"error code", util.ErrInternalServerError,
			"error", err,
			"userId", userId)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	profileImageId := int64(0)
	profileImageUrl := defaultUserProfileImage
	if user.ProfileImage.Valid {
		image, err := h.ImageRepository.SelectImage(userId, user.ProfileImage.Int64)
		if err != nil {
			log.Errorw("failed to select image",
				"error code", util.ErrInternalServerError,
				"error", err)
			util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
			return
		}

		profileImageId = image.Id
		profileImageUrl = image.Url
	}

	resp := GetUserHandlerResponseBody{
		Name: user.Name,
		ProfileImage: struct {
			Id  int64  `json:"id"`
			Url string `json:"url"`
		}{Id: profileImageId, Url: profileImageUrl},
		Description: user.Description.String,
		Email:       user.Email.String,
		WebSite:     user.WebSite.String,
		CreateTime:  user.CreateTime,
		UpdateTime:  user.UpdateTime,
	}
	util.WriteJson(w, http.StatusOK, resp)
}

type UpdateUserHandlerRequestBody struct {
	ProfileImage int64  `json:"profileImage"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Email        string `json:"email"`
	WebSite      string `json:"webSite"`
}

func (b UpdateUserHandlerRequestBody) Validate() error {
	if err := checkUserNameLength(b.Name); err != nil {
		return err
	}

	if err := checkUserDescriptionLength(b.Description); err != nil {
		return err
	}

	return nil
}

const (
	_maximumUserNameLength        = 45
	_maximumUserDescriptionLength = 2000
)

func checkUserNameLength(userName string) error {
	if utf8.RuneCountInString(userName) > _maximumUserNameLength {
		return errors.New("user name too long")
	}
	return nil
}

func checkUserDescriptionLength(userDescription string) error {
	if utf8.RuneCountInString(userDescription) > _maximumUserDescriptionLength {
		return errors.New("user description too long")
	}
	return nil
}

func (h *userHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorw("failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"))
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	reqBody := UpdateUserHandlerRequestBody{}
	if err := util.BindJson(r.Body, &reqBody); err != nil {
		log.Warnw("failed to bind request body to json",
			"error code", util.ErrInvalidRequestBody,
			"error", err)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidRequestBody)
		return
	}

	user, err := h.UserRepository.SelectUser(repository.ClassifiedById(userId))
	if err != nil {
		log.Errorw("failed to select user",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
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
		if _, err := h.ImageRepository.SelectImage(userId, reqBody.ProfileImage); err != nil {
			if err == sql.ErrNoRows {
				log.Warnw("invalid image id",
					"request image ID", reqBody.ProfileImage,
					"request user ID", userId)
				util.WriteError(w, http.StatusBadRequest, util.ErrInvalidImageId)
				return
			}

			log.Errorw("failed to select image",
				"error code", util.ErrInternalServerError,
				"error", err)
			util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
			return
		}
	}

	user.ProfileImage = sql.NullInt64{
		Int64: reqBody.ProfileImage,
		Valid: reqBody.ProfileImage != 0,
	}

	if err := h.UserRepository.Update(user); err != nil {
		log.Errorw("failed to update user",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type UpdateUserPasswordHandlerRequestBody struct {
	PW string `json:"pw"`
}

func (b UpdateUserPasswordHandlerRequestBody) Validate() error {
	if err := validatePassword(b.PW); err != nil {
		return errors.Wrap(err, "invalid password")
	}

	return nil
}

func (h *userHandler) UpdateUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorw("failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"))
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	reqBody := UpdateUserPasswordHandlerRequestBody{}
	if err := util.BindJson(r.Body, &reqBody); err != nil {
		log.Warnw("failed to bind request body to json",
			"error code", util.ErrInvalidRequestBody,
			"error", err)
		util.WriteError(w, http.StatusBadRequest, util.ErrInvalidRequestBody)
		return
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reqBody.PW), bcrypt.DefaultCost)
	if err != nil {
		log.Errorw("failed to generate hashed password",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}
	log.Debugw("password hashed finish",
		"hashed password", hashedPassword,
		"hashed password len", len(hashedPassword))

	user, err := h.UserRepository.SelectUser(repository.ClassifiedById(userId))
	if err != nil {
		log.Errorw("failed to select user",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	user.LoginPw = util.NullBytes{
		Bytes: hashedPassword,
		Valid: true,
	}

	if err := h.UserRepository.Update(user); err != nil {
		log.Errorw("failed to update user",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *userHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		log.Errorw("failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"))
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	user, err := h.UserRepository.SelectUser(repository.ClassifiedById(userId))
	if err != nil {
		log.Errorw("failed to select user",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	if err := h.UserRepository.Delete(user); err != nil {
		log.Errorw("failed to delete user",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	if err := h.SessionService.Logout(w, r); err != nil {
		log.Errorw("failed to logout",
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
