package query

import (
	"fmt"
	"strings"
	"testing"
)

func TestQuery_Apply(t *testing.T) {
	expectString := `SELECT user_id, project_no FROM train t WHERE t.id = ? ORDER BY user_id DESC LIMIT ?, ?`

	var q Builder
	q.AddSelect("user_id, project_no").
		AddFrom("train t").
		AddWhere("t.id = ?", 2).
		AddLimit(0, 0).
		AddOrder("user_id DESC")

	err := q.Build()
	if err != nil {
		t.Errorf(err.Error())
	}

	fmt.Printf(q.QueryString)

	if expectString != q.QueryString {
		t.Errorf("Result is not same")
	}
}

func TestBuilder_AddInsert(t *testing.T) {
	expectString := "INSERT INTO " +
			"epoch(train_id, epoch, acc, loss, val_acc, val_loss, learning_rate) " +
			"VALUES(:train_id, :epoch, :acc, :loss, :val_acc, :val_loss, :learning_rate)"

	var q Builder
	q.AddInsert(
		"epoch",
		"train_id, epoch, acc, loss, val_acc, val_loss, learning_rate",
		":train_id, :epoch, :acc, :loss, :val_acc, :val_loss, :learning_rate",
		)

	err := q.Build()
	if err != nil {
		t.Errorf(err.Error())
	}

	fmt.Printf(q.QueryString)

	if strings.TrimSuffix(expectString, " ") != strings.TrimSuffix(q.QueryString, " ") {
		t.Errorf("Result is not same")
	}
}