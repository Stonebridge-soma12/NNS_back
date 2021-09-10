package trainMonitor

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

func getDBInfo() string {
	id := os.Getenv("id")
	pw := os.Getenv("pw")
	url := os.Getenv("url")

	return fmt.Sprintf("%s:%s@tcp(%s)/nns?parseTime=true&charset=utf8mb4", id, pw, url)
}
