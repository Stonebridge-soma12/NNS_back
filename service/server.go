package service

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"os"
)

type Env struct {
	Logger *zap.SugaredLogger
	DB     *sqlx.DB
}

var (
	_Get    = []string{http.MethodGet, http.MethodOptions}
	_Post   = []string{http.MethodPost, http.MethodOptions}
	_Put    = []string{http.MethodPut, http.MethodOptions}
	_Delete = []string{http.MethodDelete, http.MethodOptions}
)

func Start(port string, logger *zap.SugaredLogger, db *sqlx.DB, sessionStore sessions.Store) {
	e := Env{
		Logger: logger,
		DB:     db,
	}
	e.Logger.Info("Start server")

	// default router
	router := mux.NewRouter()

	auth := Auth{
		Env:          e,
		SessionStore: sessionStore,
	}
	authRouter := router.PathPrefix("").Subrouter()
	authRouter.Use(auth.middleware)

	// auth
	router.HandleFunc("/api/login", auth.LoginHandler).Methods(_Post...)
	authRouter.HandleFunc("/api/logout", auth.LogoutHandler).Methods(_Delete...)

	// image
	authRouter.HandleFunc("/api/image", e.UploadImageHandler).Methods(_Post...)

	// user
	router.HandleFunc("/api/user", e.SignUpHandler).Methods(_Post...)
	authRouter.HandleFunc("/api/user", e.GetUserHandler).Methods(_Get...)
	authRouter.HandleFunc("/api/user", e.UpdateUserHandler).Methods(_Put...)
	authRouter.HandleFunc("/api/user/password", e.UpdateUserPasswordHandler).Methods(_Put...)
	authRouter.HandleFunc("/api/user", auth.DeleteUserHandler).Methods(_Delete...)

	// project
	authRouter.HandleFunc("/api/projects", e.GetProjectListHandler).Methods(_Get...)
	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}", e.GetProjectHandler).Methods(_Get...)
	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/content", e.GetProjectContentHandler).Methods(_Get...)
	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/config", e.GetProjectConfigHandler).Methods(_Get...)
	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/code", e.GetPythonCodeHandler).Methods(_Get...)

	authRouter.HandleFunc("/api/project", e.CreateProjectHandler).Methods(_Post...)

	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/info", e.UpdateProjectInfoHandler).Methods(_Put...)
	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/content", e.UpdateProjectContentHandler).Methods(_Put...)
	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/config", e.UpdateProjectConfigHandler).Methods(_Put...)

	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}", e.DeleteProjectHandler).Methods(_Delete...)

	router.Use(handlers.CORS(
		handlers.AllowedMethods([]string{http.MethodOptions, http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}),
		handlers.AllowedHeaders([]string{"Accept", "Accept-Language", "Content-Type", "Content-Language", "Origin"}),
		handlers.AllowCredentials(),

		// This option is used to bypass a well known security issue
		// when configured with AllowedOrigins to * and AllowCredentials to true.
		//
		// Must change to the option below in production.
		// handlers.AllowedOrigins([]string{"specific origin"})
		handlers.AllowedOriginValidator(func(s string) bool {
			return true
		}),
	))

	srv := &http.Server{
		Handler: handlers.CombinedLoggingHandler(os.Stderr, router),
		Addr:    port,
	}
	e.Logger.Fatal(srv.ListenAndServeTLS("server.crt", "server.key"))
}
