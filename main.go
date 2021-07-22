package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"nns_back/service"
	"os"
	"time"
)

func main() {
	// set logger
	config := zap.NewProductionConfig()
	config.DisableStacktrace = true
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	config.Level.SetLevel(zap.DebugLevel)
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// set db connection
	dbUser := os.Getenv("DBUSER")
	dbPW := os.Getenv("DBPW")
	dbIP := os.Getenv("DBIP")
	dbPort := os.Getenv("DBPORT")
	db, err := sqlx.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%s)/nns?parseTime=true", dbUser, dbPW, dbIP, dbPort))
	if err != nil {
		sugar.Fatal("failed to db open", err)
	}
	defer db.Close()

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// start server
	service.Start(":8080", sugar, db, service.SetSessionStore([]byte(os.Getenv("SESSKEY"))))
}
