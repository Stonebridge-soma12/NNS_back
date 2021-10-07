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
	"net/http"
	"nns_back/cloud"
	"nns_back/dataset"
	"nns_back/externalAPI"
	"nns_back/log"
	"nns_back/repository"
	"nns_back/train"
	"nns_back/ws"
	"os"
	"time"
)

type Env struct {
	DB            *sqlx.DB
	CodeConverter externalAPI.CodeConverter
}

var (
	_Get    = []string{http.MethodGet, http.MethodOptions}
	_Post   = []string{http.MethodPost, http.MethodOptions}
	_Put    = []string{http.MethodPut, http.MethodOptions}
	_Delete = []string{http.MethodDelete, http.MethodOptions}
)

func Start(port string, db *sqlx.DB, sessionStore sessions.Store) {
	defaultTransportPointer, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		panic("failed to interface conversion")
	}
	defaultTransport := *defaultTransportPointer
	defaultTransport.MaxIdleConns = 100
	defaultTransport.MaxIdleConnsPerHost = 100

	httpClient := &http.Client{
		Transport: &defaultTransport,
		Timeout:   time.Second * 10,
	}

	e := Env{
		DB:            db,
		CodeConverter: externalAPI.NewCodeConverter(httpClient),
	}
	log.Info("Start server")

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
	imageHandler := ImageHandler{
		ImageRepository: repository.NewImageMysqlRepository(db),
	}
	authRouter.HandleFunc("/api/image", imageHandler.UploadImage).Methods(_Post...)

	// user
	router.HandleFunc("/api/user", e.SignUpHandler).Methods(_Post...)
	authRouter.HandleFunc("/api/user", e.GetUserHandler).Methods(_Get...)
	authRouter.HandleFunc("/api/user", e.UpdateUserHandler).Methods(_Put...)
	authRouter.HandleFunc("/api/user/password", e.UpdateUserPasswordHandler).Methods(_Put...)
	authRouter.HandleFunc("/api/user", auth.DeleteUserHandler).Methods(_Delete...)

	// project
	projectHandler := ProjectHandler{
		ProjectRepository: repository.NewProjectMysqlRepository(db),
		CodeConverter:     externalAPI.NewCodeConverter(httpClient),
	}
	authRouter.HandleFunc("/api/projects", projectHandler.GetProjectListHandler).Methods(_Get...)
	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}", projectHandler.GetProjectHandler).Methods(_Get...)
	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/content", projectHandler.GetProjectContentHandler).Methods(_Get...)
	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/config", projectHandler.GetProjectConfigHandler).Methods(_Get...)
	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/code", projectHandler.GetPythonCodeHandler).Methods(_Get...)

	authRouter.HandleFunc("/api/project", projectHandler.CreateProjectHandler).Methods(_Post...)

	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/info", projectHandler.UpdateProjectInfoHandler).Methods(_Put...)
	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/content", projectHandler.UpdateProjectContentHandler).Methods(_Put...)
	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/config", projectHandler.UpdateProjectConfigHandler).Methods(_Put...)

	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}", projectHandler.DeleteProjectHandler).Methods(_Delete...)

	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/share", projectHandler.GenerateShareKeyHandler).Methods(_Get...)


	// web socket
	hub := ws.NewHub(e.DB)

	//router.HandleFunc("/ws", hub.WsHandler)
	authRouter.HandleFunc("/ws/{key}", hub.WsHandler)

	// Train log monitor
	bridge := train.NewBridge(
		&train.EpochDbRepository{DB: db},
		&train.TrainDbRepository{DB: db},
		&train.TrainLogDbRepository{DB: db},
	)

	// Train monitor
	router.HandleFunc("/api/project/{projectNo:[0-9]+}/train/{trainNo:[0-9]+}/epoch", bridge.NewEpochHandler).Methods(_Post...)
	router.HandleFunc("/api/project/{projectNo:[0-9]+}/train/{trainNo:[0-9]+}/reply", bridge.TrainReplyHandler).Methods(_Post...)
	authRouter.HandleFunc("/ws/project/{projectNo:[0-9]+}/train/{trainNo:[0-9]+}", bridge.MonitorWsHandler)

	///////////////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////////////

	awsAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsSessionToken := os.Getenv("AWS_SESSION_TOKEN")
	//imageBucketName := os.Getenv("IMAGE_BUCKET_NAME")
	datasetBucketName := os.Getenv("DATASET_BUCKET_NAME")
	trainedModelBucketName := os.Getenv("TRAINED_MODEL_BUCKET_NAME")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKeyId, awsSecretAccessKey, awsSessionToken)),
		config.WithRegion("ap-northeast-2"),
	)
	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.NewFromConfig(cfg)

	datasetHandler := &dataset.Handler{
		Repository: &dataset.MysqlRepository{
			DB: db,
		},
		AwsS3Client: &cloud.AwsS3Client{
			Client:     s3Client,
			BucketName: datasetBucketName,
		},
	}

	authRouter.HandleFunc("/api/datasets", datasetHandler.GetList).Methods(_Get...)
	authRouter.HandleFunc("/api/dataset/file", datasetHandler.UploadFile).Methods(_Post...)
	authRouter.HandleFunc("/api/dataset", datasetHandler.UpdateFileConfig).Methods(_Put...)
	authRouter.HandleFunc("/api/dataset/{datasetId:[0-9]+}", datasetHandler.DeleteDataset).Methods(_Delete...)

	authRouter.HandleFunc("/api/dataset/library", datasetHandler.GetLibraryList).Methods(_Get...)
	authRouter.HandleFunc("/api/dataset/library", datasetHandler.AddNewDatasetToLibrary).Methods(_Post...)
	authRouter.HandleFunc("/api/dataset/library/{datasetId:[0-9]+}", datasetHandler.DeleteDatasetFromLibrary).Methods(_Delete...)

	// Train Handler
	trainHandler := train.Handler{
		Fitter:            externalAPI.NewFitter(httpClient),
		ProjectRepository: repository.NewProjectMysqlRepository(db),
		TrainRepository: &train.TrainDbRepository{
			DB: db,
		},
		AwsS3Uploader: &cloud.AwsS3Client{
			Client:     s3Client,
			BucketName: trainedModelBucketName,
		},
	}

	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/train", trainHandler.GetTrainHistoryListHandler).Methods(_Get...)
	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/train/{trainNo:[0-9]+}", trainHandler.DeleteTrainHistoryHandler).Methods(_Delete...)
	authRouter.HandleFunc("/api/project/{projectNo:[0-9]+}/train/{trainNo:[0-9]+}", trainHandler.UpdateTrainHistoryHandler).Methods(_Put...)

	authRouter.HandleFunc("/api/model", trainHandler.SaveTrainModelHandler).Methods(_Post...)

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
	//log.Fatal(srv.ListenAndServeTLS("server.crt", "server.key"))
	log.Fatal(srv.ListenAndServe())
}
