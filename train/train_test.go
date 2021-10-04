package train

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestQuery_Apply(t *testing.T) {
	expectString := `SELECT user_id, project_no FROM train t WHERE t.id = ? LIMIT ?, ?`

	var q Query
	q.AddSelect("user_id, project_no").
		AddFrom("train t").
		AddWhere("t.id = ?", 2).
		AddLimit(0, 0)

	err := q.Apply()
	if err != nil {
		t.Errorf(err.Error())
	}

	fmt.Println(q.QueryString)

	if expectString != q.QueryString {
		t.Errorf("Result is not same")
	}
}

func TestTrainDbRepository_FindAll(t *testing.T) {
	dbUrl := getDBInfo()

	db, err := sqlx.Open("mysql", dbUrl)
	if err != nil {
		t.Errorf(err.Error())
	}

	var q Query
	q.AddSelect(defaultSelectTrainHistoryColumns).
		AddFrom(`train t`).
		AddJoin(`train_config tc ON t.id = tc.train_id`).
		AddJoin(`project p ON t.project_id = p.id`).
		AddWhere(`p.user_id = ?`, 2).
		AddWhere(`p.project_no = ?`, 1).
		AddWhere(`t.status != ?`, "'DEL'").
		AddLimit(0, 100)

	err = q.Apply()
	if err != nil {
		t.Errorf(err.Error())
	}

	var historyList []History

	fmt.Println(q.QueryString)

	rows, err := db.Queryx(q.QueryString, q.Args...)
	if err != nil {
		t.Errorf(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var history History
		err = rows.StructScan(history)
		if err != nil {
			t.Errorf(err.Error())
		}
		historyList = append(historyList, history)
	}

	fmt.Println(historyList)
}
