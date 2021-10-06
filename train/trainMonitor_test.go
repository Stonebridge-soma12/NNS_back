package train

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"testing"
)

func getDBInfo() string {
	id := os.Getenv("DBUSER")
	pw := os.Getenv("DBPW")
	ip := os.Getenv("DBIP")
	port := os.Getenv("DBPORT")

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/nns?parseTime=true", id, pw, ip, port)
}

func TestPostEpochHandler(t *testing.T) {

}