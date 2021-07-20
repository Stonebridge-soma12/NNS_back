package service

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"os"
)

type Env struct {
	Logger *zap.SugaredLogger
	DB     *sqlx.DB
}

func (e Env) Start(port string) {
	e.Logger.Info("Start server")

	router := mux.NewRouter()

	router.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	}).Methods(http.MethodGet)

	// project
	router.HandleFunc("/api/projects", e.GetProjectListHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/project/{projectNo:[0-9]+}", e.GetProjectHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/project/{projectNo:[0-9]+}/content", e.GetProjectContentHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/project/{projectNo:[0-9]+}/config", e.GetProjectConfigHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/project", e.CreateProjectHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/project/{projectNo:[0-9]+}/info", e.UpdateProjectInfoHandler).Methods(http.MethodPut)
	router.HandleFunc("/api/project/{projectNo:[0-9]+}/content", e.UpdateProjectContentHandler).Methods(http.MethodPut)
	router.HandleFunc("/api/project/{projectNo:[0-9]+}/config", e.UpdateProjectConfigHandler).Methods(http.MethodPut)
	router.HandleFunc("/api/project/{projectNo:[0-9]+}", e.DeleteProjectHandler).Methods(http.MethodDelete)
	
	router.Use(handlers.CORS(
		handlers.AllowedMethods([]string{"*"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedHeaders([]string{"*"}),
	))

	srv := &http.Server{
		Handler: handlers.CombinedLoggingHandler(os.Stderr, router),
		Addr:    port,
	}
	e.Logger.Fatal(srv.ListenAndServe())
}
