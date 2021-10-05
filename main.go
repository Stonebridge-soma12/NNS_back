package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"nns_back/log"
	"nns_back/service"
	"os"
	"time"
)

func main() {
	// set logger
	log.Init(zap.DebugLevel)

	// set db connection
	dbUser := os.Getenv("DBUSER")
	dbPW := os.Getenv("DBPW")
	dbIP := os.Getenv("DBIP")
	dbPort := os.Getenv("DBPORT")
	db, err := sqlx.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%s)/nns?parseTime=true", dbUser, dbPW, dbIP, dbPort))
	if err != nil {
		log.Fatal("failed to db open", err)
	}
	defer db.Close()

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// start server
	service.Start(":8080", db, service.SetSessionStore([]byte(os.Getenv("SESSKEY"))))
}
