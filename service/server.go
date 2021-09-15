package service

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"nns_back/cloud"
	"nns_back/dataset"
	"nns_back/ws"
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

	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/share", e.GenerateShareKeyHandler).Methods(_Get...)

	// web socket
	hub := ws.NewHub(e.DB)

	//router.HandleFunc("/ws", hub.WsHandler)
	authRouter.HandleFunc("/ws/{key}", hub.WsHandler)

	///////////////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////////////

	awsAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsSessionToken := os.Getenv("AWS_SESSION_TOKEN")
	//imageBucketName := os.Getenv("IMAGE_BUCKET_NAME")
	datasetBucketName := os.Getenv("DATASET_BUCKET_NAME")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKeyId, awsSecretAccessKey, awsSessionToken)),
		config.WithRegion("ap-northeast-2"),
	)
	if err != nil {
		logger.Fatal(err)
	}

	s3Client := s3.NewFromConfig(cfg)

	datasetHandler := &dataset.Handler{
		Repository: &dataset.MysqlRepository{
			DB: db,
		},
		Logger:      logger,
		AwsS3Client: &cloud.AwsS3Client{
			Client:     s3Client,
			BucketName: datasetBucketName,
		},
	}

	authRouter.HandleFunc("/api/datasets", datasetHandler.GetList).Methods(_Get...)
	authRouter.HandleFunc("/api/dataset/file", datasetHandler.UploadFile).Methods(_Post...)
	authRouter.HandleFunc("/api/dataset", datasetHandler.UpdateFileConfig).Methods(_Put...)

	///////////////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////////////

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
	//e.Logger.Fatal(srv.ListenAndServeTLS("server.crt", "server.key"))
	e.Logger.Fatal(srv.ListenAndServe())
}
