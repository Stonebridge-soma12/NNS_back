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
	Logger       *zap.SugaredLogger
	DB           *sqlx.DB
	SessionStore sessions.Store
}

func (e Env) Start(port string) {
	e.Logger.Info("Start server")

	router := mux.NewRouter()

	router.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	}).Methods(http.MethodGet, http.MethodOptions)

	router.HandleFunc("/api/signup", e.SignUpHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/login", e.LoginHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/logout", e.LogoutHandler).Methods(http.MethodDelete)
	router.HandleFunc("/api/secret", e.Secret).Methods(http.MethodGet)

	// project
	router.HandleFunc("/api/projects", e.GetProjectListHandler).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/api/project/{projectNo:[0-9]+}", e.GetProjectHandler).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/api/project/{projectNo:[0-9]+}/content", e.GetProjectContentHandler).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/api/project/{projectNo:[0-9]+}/config", e.GetProjectConfigHandler).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/api/project/{projectNo:[0-9]+}/code", e.GetPythonCodeHandler).Methods(http.MethodGet, http.MethodOptions)

	router.HandleFunc("/api/project", e.CreateProjectHandler).Methods(http.MethodPost, http.MethodOptions)

	router.HandleFunc("/api/project/{projectNo:[0-9]+}/info", e.UpdateProjectInfoHandler).Methods(http.MethodPut, http.MethodOptions)
	router.HandleFunc("/api/project/{projectNo:[0-9]+}/content", e.UpdateProjectContentHandler).Methods(http.MethodPut, http.MethodOptions)
	router.HandleFunc("/api/project/{projectNo:[0-9]+}/config", e.UpdateProjectConfigHandler).Methods(http.MethodPut, http.MethodOptions)

	router.HandleFunc("/api/project/{projectNo:[0-9]+}", e.DeleteProjectHandler).Methods(http.MethodDelete, http.MethodOptions)

	router.Use(handlers.CORS(
		handlers.AllowedMethods([]string{http.MethodOptions, http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedHeaders([]string{"Accept", "Accept-Language", "Content-Type", "Content-Language", "Origin"}),
	))

	srv := &http.Server{
		Handler: handlers.CombinedLoggingHandler(os.Stderr, router),
		Addr:    port,
	}
	e.Logger.Fatal(srv.ListenAndServe())
}
